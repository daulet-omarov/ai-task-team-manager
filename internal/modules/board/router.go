package board

import (
	"github.com/daulet-omarov/ai-task-team-manager/internal/middleware"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, h *Handler) {
	// WS route authenticates via ?token= query param — must be outside JWTMiddleware.
	r.Get("/boards/{id}/events", h.ServeWS)

	r.Group(func(r chi.Router) {
		r.Use(middleware.JWTMiddleware)
		r.Get("/dashboard", h.GetDashboard)
		r.Post("/boards", h.Create)
		r.Get("/boards/{id}", h.GetByID)
		r.Patch("/boards/{id}", h.UpdateBoard)
		r.Delete("/boards/{id}", h.Delete)
		r.Get("/boards/{id}/members", h.GetMembers)
		r.Get("/boards/{id}/member-stats", h.GetMemberStats)
		r.Delete("/board-members/{boardMemberId}", h.DeleteMember)
		r.Get("/boards/{id}/statuses", h.GetStatuses)
		r.Post("/statuses", h.CreateStatus)
		r.Patch("/statuses/reorder", h.ReorderStatuses)
		r.Patch("/statuses/{boardStatusId}/set-default", h.SetDefaultStatus)
		r.Patch("/statuses/{boardStatusId}", h.UpdateStatus)
		r.Delete("/statuses/{boardStatusId}", h.DeleteStatus)
	})
}
