package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type UserHandler interface {
	Register(http.ResponseWriter, *http.Request)
}

type AuthHandler interface {
	Login(http.ResponseWriter, *http.Request)
	Logout(http.ResponseWriter, *http.Request)
	Refresh(http.ResponseWriter, *http.Request)
}

func NewRouter(userH UserHandler, authH AuthHandler) http.Handler {
	r := chi.NewRouter()
	r.Use(
		middleware.RequestID,
		middleware.RealIP,
		middleware.Logger,
		middleware.Recoverer,
		middleware.Timeout(30*time.Second),
	)

	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/register", userH.Register)
		r.Post("/login", authH.Login)
		r.Post("/logout", authH.Logout)
		r.Post("/refresh", authH.Refresh)
	})

	return r
}
