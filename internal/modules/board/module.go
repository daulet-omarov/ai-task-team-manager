package board

import (
	"github.com/daulet-omarov/ai-task-team-manager/internal/hub"
	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/employee"
	"gorm.io/gorm"
)

func NewModule(db *gorm.DB, h *hub.Hub) *Handler {
	repo := NewRepository(db)
	empRepo := employee.NewRepository(db)
	service := NewService(repo, empRepo)
	return NewHandler(service, h)
}
