package employee

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

func (r *Repository) Create(e *models.Employee) error {
	return r.db.Create(e).Error
}

func (r *Repository) GetByID(id uint) (*models.Employee, error) {
	var e models.Employee
	err := r.db.
		Preload("Team").
		Preload("Gender").
		First(&e, id).Error
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *Repository) GetByUserID(userID uint) (*models.Employee, error) {
	var e models.Employee
	err := r.db.
		Preload("Team").
		Preload("Gender").
		Where("user_id = ?", userID).
		First(&e).Error
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *Repository) GetAll() ([]*models.Employee, error) {
	var employees []*models.Employee
	err := r.db.
		Preload("Team").
		Preload("Gender").
		Find(&employees).Error
	return employees, err
}

func (r *Repository) Update(e *models.Employee) error {
	return r.db.Save(e).Error
}

func (r *Repository) Delete(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&models.Employee{}).Error
}
