package attachment

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

func (r *Repository) Create(a *models.Attachment) error {
	return r.db.Create(a).Error
}

func (r *Repository) GetByID(id uint) (*models.Attachment, error) {
	var a models.Attachment
	err := r.db.First(&a, id).Error
	return &a, err
}

func (r *Repository) GetByTaskID(taskID uint) ([]*models.Attachment, error) {
	var attachments []*models.Attachment
	err := r.db.Where("task_id = ?", taskID).Order("created_at").Find(&attachments).Error
	return attachments, err
}

func (r *Repository) Delete(id uint) error {
	return r.db.Delete(&models.Attachment{}, id).Error
}
