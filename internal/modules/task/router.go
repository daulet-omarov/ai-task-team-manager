package task

import (
	"github.com/daulet-omarov/ai-task-team-manager/internal/middleware"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, h *Handler) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.JWTMiddleware)
		r.Post("/boards/{boardId}/tasks", h.Create)
		r.Get("/boards/{boardId}/tasks", h.GetByBoardID)
		r.Get("/tasks/{taskId}", h.GetByID)
		r.Patch("/tasks/{taskId}", h.Update)
		r.Delete("/tasks/{taskId}", h.Delete)
	})
}
