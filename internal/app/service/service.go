package service

import (
	"context"

	"yp-diploma/internal/app/config"
	"yp-diploma/internal/app/repository"
)

type Service struct {
	repo      *repository.Repository
	conf      *config.Config
	orderPool *jobPool
	orderDisp *jobDispatcher
}

func New(repo *repository.Repository, conf *config.Config) *Service {
	s := &Service{}
	s.repo = repo
	s.conf = conf
	s.orderPool = NewJobPool()
	s.orderDisp = NewDispatcher(context.Background(), s.orderPool, s.GetAccrual, s.SaveResults)
	return s
}
