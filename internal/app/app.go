package app

import (
	"fmt"
	"net/http"

	"github.com/daulet-omarov/ai-task-team-manager/docs"
	"github.com/daulet-omarov/ai-task-team-manager/internal/config"
	"github.com/daulet-omarov/ai-task-team-manager/internal/database"
	"github.com/daulet-omarov/ai-task-team-manager/internal/hub"
	"github.com/daulet-omarov/ai-task-team-manager/internal/logger"
	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/attachment"
	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/auth"
	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/board"
	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/chat"
	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/comment"
	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/employee"
	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/invite"
	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/notion"
	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/task"
	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/upload"
	"github.com/daulet-omarov/ai-task-team-manager/internal/router"
	"github.com/daulet-omarov/ai-task-team-manager/internal/validator"
	"github.com/daulet-omarov/ai-task-team-manager/pkg/jwt"
	"github.com/daulet-omarov/ai-task-team-manager/pkg/mailer"
	"github.com/daulet-omarov/ai-task-team-manager/pkg/uploader"
	httpSwagger "github.com/swaggo/http-swagger"
)

type App struct {
	Server  *http.Server
	tlsCert string
	tlsKey  string
}

func New() *App {

	// init logger
	logger.Init()

	// init validator
	validator.Init()

	// load configs
	cfg := config.Load()

	docs.SwaggerInfo.Host = cfg.AppUrl

	uploader.Init(uploader.Config{
		Endpoint:  cfg.S3Endpoint,
		Region:    cfg.S3Region,
		AccessKey: cfg.S3AccessKey,
		SecretKey: cfg.S3SecretKey,
		Bucket:    cfg.S3Bucket,
		PublicURL: cfg.S3PublicURL,
	})

	// init JWT
	jwt.Init(cfg.JWTSecret)

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
		cfg.DBSSLMode,
	)

	// singleton DB
	db, _ := database.NewPostgres(dsn)

	// mailer
	m := mailer.New(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPassword, cfg.SMTPFrom)

	// shared hub for board + task real-time events
	boardHub := hub.New()

	// modules
	authHandler := auth.NewModule(db, m, cfg.AppBaseURL)
	employeeHandler := employee.NewModule(db)
	boardHandler := board.NewModule(db, boardHub)
	taskHandler := task.NewModule(db, boardHub)
	inviteHandler := invite.NewModule(db, boardHub)
	uploadHandler := upload.NewHandler()
	commentHandler := comment.NewModule(db)
	attachmentHandler := attachment.NewModule(db)
	notionHandler := notion.NewModule(db)
	chatHandler := chat.NewModule(db, boardHub)

	// router
	r := router.SetupRouter(authHandler, employeeHandler, boardHandler, taskHandler, inviteHandler, uploadHandler, commentHandler, attachmentHandler, notionHandler, chatHandler, cfg.AllowedOrigins)

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	server := &http.Server{
		Addr:    "0.0.0.0:" + cfg.AppPort,
		Handler: r,
	}

	return &App{
		Server:  server,
		tlsCert: cfg.TLSCert,
		tlsKey:  cfg.TLSKey,
	}
}

func (a *App) Run() error {
	if a.tlsCert != "" && a.tlsKey != "" {
		logger.Log.Info("starting server (TLS)")
		return a.Server.ListenAndServeTLS(a.tlsCert, a.tlsKey)
	}
	logger.Log.Info("starting server")
	return a.Server.ListenAndServe()
}
