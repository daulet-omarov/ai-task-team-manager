package auth

import (
	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/employee"
	"gorm.io/gorm"

	"github.com/daulet-omarov/ai-task-team-manager/pkg/mailer"
)

func NewModule(db *gorm.DB, m *mailer.Mailer, baseURL string) *Handler {
	repo := NewRepository(db)
	employeeRepo := employee.NewRepository(db)
	service := NewService(repo, employeeRepo, m, baseURL)
	handler := NewHandler(service)

	return handler
}
