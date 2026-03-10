package auth

import "github.com/go-chi/chi/v5"

func RegisterRoutes(r chi.Router, handler *Handler) {

	r.Route("/auth", func(r chi.Router) {

		r.Post("/register", handler.Register)
		r.Post("/login", handler.Login)
		r.Post("/forgot-password", handler.ForgotPassword)
		r.Delete("/account", handler.DeleteAccount)

	})
}
