package service

import (
	"context"
	"time"
	"yp-diploma/internal/app/config"
	"yp-diploma/internal/app/model"
	"yp-diploma/internal/app/util"
)

func (s *Service) NewWithdraw(ctx context.Context, ws model.Withdraw) error {
	userID := getUserIDFromCtx(ctx)
	ws.UserID = userID
	ws.GenTime = time.Now()

	if !util.LuhnCheck(ws.OrderID) {
		return config.ErrLuhnCheckFailed
	}

	bal, _ := s.repo.GetBalance(ctx, userID)
	if bal.Balance < ws.Withdraw {
		return config.ErrNotEnoughAccruals
	}

	return s.repo.AddWithdraw(ctx, ws)
}

func (s *Service) GetWithdrawList(ctx context.Context) ([]model.Withdraw, error) {
	userID := getUserIDFromCtx(ctx)
	return s.repo.GetWithdrawList(ctx, userID)
}

func (s *Service) GetBalance(ctx context.Context) (model.Balance, error) {
	userID := getUserIDFromCtx(ctx)
	bal, err := s.repo.GetBalance(ctx, userID)
	switch err {
	case nil, config.ErrNoSuchRecord:
		return bal, nil
	default:
		return model.Balance{}, err
	}
}
