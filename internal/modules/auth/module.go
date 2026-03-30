package auth

import (
	"database/sql"
	"github.com/daulet-omarov/ai-task-team-manager/pkg/mailer"
)

func NewModule(db *sql.DB, m *mailer.Mailer, baseURL string) *Handler {
	repo := NewRepository(db)
	service := NewService(repo, m, baseURL)
	handler := NewHandler(service)

	return handler
}
