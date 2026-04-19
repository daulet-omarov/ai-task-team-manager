package board

import (
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
func (r *Repository) AddDefaultStatuses(boardID uint) error {
	return r.db.Exec(`
		INSERT INTO board_statuses (board_id, status_id, position)
		SELECT ?, s.id,
		       CASE s.code
		           WHEN 'to_do'       THEN 1
		           WHEN 'in_progress' THEN 2
		           WHEN 'done'        THEN 3
		       END
		FROM statuses s
		WHERE s.code IN ('to_do', 'in_progress', 'done')
		ON CONFLICT (status_id, board_id) DO NOTHING
	`, boardID).Error
}

// GetBoardStatuses returns the board's statuses ordered by position.
func (r *Repository) GetBoardStatuses(boardID uint) ([]*StatusResponse, error) {
	var rows []*StatusResponse
	err := r.db.Raw(`
		SELECT bs.id AS board_status_id, s.id AS status_id, s.name, s.code, bs.position, bs.colour
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
	err := r.db.Exec(`
		INSERT INTO statuses (name, code) VALUES (?, ?)
		ON CONFLICT (code) DO NOTHING
	`, name, code).Error
	if err != nil {
		return nil, err
	}

	var statusID uint
	if err := r.db.Raw("SELECT id FROM statuses WHERE code = ?", code).Scan(&statusID).Error; err != nil {
		return nil, err
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
		SELECT bs.id AS board_status_id, s.id AS status_id, s.name, s.code, bs.position, bs.colour
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
		SELECT bm.user_id,
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
