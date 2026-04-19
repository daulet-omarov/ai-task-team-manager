package router

import (
	"net/http"

	"github.com/daulet-omarov/ai-task-team-manager/internal/middleware"
	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/attachment"
	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/auth"
	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/board"
	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/comment"
	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/employee"
	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/invite"
	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/task"
	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/upload"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func SetupRouter(
	authHandler *auth.Handler,
	employeeHandler *employee.Handler,
	boardHandler *board.Handler,
	taskHandler *task.Handler,
	inviteHandler *invite.Handler,
	uploadHandler *upload.Handler,
	commentHandler *comment.Handler,
	attachmentHandler *attachment.Handler,
) *chi.Mux {

	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{
			"http://192.168.100.23:5173",
		},
		AllowedMethods: []string{
			"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS",
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
	board.RegisterRoutes(r, boardHandler)
	task.RegisterRoutes(r, taskHandler)
	invite.RegisterRoutes(r, inviteHandler)
	upload.RegisterRoutes(r, uploadHandler)
	comment.RegisterRoutes(r, commentHandler)
	attachment.RegisterRoutes(r, attachmentHandler)

	// Serve uploaded files as static assets: GET /uploads/<filename>
	fileServer := http.FileServer(http.Dir("./uploads"))
	r.Handle("/uploads/*", http.StripPrefix("/uploads", fileServer))

	return r
}
