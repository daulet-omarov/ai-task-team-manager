package comment

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

func (r *Repository) Create(c *models.Comment) error {
	return r.db.Create(c).Error
}

func (r *Repository) GetByID(id uint) (*models.Comment, error) {
	var c models.Comment
	err := r.db.First(&c, id).Error
	return &c, err
}

func (r *Repository) GetByTaskID(taskID uint) ([]*models.Comment, error) {
	var comments []*models.Comment
	err := r.db.Where("task_id = ?", taskID).Order("created_at").Find(&comments).Error
	return comments, err
}

func (r *Repository) Delete(id uint) error {
	return r.db.Delete(&models.Comment{}, id).Error
}
