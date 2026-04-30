package board

import (
	"fmt"

	"github.com/daulet-omarov/ai-task-team-manager/internal/models"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(b *models.Board) error {
	return r.db.Create(b).Error
}

func (r *Repository) GetByID(id uint) (*models.Board, error) {
	var b models.Board
	err := r.db.Preload("Members").First(&b, id).Error
	if err != nil {
		return nil, err
	}
	return &b, nil
}

// GetBoardsByUserID returns all boards where the user is a member (including as owner).
func (r *Repository) GetBoardsByUserID(userID int64) ([]*models.Board, error) {
	var boards []*models.Board
	err := r.db.
		Joins("JOIN board_members ON board_members.board_id = boards.id").
		Where("board_members.user_id = ?", userID).
		Preload("Members").
		Find(&boards).Error
	return boards, err
}

func (r *Repository) AddMember(boardID uint, userID int64, role string) error {
	member := &models.BoardMember{
		BoardID: boardID,
		UserID:  userID,
		Role:    role,
	}
	return r.db.Create(member).Error
}

// IsMember reports whether userID belongs to boardID.
func (r *Repository) IsMember(boardID uint, userID int64) (bool, error) {
	var count int64
	err := r.db.Model(&models.BoardMember{}).
		Where("board_id = ? AND user_id = ?", boardID, userID).
		Count(&count).Error
	return count > 0, err
}

// AddDefaultStatuses links TO DO (pos 1), IN PROGRESS (pos 2), DONE (pos 3) to the board.
// The first status (to_do, position=1) is marked as is_default=true and is_reopen=true.
// The last status (done, position=3) is marked as is_completed=true.
func (r *Repository) AddDefaultStatuses(boardID uint) error {
	return r.db.Exec(`
		INSERT INTO board_statuses (board_id, status_id, position, is_default, is_completed, is_reopen)
		SELECT ?, s.id,
		       CASE s.code
		           WHEN 'to_do'       THEN 1
		           WHEN 'in_progress' THEN 2
		           WHEN 'done'        THEN 3
		       END,
		       CASE s.code WHEN 'to_do' THEN true ELSE false END,
		       CASE s.code WHEN 'done'  THEN true ELSE false END,
		       CASE s.code WHEN 'to_do' THEN true ELSE false END
		FROM statuses s
		WHERE s.code IN ('to_do', 'in_progress', 'done')
		ON CONFLICT (status_id, board_id) DO NOTHING
	`, boardID).Error
}

// GetBoardStatuses returns the board's statuses ordered by position.
func (r *Repository) GetBoardStatuses(boardID uint) ([]*StatusResponse, error) {
	var rows []*StatusResponse
	err := r.db.Raw(`
		SELECT bs.id AS board_status_id, s.id AS status_id, s.name, s.code, bs.position, bs.colour,
		       bs.is_default, bs.is_completed, bs.is_reopen
		FROM board_statuses bs
		JOIN statuses s ON s.id = bs.status_id
		WHERE bs.board_id = ?
		ORDER BY bs.position
	`, boardID).Scan(&rows).Error
	return rows, err
}

// nextPosition returns max(position)+1 for the board, starting at 1.
func (r *Repository) nextPosition(boardID uint) (int, error) {
	var max int
	err := r.db.Raw(
		"SELECT COALESCE(MAX(position), 0) FROM board_statuses WHERE board_id = ?",
		boardID,
	).Scan(&max).Error
	return max + 1, err
}

// UpsertStatus inserts a status by code (generating from title) if it doesn't exist,
// then links it to the board. Returns the StatusResponse.
func (r *Repository) UpsertStatus(boardID uint, name, code, colour string) (*StatusResponse, error) {
	// SELECT first to avoid hitting the INSERT (and the ID sequence) when the status already exists.
	var statusID uint
	r.db.Raw("SELECT id FROM statuses WHERE code = ?", code).Scan(&statusID)

	if statusID == 0 {
		// Sync the sequence so seeded rows with explicit IDs don't block new inserts.
		r.db.Exec("SELECT setval(pg_get_serial_sequence('statuses', 'id'), COALESCE((SELECT MAX(id) FROM statuses), 0))")

		if err := r.db.Exec(`
			INSERT INTO statuses (name, code) VALUES (?, ?)
			ON CONFLICT (code) DO NOTHING
		`, name, code).Error; err != nil {
			return nil, err
		}
		r.db.Raw("SELECT id FROM statuses WHERE code = ?", code).Scan(&statusID)
	}

	if statusID == 0 {
		return nil, fmt.Errorf("failed to upsert status with code %q", code)
	}

	pos, err := r.nextPosition(boardID)
	if err != nil {
		return nil, err
	}

	err = r.db.Exec(`
		INSERT INTO board_statuses (board_id, status_id, position, colour) VALUES (?, ?, ?, ?)
		ON CONFLICT (status_id, board_id) DO NOTHING
	`, boardID, statusID, pos, colour).Error
	if err != nil {
		return nil, err
	}

	var boardStatusID uint
	if err := r.db.Raw(
		"SELECT id FROM board_statuses WHERE board_id = ? AND status_id = ?",
		boardID, statusID,
	).Scan(&boardStatusID).Error; err != nil {
		return nil, err
	}

	return &StatusResponse{
		BoardStatusID: boardStatusID,
		StatusID:      statusID,
		Name:          name,
		Code:          code,
		Position:      pos,
		Colour:        colour,
	}, nil
}

func (r *Repository) UpdateBoardStatus(boardStatusID uint, title, colour string) (*StatusResponse, error) {
	if title != "" {
		if err := r.db.Exec(
			"UPDATE statuses SET name = ? WHERE id = (SELECT status_id FROM board_statuses WHERE id = ?)",
			title, boardStatusID,
		).Error; err != nil {
			return nil, err
		}
	}
	if colour != "" {
		if err := r.db.Exec(
			"UPDATE board_statuses SET colour = ? WHERE id = ?",
			colour, boardStatusID,
		).Error; err != nil {
			return nil, err
		}
	}

	var row StatusResponse
	err := r.db.Raw(`
		SELECT bs.id AS board_status_id, s.id AS status_id, s.name, s.code, bs.position, bs.colour,
		       bs.is_default, bs.is_completed, bs.is_reopen
		FROM board_statuses bs
		JOIN statuses s ON s.id = bs.status_id
		WHERE bs.id = ?
	`, boardStatusID).Scan(&row).Error
	return &row, err
}

// ReorderStatuses updates positions using board_statuses.id (board_status_id).
func (r *Repository) ReorderStatuses(positions []StatusPosition) error {
	tx := r.db.Begin()
	for _, sp := range positions {
		if err := tx.Exec(
			"UPDATE board_statuses SET position = ? WHERE id = ?",
			sp.Position, sp.BoardStatusID,
		).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

// DeleteBoardStatus removes a status from a board by board_statuses.id.
// Returns the boardID so the caller can verify membership.
func (r *Repository) GetBoardIDByBoardStatusID(boardStatusID uint) (uint, error) {
	var boardID uint
	err := r.db.Raw("SELECT board_id FROM board_statuses WHERE id = ?", boardStatusID).Scan(&boardID).Error
	return boardID, err
}

func (r *Repository) DeleteBoardStatus(boardStatusID uint) error {
	return r.db.Exec("DELETE FROM board_statuses WHERE id = ?", boardStatusID).Error
}

func (r *Repository) CountMembers(boardID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.BoardMember{}).
		Where("board_id = ?", boardID).
		Count(&count).Error
	return count, err
}

// GetMembersWithDetails returns board members joined with their employee profile (LEFT JOIN).
// Users without an employee profile still appear with empty name/photo/email fields.
func (r *Repository) GetMembersWithDetails(boardID uint) ([]*MemberResponse, error) {
	var rows []*MemberResponse
	err := r.db.Raw(`
		SELECT bm.id AS board_member_id,
		       bm.user_id,
		       bm.role,
		       COALESCE(e.full_name, '')  AS full_name,
		       COALESCE(e.photo, '')      AS photo,
		       COALESCE(e.email, '')      AS email
		FROM board_members bm
		LEFT JOIN employees e ON e.user_id = bm.user_id
		WHERE bm.board_id = ?
		ORDER BY bm.joined_at
	`, boardID).Scan(&rows).Error
	return rows, err
}

func (r *Repository) UpdateBoard(boardID uint, name, description string) error {
	updates := map[string]any{}
	if name != "" {
		updates["name"] = name
	}
	if description != "" {
		updates["description"] = description
	}
	if len(updates) == 0 {
		return nil
	}
	return r.db.Model(&models.Board{}).Where("id = ?", boardID).Updates(updates).Error
}

// Delete removes a board by ID.
func (r *Repository) Delete(boardID uint) error {
	return r.db.Delete(&models.Board{}, boardID).Error
}

// GetBoardStatusFlags returns the is_completed and is_reopen flags for a given
// status within a board. Both flags are false when the row is not found.
func (r *Repository) GetBoardStatusFlags(boardID uint, statusID uint) (isCompleted, isReopen bool, err error) {
	var row struct {
		IsCompleted bool
		IsReopen    bool
	}
	err = r.db.Raw(`
		SELECT is_completed, is_reopen
		FROM board_statuses
		WHERE board_id = ? AND status_id = ?
		LIMIT 1
	`, boardID, statusID).Scan(&row).Error
	return row.IsCompleted, row.IsReopen, err
}

// IsOwner reports whether userID is the owner of boardID.
func (r *Repository) IsOwner(boardID uint, userID int64) (bool, error) {
	var count int64
	err := r.db.Model(&models.BoardMember{}).
		Where("board_id = ? AND user_id = ? AND role = 'owner'", boardID, userID).
		Count(&count).Error
	return count > 0, err
}

// GetMemberByID returns a board member row by its primary key.
func (r *Repository) GetMemberByID(boardMemberID uint) (*models.BoardMember, error) {
	var m models.BoardMember
	err := r.db.First(&m, boardMemberID).Error
	return &m, err
}

// DeleteMember removes a board_members row by its primary key.
func (r *Repository) DeleteMember(boardMemberID uint) error {
	return r.db.Delete(&models.BoardMember{}, boardMemberID).Error
}

// SetDefaultFirstStatus marks the status at position=1 as the default for the board,
// clearing any previous default. Used after bulk imports when no status has is_default=true yet.
func (r *Repository) SetDefaultFirstStatus(boardID uint) error {
	var firstID uint
	if err := r.db.Raw(
		"SELECT id FROM board_statuses WHERE board_id = ? ORDER BY position ASC LIMIT 1",
		boardID,
	).Scan(&firstID).Error; err != nil || firstID == 0 {
		return err
	}
	return r.SetDefaultBoardStatus(firstID)
}

// SetDefaultBoardStatus sets is_default=true for the given board_status_id and
// false for all other statuses of the same board. Runs in a transaction.
func (r *Repository) SetDefaultBoardStatus(boardStatusID uint) error {
	var boardID uint
	if err := r.db.Raw("SELECT board_id FROM board_statuses WHERE id = ?", boardStatusID).Scan(&boardID).Error; err != nil {
		return err
	}
	if boardID == 0 {
		return fmt.Errorf("board status not found")
	}

	tx := r.db.Begin()
	if err := tx.Exec("UPDATE board_statuses SET is_default = false WHERE board_id = ?", boardID).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Exec("UPDATE board_statuses SET is_default = true WHERE id = ?", boardStatusID).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

// SetCompletedBoardStatus marks the given board_status as the completed status for the board
// and clears the flag on all other statuses of the same board.
func (r *Repository) SetCompletedBoardStatus(boardStatusID uint) error {
	var boardID uint
	if err := r.db.Raw("SELECT board_id FROM board_statuses WHERE id = ?", boardStatusID).Scan(&boardID).Error; err != nil {
		return err
	}
	if boardID == 0 {
		return fmt.Errorf("board status not found")
	}

	tx := r.db.Begin()
	if err := tx.Exec("UPDATE board_statuses SET is_completed = false WHERE board_id = ?", boardID).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Exec("UPDATE board_statuses SET is_completed = true WHERE id = ?", boardStatusID).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

// SetReopenBoardStatus marks the given board_status as the reopen status for the board
// and clears the flag on all other statuses of the same board.
func (r *Repository) SetReopenBoardStatus(boardStatusID uint) error {
	var boardID uint
	if err := r.db.Raw("SELECT board_id FROM board_statuses WHERE id = ?", boardStatusID).Scan(&boardID).Error; err != nil {
		return err
	}
	if boardID == 0 {
		return fmt.Errorf("board status not found")
	}

	tx := r.db.Begin()
	if err := tx.Exec("UPDATE board_statuses SET is_reopen = false WHERE board_id = ?", boardID).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Exec("UPDATE board_statuses SET is_reopen = true WHERE id = ?", boardStatusID).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (r *Repository) GetMemberStats(boardID uint) ([]*MemberStatsResponse, error) {
	var rows []*MemberStatsResponse
	err := r.db.Raw(`
		SELECT
		  bm.user_id,
		  COUNT(DISTINCT t.id) FILTER (WHERE s.code = 'done' AND d.code = 'hard') AS hard_tasks,
		  COUNT(DISTINCT t.id) FILTER (WHERE s.code = 'done')                     AS completed_tasks,
		  COUNT(DISTINCT p.id)                                                     AS polls_created,
		  COUNT(DISTINCT m.id)                                                     AS messages_sent
		FROM board_members bm
		LEFT JOIN tasks t
		  ON t.developer_id = bm.user_id AND t.board_id = bm.board_id
		LEFT JOIN statuses s
		  ON s.id = t.status_id
		LEFT JOIN difficulties d
		  ON d.id = t.difficulty_id
		LEFT JOIN board_chat_messages m
		  ON m.author_id = bm.user_id AND m.board_id = bm.board_id
		LEFT JOIN board_polls p
		  ON p.message_id = m.id
		WHERE bm.board_id = ?
		GROUP BY bm.user_id
		ORDER BY bm.user_id
	`, boardID).Scan(&rows).Error
	return rows, err
}

// GetDefaultStatusIDForBoard returns the status_id of the default board status.
// Falls back to the status with position=1 if none is marked default.
func (r *Repository) GetDefaultStatusIDForBoard(boardID uint) (uint, error) {
	var statusID uint
	err := r.db.Raw(`
		SELECT s.id FROM board_statuses bs
		JOIN statuses s ON s.id = bs.status_id
		WHERE bs.board_id = ? AND bs.is_default = true
		LIMIT 1
	`, boardID).Scan(&statusID).Error
	if err != nil {
		return 0, err
	}
	if statusID != 0 {
		return statusID, nil
	}
	// Fallback: position=1
	err = r.db.Raw(`
		SELECT s.id FROM board_statuses bs
		JOIN statuses s ON s.id = bs.status_id
		WHERE bs.board_id = ?
		ORDER BY bs.position ASC
		LIMIT 1
	`, boardID).Scan(&statusID).Error
	return statusID, err
}
