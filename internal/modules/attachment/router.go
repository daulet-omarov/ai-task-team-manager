package attachment

import (
	"github.com/daulet-omarov/ai-task-team-manager/internal/middleware"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, h *Handler) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.JWTMiddleware)
		r.Post("/tasks/{taskId}/attachments", h.Upload)
		r.Get("/tasks/{taskId}/attachments", h.GetByTaskID)
		r.Delete("/attachments/{attachmentId}", h.Delete)
	})
}
