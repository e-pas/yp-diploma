package mware

import (
	"context"
	"log"
	"net/http"

	"yp-diploma/internal/app/config"
	"yp-diploma/internal/app/util"
)

type loginVerifyer interface {
	VerifySessionKey(ctx context.Context, key string) (string, error)
}

type LoginHandler struct {
	keyName string
	lv      loginVerifyer
}

func NewLoginHandler(keyName string, lv loginVerifyer) *LoginHandler {
	return &LoginHandler{
		keyName: keyName,
		lv:      lv,
	}
}

func (lh *LoginHandler) AuthUser(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		usercookie, err := r.Cookie(lh.keyName)
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

		key, err := lh.lv.VerifySessionKey(r.Context(), sessKey)
		if err != nil {
			log.Println("error verifying session key: ", err)
			http.Error(w, "Unautorized", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), config.ContextKeyUserID, key)
		log.Printf("User id: %s\n", key)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}
