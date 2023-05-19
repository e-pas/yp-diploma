package service

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
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
	// Проверяем есть ли заказы покупателя еще в процессе обработки
	for i := range orderList {
		procOrder, err := s.orderPool.GetJobByID(orderList[i].ID)
		if err == nil {
			// если такие заказы есть, то берем этот заказ со статусом из очереди обработки
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

// callback функция выполняется процессором для каждого заказа
func (s *Service) GetAccrual(ctx context.Context, jobOrder model.Order) (model.Order, error) {
	res := jobOrder
	if res.Status != model.Processing {
		// Проверяем задание, если оно ранее обработано, то сразу отправляем его в ответ.
		// Так может быть, если при записи уже отработанных заданий в БД произошла ошибка,
		// то обработанные задания остаются в пуле.
		return res, nil
	}

	addr := fmt.Sprintf("%s/api/orders/%s", s.conf.AccrualSystem, res.ID)
	resp, err := http.Get(addr)
	if err != nil {
		log.Printf("get error: %w", err)
		return res, config.ErrGetAccrual
	}
	if resp.StatusCode != http.StatusOK {
		log.Printf("order: %s, accrual server return status code: %d", res.ID, resp.StatusCode)
		return res, config.ErrNoSuchOrder
	}

	log.Printf("order: %s, accrual server return status code: %d", res.ID, resp.StatusCode)
	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("read response body error: %s", err.Error())
		return res, config.ErrUnsupportedResponse
	}
	AccrRes, err := model.UnmarshalAcrrualResponse(buf)
	if err != nil {
		return res, err
	}
	switch AccrRes.Status {
	case "REGISTERED", "PROCESSING":
		// ничего не делаем, ждем следующей итерации
		return jobOrder, nil
	case "INVALID":
		res.Status = model.Invalid
		log.Printf("order: %s, status Invalid", res.ID)
		return res, nil
	case "PROCESSED":
		res.Status = model.Processed
		res.Accrual = int(AccrRes.Accrual)
		log.Printf("order: %s, status Ok, Accr:%d", res.ID, res.Accrual)
		return res, nil
	default:
		log.Printf("unsopported response body: \n%s ", string(buf))
		return res, config.ErrUnsupportedResponse
	}
}

// callback функция, получает обработанные заказы,
// сохраняет в БД зазказы с окончательными результатами
// возвращает список сохраненных заказов
func (s *Service) SaveResults(ctx context.Context, doneOrders []model.Order) ([]model.Order, error) {
	saveOrders := make([]model.Order, 0)
	for i, order := range doneOrders {
		log.Printf("result %d/%d job:%s, status:%d, accrual:%d, user:%s", i+1, len(doneOrders), order.ID, order.Status, order.Accrual, order.UserID)
		switch order.Status {
		case model.Processed, model.Invalid:
			saveOrders = append(saveOrders, order)
		default:
			continue
		}
	}
	err := s.repo.UpdateAccruals(ctx, saveOrders)
	if err != nil {
		return nil, err
	}
	return saveOrders, nil
}
