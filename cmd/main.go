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
	httpserver "github.com/kuromii5/chat-bot-auth-service/internal/handlers/http"
	authhandler "github.com/kuromii5/chat-bot-auth-service/internal/handlers/http/auth"
	userhandler "github.com/kuromii5/chat-bot-auth-service/internal/handlers/http/user"
	"github.com/kuromii5/chat-bot-auth-service/internal/service/session"
	userservice "github.com/kuromii5/chat-bot-auth-service/internal/service/user"
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

	jwtManager := jwt.NewJWTManager(
		cfg.JWT.Secret,
		cfg.JWT.AccessTokenExpiry,
		cfg.JWT.RefreshTokenExpiry,
	)

	userSvc := userservice.NewService(pg)
	sessionSvc := session.NewService(pg, pg, jwtManager)

	router := httpserver.NewRouter(
		userhandler.NewHandler(userSvc),
		authhandler.NewHandler(sessionSvc),
	)

	httpserver.InitMetrics(cfg.Metrics.Port)
	server := httpserver.NewServer(cfg.Server.Host, cfg.Server.Port, router)

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
		if closeErr := pg.DB.Close(); closeErr != nil {
			logrus.WithError(closeErr).Error("Database close failed")
		}
		return
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
		DisableTimestamp: true,
	})
}
