package config

import (
	"errors"
	"flag"
	"log"
	"time"

	"github.com/caarlos0/env/v7"
)

type Config struct {
	Listen        string `env:"RUN_ADDRESS"`
	PgConnString  string `env:"DATABASE_URI"`
	AccrualSystem string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

type ctxKey string

const (
	CookieName       string = "LOGININFO"
	PassCiph         string = "AF12345"
	ContextKeyUserID ctxKey = ctxKey(CookieName)

	SessionKeyDuration time.Duration = 30 * 24 * time.Hour
)

func New() *Config {
	c := &Config{}
	err := env.Parse(c)
	if err != nil {
		log.Fatal(err)
	}
	flag.StringVar(&c.Listen, "a", c.Listen, "HTTP listen addr")
	flag.StringVar(&c.PgConnString, "d", c.PgConnString, "Postgres connect URL")
	flag.StringVar(&c.AccrualSystem, "r", c.AccrualSystem, "Accrual system URL")
	flag.Parse()
	if c.Listen == "" || c.PgConnString == "" || c.AccrualSystem == "" {
		log.Fatal("not enought parameters to work.")
	}
	return c
}

var (
	ErrUserNameBusy          = errors.New("user name busy")
	ErrUserInvalidPassword   = errors.New("invalid password")
	ErrInvalidData           = errors.New("invalid data")
	ErrNoSuchRecord          = errors.New("no such record")
	ErrLuhnCheckFailed       = errors.New("luhn check failed")
	ErrOrderRegisteredByUser = errors.New("same order registered by customer")
	ErrOrderRegistered       = errors.New("same order registered in system")
	ErrNoSuchOrder           = errors.New("customer does not have such an order")
	ErrNotEnoughAccruals     = errors.New("not enough accruals")
	ErrGetAccrual            = errors.New("can't get accrual information")
	ErrUnsupportedResponse   = errors.New("accrual server return unsupported result")
)
