package main

import (
	"github.com/daulet-omarov/ai-task-team-manager/internal/app"
)

// @title AI Task Team Manager API
// @version 1.0
// @description AI Task Team Manager API is a backend service designed to manage team tasks using artificial intelligence.
// @host localhost:7777
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter the token with the `Bearer ` prefix, e.g. "Bearer abcde12345".
func main() {

	application := app.New()

	if err := application.Run(); err != nil {
		panic(err)
	}
}
