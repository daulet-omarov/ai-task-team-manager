package employee

import (
	"github.com/daulet-omarov/ai-task-team-manager/internal/middleware"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, h *Handler) {
	r.Route("/employees", func(r chi.Router) {
		r.Use(middleware.JWTMiddleware)
		r.Post("/", h.Create)
		r.Get("/", h.GetAll)
		r.Get("/me", h.GetByUserID)
		r.Get("/me/activities", h.GetActivities)
		r.Put("/", h.Update)
		r.Delete("/", h.Delete)
		r.Get("/exists", h.Exists)
	})
}
