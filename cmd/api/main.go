package main

import (
	"log"
	"net/http"

	"github.com/daulet-omarov/ai-task-team-manager/internal/database"
	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/auth"
	"github.com/go-chi/chi/v5"
)

func main() {
	db := database.NewPostgres()

	repo := auth.NewRepository(db)
	service := auth.NewService(repo)
	handler := auth.NewHandler(service)

	r := chi.NewRouter()

	auth.RegisterRoutes(r, handler)

	log.Println("starting server on port 8080")

	err := http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatal(err)
	}
}
