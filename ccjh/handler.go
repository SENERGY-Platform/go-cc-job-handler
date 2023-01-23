/*
 * Copyright 2023 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package ccjh

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type Handler struct {
	pJobs   chan Job
	jCount  counter
	jWG     sync.WaitGroup
	sChan   chan bool
	running bool
	ticker  *time.Ticker
	mu      sync.Mutex
}

func New(buffer int) *Handler {
	return &Handler{
		pJobs:  make(chan Job, buffer),
		jCount: counter{c: 0},
		sChan:  make(chan bool, 1),
	}
}

func (h *Handler) Add(job Job) error {
	select {
	case h.pJobs <- job:
	default:
		return fmt.Errorf("buffer full")
	}
	return nil
}

func (h *Handler) Run(maxJobs int, interval time.Duration) error {
	return h.run(maxJobs, interval, false)
}

func (h *Handler) RunAsync(maxJobs int, interval time.Duration) error {
	return h.run(maxJobs, interval, true)
}

func (h *Handler) Running() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.running
}

func (h *Handler) Stop() {
	h.mu.Lock()
	if h.running {
		h.ticker.Stop()
		h.sChan <- true
	}
	h.mu.Unlock()
}

func (h *Handler) Wait() {
	h.jWG.Wait()
}

func (h *Handler) Pending() int {
	return len(h.pJobs)
}

func (h *Handler) Reset() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.running {
		return errors.New("can't reset while running")
	}
	for len(h.pJobs) > 0 {
		<-h.pJobs
	}
	return nil
}

func (h *Handler) run(maxJobs int, interval time.Duration, async bool) error {
	if !h.Running() {
		h.mu.Lock()
		if h.ticker == nil {
			h.ticker = time.NewTicker(interval)
		} else {
			h.ticker.Reset(interval)
		}
		h.mu.Unlock()
		if async {
			go h.loop(maxJobs)
		} else {
			h.loop(maxJobs)
		}
	} else {
		return errors.New("already running")
	}
	return nil
}

func (h *Handler) loop(maxJobs int) {
	h.mu.Lock()
	h.running = true
	h.mu.Unlock()
	stop := false
	for !stop {
		select {
		case stop = <-h.sChan:
		case <-h.ticker.C:
			if maxJobs == 0 || h.jCount.Value() < maxJobs {
				select {
				case j := <-h.pJobs:
					if !j.IsCanceled() {
						h.jWG.Add(1)
						h.jCount.Increase()
						go func() {
							defer h.jWG.Done()
							j.SetStarted(time.Now().UTC())
							j.CallTarget()
							j.SetCompleted(time.Now().UTC())
							h.jCount.Decrease()
						}()
					}
				default:
				}
			}
		}
	}
	h.mu.Lock()
	h.running = false
	h.mu.Unlock()
}
