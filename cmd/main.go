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
	tracingadapter "github.com/kuromii5/chat-bot-auth-service/internal/adapters/tracing"
	apperrors "github.com/kuromii5/chat-bot-auth-service/internal/errors"
	httpserver "github.com/kuromii5/chat-bot-auth-service/internal/handlers/http"
	authhandler "github.com/kuromii5/chat-bot-auth-service/internal/handlers/http/auth"
	authmw "github.com/kuromii5/chat-bot-auth-service/internal/handlers/http/middleware"
	userhandler "github.com/kuromii5/chat-bot-auth-service/internal/handlers/http/user"
	"github.com/kuromii5/chat-bot-auth-service/internal/service/session"
	tracingsvc "github.com/kuromii5/chat-bot-auth-service/internal/service/tracing"
	userservice "github.com/kuromii5/chat-bot-auth-service/internal/service/user"
	"github.com/kuromii5/chat-bot-shared/jwt"
	"github.com/kuromii5/chat-bot-shared/tracing"
	"github.com/kuromii5/chat-bot-shared/validator"
	"github.com/kuromii5/chat-bot-shared/wrapper"
)

func main() {
	cfg := config.MustLoad()
	setupLogger(cfg.Log.Level)
	validator.Init()
	wrapper.RegisterErrors(apperrors.Registry)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	shutdownTracer, err := tracing.InitTracer(
		context.Background(),
		"auth-service",
		cfg.Tracing.Endpoint,
		cfg.Tracing.Sampler,
	)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to init OpenTelemetry")
	}
	defer func() {
		if err := shutdownTracer(context.Background()); err != nil {
			logrus.WithError(err).Error("Failed to shutdown tracer")
		}
	}()

	pg, err := postgres.New(&cfg.Database)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to connect to database")
	}

	jwtManager := jwt.NewJWTManager(
		cfg.JWT.Secret,
		cfg.JWT.AccessTokenExpiry,
		cfg.JWT.RefreshTokenExpiry,
	)

	tracingPG := tracingadapter.NewRepo(pg)

	userSvc := userservice.NewService(tracingPG)
	sessionSvc := session.NewService(tracingPG, tracingPG, jwtManager)

	jail := authmw.NewIPJail(cfg.RateLimit.MaxFailures, cfg.RateLimit.JailMinutes)

	router := httpserver.NewRouter(
		userhandler.NewHandler(tracingsvc.NewUserService(userSvc)),
		authhandler.NewHandler(tracingsvc.NewAuthService(sessionSvc)),
		cfg.JWT.Secret,
		jail,
	)

	httpserver.InitMetrics(ctx, cfg.Metrics.Port)
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
	logrus.AddHook(&tracing.OTelHook{})
}
