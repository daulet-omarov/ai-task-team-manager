package task

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

func (r *Repository) GetStatusIDByCode(code string) (uint, error) {
	var status models.Status
	err := r.db.Where("code = ?", code).First(&status).Error
	return status.ID, err
}

func (r *Repository) Create(t *models.Task) error {
	return r.db.Create(t).Error
}

func (r *Repository) GetByID(id uint) (*models.Task, error) {
	var t models.Task
	err := r.db.First(&t, id).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *Repository) Update(t *models.Task) error {
	return r.db.Save(t).Error
}

func (r *Repository) GetByBoardID(boardID uint) ([]*models.Task, error) {
	var tasks []*models.Task
	err := r.db.Where("board_id = ?", boardID).Order("created_at").Find(&tasks).Error
	return tasks, err
}
