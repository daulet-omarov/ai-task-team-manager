package chat

import (
	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/board"
	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/employee"
	"gorm.io/gorm"
)

func NewModule(db *gorm.DB) *Handler {
	repo := NewRepository(db)
	boardRepo := board.NewRepository(db)
	employeeRepo := employee.NewRepository(db)
	service := NewService(repo, boardRepo, employeeRepo)
	return NewHandler(service)
}
