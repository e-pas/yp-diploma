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
	a.s = service.New(a.db, a.c)
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
		r.Post("/api/user/orders", a.e.NewOrder)
		r.Get("/api/user/orders", a.e.UserOrders)
		r.Get("/api/user/balance", a.e.UserBalance)
		r.Post("/api/user/balance/withdraw", a.e.NewWithdraw)
		r.Get("/api/user/withdrawals", a.e.UserWithdraws)
	})
	return a
}

func (a *App) Run() error {
	err := a.db.Init(context.Background(), a.c.PgConnString)
	if err != nil {
		return err
	}
	err = a.s.ServeUndoneOrders(context.Background())
	if err != nil {
		return err
	}
	log.Println("service listening at:", a.c.Listen)
	http.ListenAndServe(a.c.Listen, a.r)
	return nil
}
