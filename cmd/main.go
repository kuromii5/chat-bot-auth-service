package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/kuromii5/chat-bot-auth-service/config"
	"github.com/kuromii5/chat-bot-auth-service/internal/adapters/postgres"
	httpHandlers "github.com/kuromii5/chat-bot-auth-service/internal/handlers/http"
	"github.com/kuromii5/chat-bot-auth-service/internal/service"
	"github.com/kuromii5/chat-bot-auth-service/pkg/jwt"
	"github.com/kuromii5/chat-bot-auth-service/pkg/validator"
)

func main() {
	cfg := config.MustLoad()
	setupLogger(cfg.Log.Level)
	validator.Init()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	pg, err := postgres.New(&cfg.Database)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to connect to database")
	}

	jwtManager := jwt.NewJWTManager(cfg.JWT.Secret, cfg.JWT.AccessTokenExpiry, cfg.JWT.RefreshTokenExpiry)
	authService := service.NewService(pg, pg, jwtManager)
	authHandler := httpHandlers.NewHandler(authService)

	router := httpHandlers.NewRouter(authHandler)
	httpHandlers.InitMetrics(cfg.Metrics.Port)
	server := httpHandlers.NewServer(cfg.Server.Host, cfg.Server.Port, router)

	errChan := make(chan error, 1)
	go func() {
		logrus.Infof("server address: %s", server.Addr())
		if err := server.Start(); err != nil {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		logrus.WithError(err).Fatal("Failed to start server")
		pg.DB.Close()
		os.Exit(1)
	case <-ctx.Done():
		logrus.Info("Server shutdown...")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logrus.WithError(err).Error("HTTP server shutdown failed, forcing close")
		}

		if err := pg.DB.Close(); err != nil {
			logrus.WithError(err).Error("Database close failed")
		}
	}

	logrus.Info("Service shutdown successfully")
}

func setupLogger(level string) {
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		lvl = logrus.InfoLevel
	}
	logrus.SetLevel(lvl)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		DisableColors:   false,
	})
}
