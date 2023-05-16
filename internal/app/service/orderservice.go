package service

import (
	"context"
	"log"
	"math/rand"
	"time"
	"yp-diploma/internal/app/config"
	"yp-diploma/internal/app/model"
	"yp-diploma/internal/app/util"
)

func (s *Service) NewOrder(ctx context.Context, orderNum string) error {
	userID := ctx.Value(config.ContextKeyUserID).(string)
	if !util.LuhnCheck(orderNum) {
		return config.ErrLuhnCheckFailed
	}
	newOrder := model.Order{
		ID:      orderNum,
		UserID:  userID,
		Status:  model.Created,
		GenTime: time.Now(),
	}

	order, err := s.repo.GetOrder(ctx, orderNum)
	switch err {
	case nil:
		if order.UserID == userID {
			return config.ErrOrderRegisteredByUser
		}
		return config.ErrOrderRegistered
	case config.ErrNoSuchRecord:
		err := s.repo.AddOrder(ctx, newOrder)
		if err != nil {
			return err
		}
		s.orderPool.PushJob(newOrder)
		return nil
	default:
		return err
	}
}

func (s *Service) GetOrdersList(ctx context.Context) ([]model.Order, error) {
	userID := ctx.Value(config.ContextKeyUserID).(string)
	orderList, err := s.repo.GetUserOrders(ctx, userID)
	if err != nil {
		return nil, err
	}
	for i := range orderList {
		procOrder, err := s.orderPool.GetJobByID(orderList[i].ID)
		if err == nil {
			orderList[i] = procOrder
		}
	}
	return orderList, nil
}

func (s *Service) ServeUndoneOrders(ctx context.Context) error {
	orderList, err := s.repo.GetUndoneOrders(ctx)
	if err != nil {
		return err
	}
	for _, undoneOrder := range orderList {
		log.Printf("found undone order: %s", undoneOrder.ID)
		s.orderPool.PushJob(undoneOrder)
	}
	return nil
}
func (s *Service) GetAccrual(ctx context.Context, jobOrder model.Order) (model.Order, error) {
	res := jobOrder
	if res.Status != model.Processing {
		// Проверяем задание, если оно ранее обработано, то сразу отправляем его в ответ.
		// Так может быть, если при записи уже отработанных заданий в БД произошла ошибка,
		// то обработанные задания остаются в пуле.
		return res, nil
	}

	i := rand.Intn(10)

	time.Sleep(time.Duration(i / 2 * int(time.Second)))
	switch {
	case i > 8:
		log.Printf("order: %s, Error genereted", res.ID)
		return res, config.ErrNoSuchRecord
	case i < 3:
		res.Status = model.Invalid
		log.Printf("order: %s, status Invalid", res.ID)
		return res, nil
	default:
		res.Status = model.Processed
		res.Accrual = i * 100
		log.Printf("order: %s, status Ok, Accr:%d", res.ID, res.Accrual)
		return res, nil
	}
}

func (s *Service) SaveResults(ctx context.Context, doneOrders []model.Order) error {
	for i, order := range doneOrders {
		log.Printf("result %d/%d job:%s, status:%d, accrual:%d, user:%s", i+1, len(doneOrders), order.ID, order.Status, order.Accrual, order.UserID)
	}
	return s.repo.UpdateAccruals(ctx, doneOrders)
}
