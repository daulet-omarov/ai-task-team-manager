package invite

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

func (r *Repository) Create(inv *models.Invite) error {
	return r.db.Create(inv).Error
}

func (r *Repository) GetByID(id uint) (*models.Invite, error) {
	var inv models.Invite
	err := r.db.Preload("Board").First(&inv, id).Error
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

// GetPendingByInvitee returns all pending invitations for a user.
func (r *Repository) GetPendingByInvitee(userID int64) ([]*models.Invite, error) {
	var invites []*models.Invite
	err := r.db.
		Preload("Board").
		Where("invitee_id = ? AND status = ?", userID, models.InviteStatusPending).
		Order("created_at DESC").
		Find(&invites).Error
	return invites, err
}

func (r *Repository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&models.Invite{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// ExistsPending checks whether a pending invite already exists for this board+invitee pair.
func (r *Repository) ExistsPending(boardID uint, inviteeID int64) (bool, error) {
	var count int64
	err := r.db.Model(&models.Invite{}).
		Where("board_id = ? AND invitee_id = ? AND status = ?", boardID, inviteeID, models.InviteStatusPending).
		Count(&count).Error
	return count > 0, err
}

func (r *Repository) IsOwner(boardID uint, userID int64) (bool, error) {
	var count int64
	err := r.db.Model(&models.Board{}).
		Where("id = ? AND owner_id = ?", boardID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *Repository) IsMember(boardID uint, userID int64) (bool, error) {
	var count int64
	err := r.db.Model(&models.BoardMember{}).
		Where("board_id = ? AND user_id = ?", boardID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *Repository) AddMember(boardID uint, userID int64) error {
	member := &models.BoardMember{
		BoardID: boardID,
		UserID:  userID,
		Role:    "member",
	}
	return r.db.Create(member).Error
}
