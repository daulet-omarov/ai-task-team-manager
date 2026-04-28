package notion

import (
	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/attachment"
	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/board"
	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/employee"
	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/task"
	"gorm.io/gorm"
)

func NewModule(db *gorm.DB) *Handler {
	boardRepo := board.NewRepository(db)
	taskRepo := task.NewRepository(db)
	attachmentRepo := attachment.NewRepository(db)
	employeeRepo := employee.NewRepository(db)
	service := NewService(boardRepo, taskRepo, attachmentRepo, employeeRepo)
	return NewHandler(service)
}
