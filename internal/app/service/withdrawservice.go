package service

import (
	"context"
	"time"
	"yp-diploma/internal/app/config"
	"yp-diploma/internal/app/model"
)

func (s *Service) NewWithdraw(ctx context.Context, ws model.Withdraw) error {
	userID := ctx.Value(config.ContextKeyUserID).(string)
	ws.GenTime = time.Now()

	order, err := s.repo.GetOrder(ctx, ws.OrderID)
	if err != nil || order.UserID != userID {
		return config.ErrNoSuchOrder
	}

	bal, _ := s.repo.GetBalance(ctx, userID)
	if bal.Balance < ws.Withdraw {
		return config.ErrNotEnoughAccruals
	}

	return s.repo.AddWithdraw(ctx, ws)
}

func (s *Service) GetWithdrawList(ctx context.Context) ([]model.Withdraw, error) {
	userID := ctx.Value(config.ContextKeyUserID).(string)
	return s.repo.GetWithdrawList(ctx, userID)
}

func (s *Service) GetBalance(ctx context.Context) (model.Balance, error) {
	userID := ctx.Value(config.ContextKeyUserID).(string)
	bal, err := s.repo.GetBalance(ctx, userID)
	switch err {
	case nil, config.ErrNoSuchRecord:
		return bal, nil
	default:
		return model.Balance{}, err
	}
}
