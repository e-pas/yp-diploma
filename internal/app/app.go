package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
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
	a.s = service.New(a.db, a.c)
	a.e = endpoint.New(a.c, a.s)
	a.lh = mware.NewLoginHandler(config.CookieName, a.s)
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
	err = a.s.ProceedUndoneOrders(context.Background())
	if err != nil {
		return err
	}

	server := newServer(a.c.Listen, a.r)
	go func() {
		log.Println("service listening at:", a.c.Listen)
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
		log.Println("server gracefully shut down")
	}()
	waitForShutDown(server)

	return nil
}

func newServer(addr string, r chi.Router) *http.Server {
	return &http.Server{
		Addr:    addr,
		Handler: r,
	}
}

func waitForShutDown(server *http.Server) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := server.Shutdown(ctx)
	if err != nil {
		log.Fatal("failed shut down server")
	}
}
