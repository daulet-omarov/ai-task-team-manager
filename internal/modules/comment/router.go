package comment

import (
	"github.com/daulet-omarov/ai-task-team-manager/internal/middleware"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, h *Handler) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.JWTMiddleware)
		r.Post("/tasks/{taskId}/comments", h.Create)
		r.Get("/tasks/{taskId}/comments", h.GetByTaskID)
		r.Delete("/comments/{commentId}", h.Delete)
	})
}
