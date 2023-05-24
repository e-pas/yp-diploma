package service

import (
	"context"
	"log"
	"sync"
	"time"
	"yp-diploma/internal/app/config"
	"yp-diploma/internal/app/model"
)

type jobFunc = func(ctx context.Context, jobOrder model.Order) (model.Order, error)
type resultFunc = func(ctx context.Context, jobOrders []model.Order) ([]model.Order, error)

type jobPool struct {
	pool map[string]model.Order
	mu   sync.RWMutex
}

type jobDispatcher struct {
	ctx       context.Context
	jobsPool  *jobPool
	processor *Processor
	resFunc   resultFunc
}

func NewDispatcher(ctx context.Context, jpool *jobPool, jFunc jobFunc, rFunc resultFunc) *jobDispatcher {
	jd := &jobDispatcher{}
	jd.ctx = ctx
	jd.jobsPool = jpool
	jd.resFunc = rFunc
	jd.processor = NewProcessor(jFunc)

	jd.Dispatch()
	return jd
}

func (jd *jobDispatcher) Dispatch() {
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				// берем все заказы в пуле ну обработку.
				jobs := jd.jobsPool.CopyState()
				if len(jobs) == 0 {
					continue
				}
				// Ограничиваем количество обращений к сервису начисления баллов.
				// По условиям ТЗ установлено ограничение не более 60 обращений в мин.
				// В тестовом скрипте, дается только 20 секунд для подключения к сервису
				// начисления баллов. Поэтому,делим минутный интервал на 15 секундные отрезки
				// и максимальное количество обращений к серверу делим на 4
				jobsPerTick := int(config.OrdersPerMinuteToAccrual / 4)
				if len(jobs) > jobsPerTick {
					jobs = jobs[:jobsPerTick]
				}
				for i := range jobs {
					// Меняем статус у новых заданий на "Processing" и обновляем его в пуле.
					if jobs[i].Status == model.Created {
						jobs[i].Status = model.Processing
						jd.jobsPool.PushJob(jobs[i])
					}
				}
				jobs, errs := jd.processor.ProceedWith(jd.ctx, jobs)
				if len(errs) > 0 {
					// обрабатываем полученные ошибки,
					// задания, на которых они возникли, остаются в пуле для повторной обработки
					for i, err := range errs {
						log.Printf("found err %d/%d for order:%s, err:%v",
							i+1, len(errs), err.Job.ID, err.Err)
					}
				}

				// полученные результаты предаем во вторую callback функцию, и, есди она завершилась
				// без ошибок, удаляем обработанные задания из пула.
				jobs, err := jd.resFunc(jd.ctx, jobs)
				if err == nil {
					for _, job := range jobs {
						jd.jobsPool.Delete(job)
					}
				}

			case <-jd.ctx.Done():
				log.Println("dispatcher stop")
				return
			}
		}
	}()
}

func NewJobPool() *jobPool {
	jp := &jobPool{}
	jp.pool = make(map[string]model.Order, 0)

	return jp
}
func (jp *jobPool) PushJob(newJob model.Order) {
	jp.mu.Lock()
	defer jp.mu.Unlock()
	jp.pool[newJob.ID] = newJob
}

func (jp *jobPool) Delete(job model.Order) {
	jp.mu.Lock()
	defer jp.mu.Unlock()
	delete(jp.pool, job.ID)
}

func (jp *jobPool) GetJobByID(id string) (model.Order, error) {
	jp.mu.RLock()
	defer jp.mu.RUnlock()
	res, ok := jp.pool[id]
	if !ok {
		return model.Order{}, config.ErrNoSuchRecord
	}
	return res, nil
}

func (jp *jobPool) CopyState() []model.Order {
	jp.mu.RLock()
	defer jp.mu.RUnlock()
	res := make([]model.Order, 0)
	for _, el := range jp.pool {
		res = append(res, el)
	}
	return res
}
