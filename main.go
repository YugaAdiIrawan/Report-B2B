package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/YugaAdiIrawan/client"
	"github.com/YugaAdiIrawan/config"
	globalRepo "github.com/YugaAdiIrawan/globals"
	"github.com/YugaAdiIrawan/middleware"
	"github.com/YugaAdiIrawan/model/report"
	"github.com/YugaAdiIrawan/module/report/handler"
	"github.com/YugaAdiIrawan/module/report/repository"
	"github.com/YugaAdiIrawan/module/report/usecase"
	"github.com/YugaAdiIrawan/module/role/middlewareUsecase"
	"github.com/YugaAdiIrawan/module/role/miiddlewareRepo"
	"github.com/YugaAdiIrawan/module/schaduler"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

const (
	serviceVersion     = "1.0.0"
	defaultGracePeriod = 10 * time.Second
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Error().Msg("Error loading .env file")
	}
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal().Msg("env PORT is required")
	}

	db, err := config.NewMySQLDB(config.MySqlConfig{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASS"),
		DBName:   os.Getenv("DB_NAME"),
		MaxIdle:  5,
		MaxOpen:  25,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer db.Close()

	cfg, errCfg := config.LoadConfig()
	if errCfg != nil {
		log.Error().Msg("Failed load config client = " + errCfg.Error())
		return
	}

	app, err := buildApp(db, cfg)
	if err != nil {
		log.Error().Err(err).Msg("failed to build application")
		db.Close()
		os.Exit(1)
	}

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           app.router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	app.scheduler.Start()
	defer app.scheduler.Stop()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Info().
			Str("version", serviceVersion).
			Str("port", port).
			Time("started_at", time.Now()).
			Msg("service running")

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	<-quit
	log.Info().Msg("shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), defaultGracePeriod)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("server forced to shutdown")
	}

	log.Info().Msg("server exited gracefully")
}

type application struct {
	router    *gin.Engine
	scheduler *schaduler.AutoUWScheduler
}

func buildApp(db *sql.DB, cfg *config.Config) (*application, error) {
	authCfg, err := middleware.LoadServiceAuthConfig()
	if err != nil {
		log.Fatal().Msg(err.Error())
		return nil, nil
	}

	//global repository
	globalRepository := globalRepo.NewGlobalRepository(db)
	if errConfigSendComo := config.LoadComoSendEmailFromDB(cfg, globalRepository); errConfigSendComo != nil {
		log.Fatal().Err(errConfigSendComo).Msg("failed to load Como Send Email config from h2h_configs (id=5), app cannot start without valid Como credentials")
	}

	//Middleware
	mwRepo := miiddlewareRepo.NewMiddlewareRepository(db)
	mwUsecase := middlewareUsecase.NewMiddlewareUsecase(mwRepo)

	//Router
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	//Report
	reportRepo := repository.NewReportRepo(db)
	emailSvc := client.NewComoEmailService(cfg.ComoEmail, reportRepo)
	reportUC := usecase.NewReportUsecase(reportRepo, emailSvc)

	handler.NewReportHandler(router, reportUC, authCfg, mwUsecase)

	//Scheduler
	scheduler, err := schaduler.NewAutoUWSchaduers(reportUC, report.AutoUWSchedulerConfig{
		Recipients: cfg.AutoUWReport.Recipients,
		CCList:     cfg.AutoUWReport.CCList,
		Timezone:   "Asia/Jakarta",
		RunHour:    3,
		RunMinute:  0,
		//IntervalMin: 3,
	})
	if err != nil {
		return nil, fmt.Errorf("build app: init scheduler: %w", err)
	}
	schaduler.NewAdminSchedulerHandler(router, scheduler, authCfg)

	return &application{
		router:    router,
		scheduler: scheduler,
	}, nil
}
