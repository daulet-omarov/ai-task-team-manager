package task

import (
	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/board"
	"gorm.io/gorm"
)

func NewModule(db *gorm.DB) *Handler {
	repo := NewRepository(db)
	boardRepo := board.NewRepository(db) // satisfies boardMemberChecker interface
	service := NewService(repo, boardRepo)
	return NewHandler(service)
}
