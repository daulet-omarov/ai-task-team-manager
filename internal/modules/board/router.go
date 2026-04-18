package board

import (
	"github.com/daulet-omarov/ai-task-team-manager/internal/middleware"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, h *Handler) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.JWTMiddleware)
		r.Get("/dashboard", h.GetDashboard)
		r.Post("/boards", h.Create)
		r.Get("/boards/{id}", h.GetByID)
	})
}
