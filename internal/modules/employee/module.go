package employee

import "gorm.io/gorm"

func NewModule(db *gorm.DB) *Handler {
	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	return handler
}
