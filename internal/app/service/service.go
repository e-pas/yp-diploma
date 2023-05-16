package service

import (
	"context"

	"yp-diploma/internal/app/repository"
)

type Service struct {
	repo      *repository.Repository
	orderPool *jobPool
	orderDisp *jobDispatcher
}

func New(repo *repository.Repository) *Service {
	s := &Service{}
	s.repo = repo
	s.orderPool = NewJobPool()
	s.orderDisp = NewDispatcher(context.Background(), s.orderPool, s.GetAccrual, s.SaveResults)
	return s
}
