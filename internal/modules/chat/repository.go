package chat

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

func (r *Repository) CreateMessage(m *models.BoardChatMessage) error {
	return r.db.Create(m).Error
}

func (r *Repository) GetMessages(boardID uint, limit, offset int) ([]models.BoardChatMessage, error) {
	var msgs []models.BoardChatMessage
	err := r.db.
		Where("board_id = ?", boardID).
		Preload("Author").
		Preload("Attachments").
		Preload("Poll.Options.Votes").
		Preload("ReplyTo.Author").
		Preload("ReplyTo.Attachments").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&msgs).Error
	return msgs, err
}

func (r *Repository) GetMessageByID(id uint) (*models.BoardChatMessage, error) {
	var m models.BoardChatMessage
	err := r.db.
		Preload("Author").
		Preload("Attachments").
		Preload("Poll.Options.Votes").
		First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *Repository) DeleteMessage(id uint) error {
	return r.db.Delete(&models.BoardChatMessage{}, id).Error
}

func (r *Repository) CreateAttachment(a *models.BoardChatAttachment) error {
	return r.db.Create(a).Error
}

func (r *Repository) CreatePoll(p *models.BoardPoll) error {
	return r.db.Create(p).Error
}

func (r *Repository) GetPollByID(id uint) (*models.BoardPoll, error) {
	var p models.BoardPoll
	err := r.db.
		Preload("Options.Votes").
		First(&p, id).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) AddVote(v *models.BoardPollVote) error {
	return r.db.Create(v).Error
}

func (r *Repository) RemoveVote(optionID, employeeID uint) error {
	return r.db.
		Where("option_id = ? AND employee_id = ?", optionID, employeeID).
		Delete(&models.BoardPollVote{}).Error
}

func (r *Repository) HasVoted(optionID, employeeID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.BoardPollVote{}).
		Where("option_id = ? AND employee_id = ?", optionID, employeeID).
		Count(&count).Error
	return count > 0, err
}

func (r *Repository) GetOptionByID(id uint) (*models.BoardPollOption, error) {
	var o models.BoardPollOption
	err := r.db.First(&o, id).Error
	if err != nil {
		return nil, err
	}
	return &o, nil
}
