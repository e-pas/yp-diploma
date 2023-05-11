package app

import (
	"context"
	"log"
	"net/http"
	"yp-diploma/internal/app/config"
	"yp-diploma/internal/app/endpoint"
	"yp-diploma/internal/app/mware"
	"yp-diploma/internal/app/repository"
	"yp-diploma/internal/app/service"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

type App struct {
	c  *config.Config
	s  *service.Service
	e  *endpoint.Endpoint
	db *repository.Repository
	lh *mware.LoginHandler
	r  chi.Router
}

func New() *App {
	a := &App{}
	a.c = config.New()
	a.db = repository.New()
	a.lh = mware.NewLoginHandler(a.db)
	a.s = service.New(a.db)
	a.e = endpoint.New(a.c, a.s)
	a.r = chi.NewRouter()

	a.r.Use(middleware.RealIP)
	a.r.Use(middleware.Logger)
	a.r.Use(middleware.Recoverer)

	a.r.Use(mware.GunzipRequest)
	a.r.Use(mware.GzipResponse)

	a.r.Post("/api/user/register", a.e.Register)
	a.r.Post("/api/user/login", a.e.Login)

	a.r.Group(func(r chi.Router) {
		r.Use(a.lh.AuthUser)
		r.Get("/", a.e.Info)
	})
	return a
}

func (a *App) Run() error {
	err := a.db.Init(context.Background(), a.c.PgConnString)
	if err != nil {
		return err
	}
	log.Println("service listening at:", a.c.Listen)
	http.ListenAndServe(a.c.Listen, a.r)
	return nil
}
