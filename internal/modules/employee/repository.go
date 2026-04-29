package employee

import (
	"time"

	"github.com/daulet-omarov/ai-task-team-manager/internal/models"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(e *models.Employee) error {
	return r.db.Create(e).Error
}

func (r *Repository) GetByID(id uint) (*models.Employee, error) {
	var e models.Employee
	err := r.db.
		Preload("Gender").
		First(&e, id).Error
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *Repository) GetByEmail(email string) (*models.Employee, error) {
	var e models.Employee
	err := r.db.Where("email = ?", email).First(&e).Error
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *Repository) GetByUserID(userID uint) (*models.Employee, error) {
	var e models.Employee
	err := r.db.
		Preload("Gender").
		Where("user_id = ?", userID).
		First(&e).Error
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *Repository) GetAll() ([]*models.Employee, error) {
	var employees []*models.Employee
	err := r.db.
		Preload("Gender").
		Find(&employees).Error
	return employees, err
}

func (r *Repository) Update(e *models.Employee) error {
	return r.db.Save(e).Error
}

func (r *Repository) Delete(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&models.Employee{}).Error
}

type dailyActivityRow struct {
	Date      string    `gorm:"column:date"`
	Count     int       `gorm:"column:count"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

// computeActivities runs the weighted SQL against tasks/comments and returns fresh rows.
//
// Scoring formula:
//   - Task created (reporter)                         → 1.0 pt
//   - Task assigned as developer, status = done       → 1.1 pt
//   - Task assigned as developer, active status       → 0.6 pt  (excludes to_do/done)
//   - Task assigned as tester                         → 0.4 pt
//   - Comment authored                                → 0.3 pt
func (r *Repository) computeActivities(employeeID uint) ([]dailyActivityRow, error) {
	var rows []dailyActivityRow
	err := r.db.Raw(`
		SELECT
			TO_CHAR(date_trunc('day', ts), 'YYYY-MM-DD') AS date,
			ROUND(SUM(pts))::int                          AS count
		FROM (
			SELECT created_at AS ts, 1.0 AS pts
			FROM tasks
			WHERE reporter_id = @id

			UNION ALL

			SELECT t.updated_at AS ts, 1.1 AS pts
			FROM tasks t
			JOIN statuses s ON s.id = t.status_id
			WHERE t.developer_id = @id AND s.code = 'done'

			UNION ALL

			SELECT t.updated_at AS ts, 0.6 AS pts
			FROM tasks t
			JOIN statuses s ON s.id = t.status_id
			WHERE t.developer_id = @id
			  AND s.code NOT IN ('to_do', 'done')

			UNION ALL

			SELECT updated_at AS ts, 0.4 AS pts
			FROM tasks
			WHERE tester_id = @id

			UNION ALL

			SELECT created_at AS ts, 0.3 AS pts
			FROM comments
			WHERE author_id = @id
		) AS contributions
		GROUP BY date
		ORDER BY date
	`, map[string]any{"id": employeeID}).Scan(&rows).Error
	return rows, err
}

// getStoredActivities reads pre-computed rows from employee_contributions.
// Returns the rows and the MAX updated_at (= last time any computation ran for this employee).
func (r *Repository) getStoredActivities(employeeID uint) ([]dailyActivityRow, time.Time, error) {
	var rows []dailyActivityRow
	err := r.db.Raw(`
		SELECT TO_CHAR(date, 'YYYY-MM-DD') AS date, count, updated_at
		FROM employee_contributions
		WHERE employee_id = ?
		ORDER BY date
	`, employeeID).Scan(&rows).Error
	if err != nil || len(rows) == 0 {
		return nil, time.Time{}, err
	}

	newest := rows[0].UpdatedAt
	for _, row := range rows[1:] {
		if row.UpdatedAt.After(newest) {
			newest = row.UpdatedAt
		}
	}
	return rows, newest, nil
}

// upsertActivities replaces ALL stored rows for the employee (used on first-time compute).
func (r *Repository) upsertActivities(employeeID uint, rows []dailyActivityRow) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(
			"DELETE FROM employee_contributions WHERE employee_id = ?", employeeID,
		).Error; err != nil {
			return err
		}
		for _, row := range rows {
			if err := tx.Exec(`
				INSERT INTO employee_contributions (employee_id, date, count, updated_at)
				VALUES (?, ?, ?, now())
			`, employeeID, row.Date, row.Count).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// computeActivitiesSince runs the same weighted formula but only for dates >= fromDate.
func (r *Repository) computeActivitiesSince(employeeID uint, fromDate string) ([]dailyActivityRow, error) {
	var rows []dailyActivityRow
	err := r.db.Raw(`
		SELECT
			TO_CHAR(date_trunc('day', ts), 'YYYY-MM-DD') AS date,
			ROUND(SUM(pts))::int                          AS count
		FROM (
			SELECT created_at AS ts, 1.0 AS pts
			FROM tasks
			WHERE reporter_id = @id AND created_at::date >= @from

			UNION ALL

			SELECT t.updated_at AS ts, 1.1 AS pts
			FROM tasks t
			JOIN statuses s ON s.id = t.status_id
			WHERE t.developer_id = @id AND s.code = 'done'
			  AND t.updated_at::date >= @from

			UNION ALL

			SELECT t.updated_at AS ts, 0.6 AS pts
			FROM tasks t
			JOIN statuses s ON s.id = t.status_id
			WHERE t.developer_id = @id
			  AND s.code NOT IN ('to_do', 'done')
			  AND t.updated_at::date >= @from

			UNION ALL

			SELECT updated_at AS ts, 0.4 AS pts
			FROM tasks
			WHERE tester_id = @id AND updated_at::date >= @from

			UNION ALL

			SELECT created_at AS ts, 0.3 AS pts
			FROM comments
			WHERE author_id = @id AND created_at::date >= @from
		) AS contributions
		GROUP BY date
		ORDER BY date
	`, map[string]any{"id": employeeID, "from": fromDate}).Scan(&rows).Error
	return rows, err
}

// findMinAffectedDate returns the earliest date whose stored contribution could be
// wrong due to changes that happened after `since`.
//
// For reporter/comment events the affected date = their created_at (immutable).
// For developer/tester events we don't know the OLD updated_at, so we use
// the task's created_at as a conservative lower bound — the task could have
// contributed to any day from creation onwards.
//
// Returns ("", false, nil) when nothing changed since `since`.
func (r *Repository) findMinAffectedDate(employeeID uint, since time.Time) (string, bool, error) {
	var result struct {
		MinDate *string `gorm:"column:min_date"`
	}
	err := r.db.Raw(`
		SELECT MIN(ts::date)::text AS min_date
		FROM (
			SELECT created_at AS ts FROM tasks
			WHERE reporter_id = @id AND created_at > @since

			UNION ALL

			SELECT created_at AS ts FROM comments
			WHERE author_id = @id AND created_at > @since

			UNION ALL

			SELECT created_at AS ts FROM tasks
			WHERE (developer_id = @id OR tester_id = @id) AND updated_at > @since
		) changes
	`, map[string]any{"id": employeeID, "since": since}).Scan(&result).Error
	if err != nil {
		return "", false, err
	}
	if result.MinDate == nil {
		return "", false, nil
	}
	return *result.MinDate, true, nil
}

type achievementCountRow struct {
	Closer        int64 `gorm:"column:closer"`
	Challenger    int64 `gorm:"column:challenger"`
	Elite         int64 `gorm:"column:elite"`
	Perfectionist int64 `gorm:"column:perfectionist"`
	Cleanworker   int64 `gorm:"column:cleanworker"`
	Teamplayer    int64 `gorm:"column:teamplayer"`
	Reviewer      int64 `gorm:"column:reviewer"`
	Communicator  int64 `gorm:"column:communicator"`
	Pollmaster    int64 `gorm:"column:pollmaster"`
	Influencer    int64 `gorm:"column:influencer"`
	Voicepioneer  int64 `gorm:"column:voicepioneer"`
	Broadcaster   int64 `gorm:"column:broadcaster"`
}

func (r *Repository) getAchievementCounts(employeeID uint) (*achievementCountRow, error) {
	var row achievementCountRow
	err := r.db.Raw(`
		SELECT
		  (SELECT COUNT(*) FROM tasks t
		   JOIN statuses s ON s.id = t.status_id
		   WHERE t.developer_id = @emp AND s.code = 'done') AS closer,

		  (SELECT COUNT(*) FROM tasks t
		   JOIN statuses s ON s.id = t.status_id
		   JOIN difficulties d ON d.id = t.difficulty_id
		   WHERE t.developer_id = @emp AND s.code = 'done' AND d.code = 'hard') AS challenger,

		  (SELECT COUNT(*) FROM tasks t
		   JOIN statuses s ON s.id = t.status_id
		   JOIN difficulties d ON d.id = t.difficulty_id
		   WHERE t.developer_id = @emp AND s.code = 'done' AND d.code = 'very_hard') AS elite,

		  (SELECT COUNT(*) FROM tasks t
		   JOIN statuses s ON s.id = t.status_id
		   WHERE t.developer_id = @emp AND s.code = 'done') AS perfectionist,

		  (SELECT COUNT(DISTINCT t.id) FROM tasks t
		   WHERE t.developer_id = @emp AND t.description != ''
		     AND EXISTS (SELECT 1 FROM attachments a WHERE a.task_id = t.id)) AS cleanworker,

		  (SELECT COUNT(DISTINCT developer_id) FROM tasks
		   WHERE reporter_id = @emp) AS teamplayer,

		  (SELECT COUNT(*) FROM tasks WHERE tester_id = @emp) AS reviewer,

		  (SELECT COUNT(*) FROM board_chat_messages WHERE author_id = @emp) AS communicator,

		  (SELECT COUNT(*) FROM board_polls p
		   JOIN board_chat_messages m ON m.id = p.message_id
		   WHERE m.author_id = @emp) AS pollmaster,

		  (SELECT COUNT(*) FROM board_chat_messages r
		   JOIN board_chat_messages orig ON orig.id = r.reply_to_id
		   WHERE orig.author_id = @emp) AS influencer,

		  (SELECT COUNT(*) FROM board_chat_attachments a
		   JOIN board_chat_messages m ON m.id = a.message_id
		   WHERE m.author_id = @emp AND a.mime_type LIKE 'audio/%') AS voicepioneer,

		  (SELECT COUNT(*) FROM board_chat_attachments a
		   JOIN board_chat_messages m ON m.id = a.message_id
		   WHERE m.author_id = @emp AND a.mime_type LIKE 'video/%') AS broadcaster
	`, map[string]any{"emp": employeeID}).Scan(&row).Error
	return &row, err
}

func (r *Repository) getActivityDates(employeeID uint) ([]string, error) {
	var dates []string
	err := r.db.Raw(
		"SELECT TO_CHAR(date, 'YYYY-MM-DD') FROM employee_contributions WHERE employee_id = ? ORDER BY date",
		employeeID,
	).Scan(&dates).Error
	return dates, err
}

// touchActivities refreshes updated_at for all stored rows of an employee
// so the freshness check doesn't re-fire when nothing actually changed.
func (r *Repository) touchActivities(employeeID uint) error {
	return r.db.Exec(
		"UPDATE employee_contributions SET updated_at = now() WHERE employee_id = ?",
		employeeID,
	).Error
}

// upsertActivitiesSince deletes rows >= fromDate, inserts recomputed rows,
// and touches updated_at on all older rows so the freshness check stays valid.
func (r *Repository) upsertActivitiesSince(employeeID uint, fromDate string, rows []dailyActivityRow) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(
			"DELETE FROM employee_contributions WHERE employee_id = ? AND date >= ?",
			employeeID, fromDate,
		).Error; err != nil {
			return err
		}
		for _, row := range rows {
			if err := tx.Exec(`
				INSERT INTO employee_contributions (employee_id, date, count, updated_at)
				VALUES (?, ?, ?, now())
			`, employeeID, row.Date, row.Count).Error; err != nil {
				return err
			}
		}
		// Refresh older rows so MAX(updated_at) = now(), preventing redundant recomputes.
		return tx.Exec(
			"UPDATE employee_contributions SET updated_at = now() WHERE employee_id = ? AND date < ?",
			employeeID, fromDate,
		).Error
	})
}
