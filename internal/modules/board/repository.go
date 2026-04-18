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

func (r *Repository) CountMembers(boardID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.BoardMember{}).
		Where("board_id = ?", boardID).
		Count(&count).Error
	return count, err
}
