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
