package service

import (
	"context"
	"fmt"
	"sync"
	"yp-diploma/internal/app/model"
)

const maxWorkers = 3

type JobError struct {
	Err error
	Job model.Order
}

type Processor struct {
	jobCh   chan model.Order
	resCh   chan model.Order
	errCh   chan JobError
	workers []*worker
	wg      *sync.WaitGroup
}

type worker struct {
	name  string
	job   jobFunc
	jobCh chan model.Order
	resCh chan model.Order
	errCh chan JobError
}

func NewProcessor(job jobFunc) *Processor {
	res := &Processor{
		workers: make([]*worker, maxWorkers),
	}

	for ik := 0; ik < maxWorkers; ik++ {
		w := &worker{
			name: fmt.Sprintf("worker %d", ik),
			job:  job,
		}
		res.workers[ik] = w
	}
	return res
}

func (p *Processor) StartWorkers(ctx context.Context) {
	p.wg = &sync.WaitGroup{}
	p.jobCh = make(chan model.Order)
	p.resCh = make(chan model.Order)
	p.errCh = make(chan JobError)

	for ik := 0; ik < maxWorkers; ik++ {
		w := p.workers[ik]
		w.jobCh = p.jobCh
		w.resCh = p.resCh
		w.errCh = p.errCh
		p.wg.Add(1)
		go w.Start(ctx, p.wg)
	}
}

func (w *worker) Start(ctx context.Context, wg *sync.WaitGroup) {
	for order := range w.jobCh {
		newOrder, err := w.job(ctx, order)
		if err != nil {
			w.errCh <- JobError{Err: err, Job: order}
		} else {
			w.resCh <- newOrder
		}
	}
	wg.Done()
}

func (p *Processor) ProceedWith(ctx context.Context, jobs []model.Order) ([]model.Order, []JobError) {
	res := make([]model.Order, 0)
	err := make([]JobError, 0)

	//  colecting results, errors
	resWg := sync.WaitGroup{}
	resWg.Add(2)
	go func() {
		for r := range p.resCh {
			res = append(res, r)
		}
		resWg.Done()
	}()
	go func() {
		for e := range p.errCh {
			err = append(err, e)
		}
		resWg.Done()
	}()

	p.jobCh = make(chan model.Order)
	p.StartWorkers(ctx)
	go func() {
	loop:
		for i := range jobs {
			select {
			case p.jobCh <- jobs[i]:
				continue
			case <-ctx.Done():
				break loop
			}
		}
		close(p.jobCh)
	}()
	p.wg.Wait()
	close(p.resCh)
	close(p.errCh)
	resWg.Wait()
	return res, err
}
