package invite

import (
	"github.com/daulet-omarov/ai-task-team-manager/internal/middleware"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, h *Handler) {
	// WS authenticates via ?token= query param — must be outside JWTMiddleware.
	r.Get("/users/ws", h.ConnectUserWS)

	r.Group(func(r chi.Router) {
		r.Use(middleware.JWTMiddleware)
		r.Post("/boards/{boardId}/invite", h.Invite)
		r.Get("/invites", h.GetInvites)
		r.Post("/invites/{inviteId}/accept", h.Accept)
		r.Post("/invites/{inviteId}/reject", h.Reject)
	})
}
