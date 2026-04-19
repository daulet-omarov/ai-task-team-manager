package attachment

import (
	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/board"
	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/task"
	"gorm.io/gorm"
)

func NewModule(db *gorm.DB) *Handler {
	repo := NewRepository(db)
	taskRepo := task.NewRepository(db)
	boardRepo := board.NewRepository(db)
	service := NewService(repo, taskRepo, boardRepo)
	return NewHandler(service)
}
