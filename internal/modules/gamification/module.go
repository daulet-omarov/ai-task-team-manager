package gamification

import "gorm.io/gorm"

func NewModule(db *gorm.DB) *Handler {
	repo := NewRepository(db)
	service := NewService(repo, db)
	return NewHandler(service)
}

// NewService exported so the task module can wire it as a notification hook
// without importing the handler layer.
func NewServiceFromDB(db *gorm.DB) *Service {
	return NewService(NewRepository(db), db)
}
