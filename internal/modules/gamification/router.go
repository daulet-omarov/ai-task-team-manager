package gamification

import (
	"github.com/daulet-omarov/ai-task-team-manager/internal/middleware"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, h *Handler) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.JWTMiddleware)

		r.Get("/leaderboard", h.GetLeaderboard)
		r.Get("/gamification/me", h.GetMyStats)
		r.Get("/gamification/users/{userId}", h.GetUserStats)
		r.Get("/gamification/history", h.GetPointsHistory)
		r.Get("/gamification/kudos/status", h.GetKudosStatus)
		r.Post("/kudos", h.GiveKudos)
	})
}
