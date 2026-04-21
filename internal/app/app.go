package app

import (
	"fmt"
	"net/http"

	_ "github.com/daulet-omarov/ai-task-team-manager/docs"
	"github.com/daulet-omarov/ai-task-team-manager/internal/config"
	"github.com/daulet-omarov/ai-task-team-manager/internal/database"
	"github.com/daulet-omarov/ai-task-team-manager/internal/logger"
	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/attachment"
	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/auth"
	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/board"
	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/comment"
	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/employee"
	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/invite"
	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/task"
	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/upload"
	"github.com/daulet-omarov/ai-task-team-manager/internal/router"
	"github.com/daulet-omarov/ai-task-team-manager/internal/validator"
	"github.com/daulet-omarov/ai-task-team-manager/pkg/jwt"
	"github.com/daulet-omarov/ai-task-team-manager/pkg/mailer"
	httpSwagger "github.com/swaggo/http-swagger"
)

type App struct {
	Server *http.Server
}

func New() *App {

	// init logger
	logger.Init()

	// init validator
	validator.Init()

	// load configs
	cfg := config.Load()

	// init JWT
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

	// mailer
	m := mailer.New(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPassword, cfg.SMTPFrom)

	// modules
	authHandler := auth.NewModule(db, m, cfg.AppBaseURL)
	employeeHandler := employee.NewModule(db)
	boardHandler := board.NewModule(db)
	taskHandler := task.NewModule(db)
	inviteHandler := invite.NewModule(db)
	uploadHandler := upload.NewHandler()
	commentHandler := comment.NewModule(db)
	attachmentHandler := attachment.NewModule(db)

	// router
	r := router.SetupRouter(authHandler, employeeHandler, boardHandler, taskHandler, inviteHandler, uploadHandler, commentHandler, attachmentHandler)

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	server := &http.Server{
		Addr:    "0.0.0.0:" + cfg.AppPort,
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
