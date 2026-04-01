package main

import (
	"fmt"

	"github.com/daulet-omarov/ai-task-team-manager/internal/config"
	"github.com/daulet-omarov/ai-task-team-manager/internal/database"
	"github.com/daulet-omarov/ai-task-team-manager/internal/seeder"
)

func main() {
	cfg := config.Load()

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)

	// singleton DB
	db, _ := database.NewPostgres(dsn)

	s := seeder.New(db)
	s.Run()
}
