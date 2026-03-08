package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/riandyrn/otelchi"

	authmw "github.com/kuromii5/chat-bot-auth-service/internal/handlers/http/middleware"
	"github.com/kuromii5/chat-bot-shared/wrapper"
)

type UserHandler interface {
	Register(http.ResponseWriter, *http.Request)
	UpdatePreferences(http.ResponseWriter, *http.Request)
}

type AuthHandler interface {
	Login(http.ResponseWriter, *http.Request)
	Logout(http.ResponseWriter, *http.Request)
	Refresh(http.ResponseWriter, *http.Request)
}

func NewRouter(
	userH UserHandler,
	authH AuthHandler,
	jwtSecret string,
	jail *authmw.IPJail,
) http.Handler {
	r := chi.NewRouter()
	r.Use(
		middleware.RequestID,
		otelchi.Middleware("auth-service", otelchi.WithChiRoutes(r)),
		middleware.RealIP,
		wrapper.AccessLog,
		middleware.Recoverer,
		middleware.Timeout(30*time.Second),
	)

	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/register", userH.Register)
		r.With(jail.Middleware).Post("/login", authH.Login)
		r.Post("/logout", authH.Logout)
		r.Post("/refresh", authH.Refresh)
	})

	r.Route("/api/v1/users", func(r chi.Router) {
		r.Use(authmw.Auth(jwtSecret))
		r.Patch("/preferences", userH.UpdatePreferences)
	})

	return r
}
