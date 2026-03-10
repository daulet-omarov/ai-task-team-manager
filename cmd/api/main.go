package main

import (
	httpSwagger "github.com/swaggo/http-swagger"
	"log"
	"net/http"

	"github.com/daulet-omarov/ai-task-team-manager/internal/database"
	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/auth"
	"github.com/go-chi/chi/v5"

	_ "github.com/daulet-omarov/ai-task-team-manager/docs"
)

// @title AI Task Team Manager API
// @version 1.0
// @description AI Task Team Manager API is a backend service designed to manage team tasks using artificial intelligence. It provides endpoints for task creation, team collaboration, role management, and AI-powered task recommendations to improve productivity and workflow management.
// @host localhost:8080
// @BasePath /
func main() {
	db := database.NewPostgres()

	repo := auth.NewRepository(db)
	service := auth.NewService(repo)
	handler := auth.NewHandler(service)

	r := chi.NewRouter()

	auth.RegisterRoutes(r, handler)

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	log.Println("starting server on port 8080")

	err := http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatal(err)
	}
}
