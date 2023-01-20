package ccjh

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type job struct {
	ctx   context.Context
	cFunc context.CancelFunc
	tFunc TargetFunc
	tArgs map[string]any
	meta  JobMeta
}

type Handler struct {
	ctx     context.Context
	pJobs   chan *job
	jCount  counter
	sChan   chan bool
	running bool
	ticker  *time.Ticker
	mu      sync.Mutex
}

func New(ctx context.Context, buffer int) *Handler {
	return &Handler{
		ctx:    ctx,
		pJobs:  make(chan *job, buffer),
		jCount: counter{},
		sChan:  make(chan bool, 1),
	}
}

func (h *Handler) Add(tFunc TargetFunc, tArgs map[string]any, jMeta JobMeta) (context.CancelFunc, error) {
	ctx, cf := context.WithCancel(h.ctx)
	j := &job{
		ctx:   ctx,
		tFunc: tFunc,
		tArgs: tArgs,
		meta:  jMeta,
	}
	select {
	case h.pJobs <- j:
	default:
		return cf, fmt.Errorf("buffer full")
	}
	return cf, nil
}

func (h *Handler) Run(maxJobs int, interval time.Duration) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.running {
		if h.ticker == nil {
			h.ticker = time.NewTicker(interval)
		} else {
			h.ticker.Reset(interval)
		}
		go func() {
			stop := false
			for !stop {
				select {
				case stop = <-h.sChan:
					h.mu.Lock()
					h.running = false
					h.mu.Unlock()
				case <-h.ticker.C:
					if h.jCount.Value() < maxJobs {
						select {
						case j := <-h.pJobs:
							if !j.meta.IsCanceled() {
								e := h.start(j)
								if e != nil {
									fmt.Println(e)
								}
							}
						default:
						}
					}
				}
			}
		}()
		h.running = true
	} else {
		return errors.New("already running")
	}
	return nil
}

func (h *Handler) Stop() {
	h.mu.Lock()
	if h.running {
		h.ticker.Stop()
		h.sChan <- true
	}
	h.mu.Unlock()
}

func (h *Handler) start(j *job) error {
	h.jCount.Increase()
	go func() {
		j.meta.SetStarted(time.Now().UTC())
		r, e := j.tFunc(j.ctx, j.tArgs)
		if e == nil {
			e = j.ctx.Err()
		}
		if e != nil {
			j.meta.SetError(e)
		} else {
			j.meta.SetResult(r)
		}
		j.meta.SetCompleted(time.Now().UTC())
		h.jCount.Decrease()
	}()
	return nil
}
