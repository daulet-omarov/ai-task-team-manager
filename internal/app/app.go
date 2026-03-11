package app

import (
	"fmt"
	"net/http"

	_ "github.com/daulet-omarov/ai-task-team-manager/docs"
	"github.com/daulet-omarov/ai-task-team-manager/internal/config"
	"github.com/daulet-omarov/ai-task-team-manager/internal/database"
	"github.com/daulet-omarov/ai-task-team-manager/internal/logger"
	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/auth"
	"github.com/daulet-omarov/ai-task-team-manager/internal/router"
	"github.com/daulet-omarov/ai-task-team-manager/internal/validator"
	"github.com/daulet-omarov/ai-task-team-manager/pkg/jwt"
	httpSwagger "github.com/swaggo/http-swagger"
)

type App struct {
	Server *http.Server
}

func New() *App {

	// init logger
	logger.Init()

	validator.Init()

	cfg := config.Load()

	jwt.Init(cfg.JWTSecret)

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

	// modules
	authRepo := auth.NewRepository(db)
	authService := auth.NewService(authRepo)
	authHandler := auth.NewHandler(authService)

	// router
	r := router.SetupRouter(authHandler)

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	return &App{
		Server: server,
	}
}

func (a *App) Run() error {
	logger.Log.Info("starting server")

	return a.Server.ListenAndServe()
}
