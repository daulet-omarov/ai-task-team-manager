package chat

import (
	"github.com/daulet-omarov/ai-task-team-manager/internal/middleware"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, h *Handler) {
	// WS route authenticates via ?token= query param — must be outside JWTMiddleware.
	r.Get("/boards/{boardId}/ws", h.ServeWS)

	r.Group(func(r chi.Router) {
		r.Use(middleware.JWTMiddleware)

		r.Get("/boards/{boardId}/chat", h.GetMessages)
		r.Post("/boards/{boardId}/chat", h.SendMessage)
		r.Delete("/boards/{boardId}/chat/{msgId}", h.DeleteMessage)
		r.Post("/boards/{boardId}/chat/polls", h.CreatePoll)
		r.Post("/boards/{boardId}/chat/polls/vote", h.Vote)
		r.Post("/boards/{boardId}/chat/polls/unvote", h.Unvote)
	})
}
