package app

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/kuromii5/chat-bot-auth-service/config"
	"github.com/kuromii5/chat-bot-auth-service/internal/adapters/postgres"
	tracingadapter "github.com/kuromii5/chat-bot-auth-service/internal/adapters/tracing"
	grpcuserhandler "github.com/kuromii5/chat-bot-auth-service/internal/handlers/grpc/user"
	httpserver "github.com/kuromii5/chat-bot-auth-service/internal/handlers/http"
	authhandler "github.com/kuromii5/chat-bot-auth-service/internal/handlers/http/auth"
	authmw "github.com/kuromii5/chat-bot-auth-service/internal/handlers/http/middleware"
	userhandler "github.com/kuromii5/chat-bot-auth-service/internal/handlers/http/user"
	"github.com/kuromii5/chat-bot-auth-service/internal/service/session"
	tracingsvc "github.com/kuromii5/chat-bot-auth-service/internal/service/tracing"
	userservice "github.com/kuromii5/chat-bot-auth-service/internal/service/user"
	"github.com/kuromii5/chat-bot-shared/jwt"
	authv1 "github.com/kuromii5/chat-bot-shared/proto/auth/v1"
	"github.com/kuromii5/chat-bot-shared/tracing"
)

type App struct {
	closer     Closer
	httpServer *httpserver.Server
	grpcServer *grpc.Server
	grpcLis    net.Listener
}

func New(ctx context.Context, cfg *config.Config) (*App, error) {
	var a App

	shutdownTracer, err := tracing.InitTracer(
		ctx,
		"auth-service",
		cfg.Tracing.Endpoint,
		cfg.Tracing.Sampler,
	)
	if err != nil {
		return nil, fmt.Errorf("init tracer: %w", err)
	}
	a.closer.Add(shutdownTracer)

	pg, err := postgres.New(&cfg.Database) //nolint:contextcheck
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}
	a.closer.Add(func(_ context.Context) error { return pg.DB.Close() })

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
	a.httpServer = httpserver.NewServer(cfg.Server.Host, cfg.Server.Port, router)
	a.closer.Add(a.httpServer.Shutdown)

	a.grpcServer = grpc.NewServer()
	authv1.RegisterUserServiceServer(a.grpcServer, grpcuserhandler.NewHandler(userSvc))
	a.closer.Add(func(_ context.Context) error {
		a.grpcServer.GracefulStop()
		return nil
	})

	grpcLis, err := (&net.ListenConfig{}).Listen(ctx, "tcp", ":"+cfg.Server.GRPCPort)
	if err != nil {
		return nil, fmt.Errorf("listen gRPC port: %w", err)
	}
	a.grpcLis = grpcLis

	return &a, nil
}

func (a *App) Run(ctx context.Context) error {
	errChan := make(chan error, 1)

	go func() {
		logrus.Infof("HTTP server listening on %s", a.httpServer.Addr())
		if err := a.httpServer.Start(); err != nil {
			errChan <- err
		}
	}()
	go func() {
		logrus.Infof("gRPC server listening on %s", a.grpcLis.Addr())
		if err := a.grpcServer.Serve(a.grpcLis); err != nil {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return nil
	}
}

func (a *App) Close(ctx context.Context) {
	shutdownCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	a.closer.Close(shutdownCtx)
}
