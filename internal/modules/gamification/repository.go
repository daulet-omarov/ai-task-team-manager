package gamification

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

// GetOrCreateUserGamification returns the row, creating it with defaults if absent.
func (r *Repository) GetOrCreateUserGamification(userID int64) (*models.UserGamification, error) {
	ug := &models.UserGamification{UserID: userID, CurrentLevel: 1}
	result := r.db.Where("user_id = ?", userID).FirstOrCreate(ug)
	return ug, result.Error
}

// SaveUserGamification upserts the row inside the provided transaction.
func (r *Repository) SaveUserGamification(tx *gorm.DB, ug *models.UserGamification) error {
	return tx.Save(ug).Error
}

// InsertTransaction inserts a PointTransaction inside the provided transaction.
func (r *Repository) InsertTransaction(tx *gorm.DB, pt *models.PointTransaction) error {
	return tx.Create(pt).Error
}

// SumPointsToday returns the sum of positive points earned today (UTC) for userID.
func (r *Repository) SumPointsToday(userID int64) (int, error) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)
	var total int
	err := r.db.Raw(`
		SELECT COALESCE(SUM(points), 0)
		FROM point_transactions
		WHERE user_id = ? AND points > 0
		  AND earned_at >= ? AND earned_at < ?
	`, userID, today, tomorrow).Scan(&total).Error
	return total, err
}

// CountCompletionsToday counts task_completed transactions recorded today for userID.
// Used to detect the combo threshold (3+ completions).
func (r *Repository) CountCompletionsToday(tx *gorm.DB, userID int64) (int64, error) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)
	var count int64
	err := tx.Model(&models.PointTransaction{}).
		Where("user_id = ? AND reason = ? AND earned_at >= ? AND earned_at < ?",
			userID, ReasonTaskCompleted, today, tomorrow).
		Count(&count).Error
	return count, err
}

// GetTransactionsByTaskAndReason finds all transactions for a task with matching reasons.
func (r *Repository) GetTransactionsByTaskAndReason(taskID uint, userID int64, reasons []string) ([]models.PointTransaction, error) {
	var pts []models.PointTransaction
	err := r.db.Where("task_id = ? AND user_id = ? AND reason IN ?", taskID, userID, reasons).Find(&pts).Error
	return pts, err
}

// SumReworkPenalties returns the total absolute penalty already applied to a task for a user.
func (r *Repository) SumReworkPenalties(taskID uint, userID int64) (int, error) {
	var total int
	err := r.db.Raw(`
		SELECT COALESCE(SUM(ABS(points)), 0)
		FROM point_transactions
		WHERE task_id = ? AND user_id = ? AND reason = ?
	`, taskID, userID, ReasonReworkPenalty).Scan(&total).Error
	return total, err
}

// SumAllPointsForUser returns the lifetime total (sum of all transactions) for userID
// within the provided transaction, so newly inserted rows are included.
func (r *Repository) SumAllPointsForUser(tx *gorm.DB, userID int64) (int, error) {
	var total int
	err := tx.Raw(`
		SELECT COALESCE(SUM(points), 0) FROM point_transactions WHERE user_id = ?
	`, userID).Scan(&total).Error
	return total, err
}

// CountKudosGivenThisWeek counts how many kudos fromUserID gave since the start
// of the current ISO week (Monday 00:00 UTC).
func (r *Repository) CountKudosGivenThisWeek(fromUserID int64) (int64, error) {
	now := time.Now().UTC()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	weekStart := now.AddDate(0, 0, -(weekday - 1))
	weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, time.UTC)

	var count int64
	err := r.db.Model(&models.Kudos{}).
		Where("from_user_id = ? AND created_at >= ?", fromUserID, weekStart).
		Count(&count).Error
	return count, err
}

// InsertKudos inserts a kudos row.
func (r *Repository) InsertKudos(k *models.Kudos) error {
	return r.db.Create(k).Error
}

// GetLeaderboard returns top `limit` users by rolling 30-day points.
func (r *Repository) GetLeaderboard(limit int) ([]leaderboardRow, error) {
	cutoff := time.Now().UTC().Add(-30 * 24 * time.Hour)
	var rows []leaderboardRow
	err := r.db.Raw(`
		SELECT
			pt.user_id,
			COALESCE(e.full_name, '')  AS full_name,
			COALESCE(e.photo, '')      AS photo,
			SUM(pt.points)             AS rolling_points,
			COALESCE(ug.total_points, 0)   AS total_points,
			COALESCE(ug.current_level, 1)  AS current_level,
			COALESCE(ug.current_streak, 0) AS current_streak
		FROM point_transactions pt
		LEFT JOIN employees e  ON e.id = pt.user_id
		LEFT JOIN user_gamification ug ON ug.user_id = pt.user_id
		WHERE pt.earned_at >= ?
		GROUP BY pt.user_id, e.full_name, e.photo,
		         ug.total_points, ug.current_level, ug.current_streak
		ORDER BY rolling_points DESC
		LIMIT ?
	`, cutoff, limit).Scan(&rows).Error
	return rows, err
}

// GetRolling30dPoints returns a single user's rolling 30-day points.
func (r *Repository) GetRolling30dPoints(userID int64) (int, error) {
	cutoff := time.Now().UTC().Add(-30 * 24 * time.Hour)
	var total int
	err := r.db.Raw(`
		SELECT COALESCE(SUM(points), 0)
		FROM point_transactions
		WHERE user_id = ? AND earned_at >= ?
	`, userID, cutoff).Scan(&total).Error
	return total, err
}

// GetPointsHistory returns the most recent `limit` transactions for userID.
func (r *Repository) GetPointsHistory(userID int64, limit int) ([]models.PointTransaction, error) {
	var pts []models.PointTransaction
	err := r.db.Where("user_id = ?", userID).
		Order("earned_at DESC").
		Limit(limit).
		Find(&pts).Error
	return pts, err
}

// GetDifficultyCode returns the code string for a difficulty ID.
func (r *Repository) GetDifficultyCode(difficultyID uint) (string, error) {
	var diff models.Difficulty
	err := r.db.First(&diff, difficultyID).Error
	return diff.Code, err
}
