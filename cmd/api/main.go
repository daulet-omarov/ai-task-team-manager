package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/daulet-omarov/ai-task-team-manager/internal/database"
	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/auth"

	_ "github.com/daulet-omarov/ai-task-team-manager/docs"
)

// @title AI Task Team Manager API
// @version 1.0
// @description AI Task Team Manager API is a backend service designed to manage team tasks using artificial intelligence.
// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter the token with the `Bearer ` prefix, e.g. "Bearer abcde12345".
func main() {

	// database
	db := database.NewPostgres()

	// auth module
	repo := auth.NewRepository(db)
	service := auth.NewService(repo)
	handler := auth.NewHandler(service)

	// router
	r := chi.NewRouter()

	// routes
	auth.RegisterRoutes(r, handler)

	// swagger
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	log.Println("starting server on port 8080")

	err := http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatal(err)
	}
}
