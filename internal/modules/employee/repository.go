package employee

import "gorm.io/gorm"

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(e *Employee) error {
	return r.db.Create(e).Error
}

func (r *Repository) GetByID(id uint) (*Employee, error) {
	var e Employee
	err := r.db.
		Preload("Role").
		Preload("Gender").
		Preload("User").
		First(&e, id).Error
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *Repository) GetAll() ([]*Employee, error) {
	var employees []*Employee
	err := r.db.
		Preload("Role").
		Preload("Gender").
		Preload("User").
		Find(&employees).Error
	return employees, err
}

func (r *Repository) Update(e *Employee) error {
	return r.db.Save(e).Error
}

func (r *Repository) Delete(id uint) error {
	return r.db.Delete(&Employee{}, id).Error
}
