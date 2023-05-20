package endpoint

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
	"yp-diploma/internal/app/config"
	"yp-diploma/internal/app/model"
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
		http.Error(w, "internal server error", http.StatusInternalServerError)
		log.Printf("error creating user\n error: %s", err.Error())
		return
	}
	defer r.Body.Close()
	req := make(map[string]string, 0)
	err = json.Unmarshal(buf, &req)
	if err != nil || req["login"] == "" || req["password"] == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	cryptKey, err := e.srv.RegisterUser(r.Context(), req["login"], req["password"])
	if err != nil {
		switch err {
		case config.ErrUserNameBusy:
			http.Error(w, err.Error(), http.StatusConflict)
			return
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
			log.Printf("error creating user: %s\n error: %s", req["login"], err.Error())
			return
		}
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
	defer r.Body.Close()
	req := make(map[string]string, 0)
	err = json.Unmarshal(buf, &req)
	if err != nil || req["login"] == "" || req["password"] == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	cryptKey, err := e.srv.LoginUser(r.Context(), req["login"], req["password"])
	if err != nil {
		switch err {
		case config.ErrUserInvalidPassword, config.ErrNoSuchRecord:
			http.Error(w, "Unautorized", http.StatusUnauthorized)
			return
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
			log.Printf("error autorize user: %s\n error: %s", req["login"], err.Error())
			return
		}
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

func (e *Endpoint) NewOrder(w http.ResponseWriter, r *http.Request) {
	buf, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		log.Printf("error creating order:\n error: %s", err.Error())
		return
	}
	defer r.Body.Close()
	str := strings.Split(string(buf), "\n")
	for _, num := range str {
		num = strings.TrimSpace(string(num))
		err = e.srv.NewOrder(r.Context(), num)
		if err != nil {
			switch err {
			case config.ErrOrderRegisteredByUser:
				http.Error(w, err.Error(), http.StatusOK)
			case config.ErrLuhnCheckFailed:
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			case config.ErrOrderRegistered:
				http.Error(w, err.Error(), http.StatusConflict)
			default:
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
			return
		}
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte(fmt.Sprintf("%s - Ok \n", num)))
	}
}

func (e *Endpoint) UserOrders(w http.ResponseWriter, r *http.Request) {
	res, err := e.srv.GetOrdersList(r.Context())
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		log.Printf("error getting orders:\n error: %s", err)
		return
	}
	if len(res) == 0 {
		http.Error(w, "no data", http.StatusNoContent)
		return
	}
	buf := model.MarshalUserOrdersDoc(res)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(buf)
}

func (e *Endpoint) NewWithdraw(w http.ResponseWriter, r *http.Request) {
	buf, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	wd, err := model.UnmarshalWithdrawRequest(buf)
	if err != nil || wd.OrderID == "" || wd.Withdraw <= 0 {
		http.Error(w, "error in request", http.StatusUnprocessableEntity)
		return
	}
	err = e.srv.NewWithdraw(r.Context(), wd)
	if err != nil {
		switch err {
		case config.ErrLuhnCheckFailed:
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		case config.ErrNotEnoughAccruals:
			http.Error(w, err.Error(), http.StatusPaymentRequired)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Ok"))
}

func (e *Endpoint) UserWithdraws(w http.ResponseWriter, r *http.Request) {
	res, err := e.srv.GetWithdrawList(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(res) == 0 {
		http.Error(w, "no data", http.StatusNoContent)
		return
	}
	buf := model.MarshalUserWithdrawsDoc(res)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(buf)
}

func (e *Endpoint) UserBalance(w http.ResponseWriter, r *http.Request) {
	res, err := e.srv.GetBalance(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	buf := model.MarshalUserBalanceDoc(res)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(buf)
}
