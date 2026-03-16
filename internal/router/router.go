package router

import (
	"github.com/daulet-omarov/ai-task-team-manager/internal/middleware"
	"github.com/go-chi/chi/v5"

	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/auth"
)

func SetupRouter(authHandler *auth.Handler) *chi.Mux {

	r := chi.NewRouter()

	r.Use(middleware.LoggerMiddleware)

	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
		r.Get("/verify-email", authHandler.VerifyEmail)
		r.Post("/forgot-password", authHandler.ForgotPassword)
		r.With(middleware.JWTMiddleware).Delete("/account", authHandler.DeleteAccount)
	})

	return r
}
