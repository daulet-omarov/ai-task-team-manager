package invite

import "gorm.io/gorm"

func NewModule(db *gorm.DB) *Handler {
	repo := NewRepository(db)
	service := NewService(repo)
	return NewHandler(service)
}
