package invite

import (
	"github.com/daulet-omarov/ai-task-team-manager/internal/hub"
	"gorm.io/gorm"
)

func NewModule(db *gorm.DB, h *hub.Hub) *Handler {
	repo := NewRepository(db)
	service := NewService(repo)
	return NewHandler(service, h)
}
