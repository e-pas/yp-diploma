package config

import (
	"errors"
	"flag"
	"log"
	"time"

	"github.com/caarlos0/env/v7"
)

type Config struct {
	Listen       string `env:"RUN_ADDRESS"`
	PgConnString string `env:"DATABASE_URI"`
}

type ctxKey int

const (
	CookieName       string = "LOGININFO"
	PassCiph         string = "AF12345"
	ContextKeyUserID ctxKey = 1

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
	flag.Parse()
	return c
}

var (
	ErrNoSuchRecord = errors.New("no such record")
)
