package mware

import (
	"context"
	"log"
	"net/http"
	"time"

	"yp-diploma/internal/app/config"
	"yp-diploma/internal/app/model"
	"yp-diploma/internal/app/util"
)

type loginsaver interface {
	GetSessKey(ctx context.Context, key string) (model.SessKey, error)
}

type LoginHandler struct {
	repo loginsaver
}

func NewLoginHandler(repo loginsaver) *LoginHandler {
	return &LoginHandler{
		repo: repo,
	}
}

func (lh *LoginHandler) AuthUser(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		usercookie, err := r.Cookie(config.CookieName)
		if err != nil {
			http.Error(w, "Unautorized", http.StatusUnauthorized)
			return
		}

		cryptKey := usercookie.Value
		signKey, err := util.DecodeString(cryptKey)
		if err != nil {
			log.Printf("cookie damaged: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		sessKey, ok := util.UnsignString(signKey)
		if !ok {
			log.Println("sign damaged")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		key, err := lh.repo.GetSessKey(r.Context(), sessKey)
		if err != nil || time.Now().After(key.Expires) {
			log.Println("error geting session key or session key expired", err)
			http.Error(w, "Unautorized", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), config.ContextKeyUserID, key.UserID)
		log.Printf("User id: %s", key.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}
