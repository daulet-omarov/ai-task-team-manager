package router

import (
	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/employee"
	_ "net/http"

	"github.com/daulet-omarov/ai-task-team-manager/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/auth"
)

func SetupRouter(
	authHandler *auth.Handler,
	employeeHandler *employee.Handler,
) *chi.Mux {

	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{
			"http://192.168.100.23:5173",
		},
		AllowedMethods: []string{
			"GET", "POST", "PUT", "DELETE", "OPTIONS",
		},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-CSRF-Token",
		},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Use(middleware.LoggerMiddleware)

	auth.RegisterRoutes(r, authHandler)
	employee.RegisterRoutes(r, employeeHandler)

	return r
}
