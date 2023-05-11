package endpoint

import (
	"encoding/json"
	"io"
	"net/http"
	"time"
	"yp-diploma/internal/app/config"
	"yp-diploma/internal/app/service"
)

type Endpoint struct {
	cfg *config.Config
	srv *service.Service
}

func New(cfg *config.Config, srv *service.Service) *Endpoint {
	e := &Endpoint{}
	e.cfg = cfg
	e.srv = srv
	return e
}

func (e *Endpoint) Info(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello again!"))
}

func (e *Endpoint) Register(w http.ResponseWriter, r *http.Request) {
	buf, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	req := make(map[string]string, 0)
	err = json.Unmarshal(buf, &req)
	if err != nil || req["login"] == "" || req["password"] == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	cryptKey, err := e.srv.RegisterUser(r.Context(), req["login"], req["password"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	cookie := http.Cookie{
		Name:    config.CookieName,
		Value:   cryptKey,
		Path:    "/",
		Expires: time.Now().Add(config.SessionKeyDuration),
	}
	http.SetCookie(w, &cookie)
	w.WriteHeader(http.StatusOK)
}

func (e *Endpoint) Login(w http.ResponseWriter, r *http.Request) {
	buf, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	req := make(map[string]string, 0)
	err = json.Unmarshal(buf, &req)
	if err != nil || req["login"] == "" || req["password"] == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	cryptKey, err := e.srv.LoginUser(r.Context(), req["login"], req["password"])
	if err != nil {
		http.Error(w, "Unautorized", http.StatusUnauthorized)
		return
	}

	cookie := http.Cookie{
		Name:    config.CookieName,
		Value:   cryptKey,
		Path:    "/",
		Expires: time.Now().Add(config.SessionKeyDuration),
	}
	http.SetCookie(w, &cookie)
	w.WriteHeader(http.StatusOK)

}
