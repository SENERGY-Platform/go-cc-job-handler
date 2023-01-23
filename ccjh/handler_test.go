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
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

type testJob struct {
	mu        sync.RWMutex
	tFunc     func()
	Result    int       `json:"result"`
	Error     string    `json:"error"`
	Created   time.Time `json:"created"`
	Started   time.Time `json:"started"`
	Completed time.Time `json:"completed"`
	Canceled  time.Time `json:"canceled"`
}

func (j *testJob) CallTarget() {
	j.tFunc()
}

func (j *testJob) SetStarted(t time.Time) {
	j.mu.Lock()
	j.Started = t
	j.mu.Unlock()
}

func (j *testJob) SetCompleted(t time.Time) {
	j.mu.Lock()
	j.Completed = t
	j.mu.Unlock()
}

func (j *testJob) IsCanceled() bool {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return !j.Canceled.IsZero()
}

func testFunc(ctx context.Context, t int) (int, error) {
	if t <= 0 {
		return 0, fmt.Errorf("%d <= 0", t)
	}
	var i int
	for i = 0; i < t; i++ {
		if ctx.Err() != nil {
			return 0, ctx.Err()
		}
		time.Sleep(time.Second)
	}
	return i, nil
}

func TestIO(t *testing.T) {
	jh := New(1)
	err := jh.Add(&testJob{})
	if err != nil {
		t.Error(err)
		return
	}
	if jh.Pending() != 1 {
		t.Errorf("pending jobs != 1")
		return
	}
	err = jh.Add(&testJob{})
	if err == nil {
		t.Error("buffer error == nil")
		return
	}
	err = jh.Reset()
	if err != nil {
		t.Error(err)
		return
	}
	if jh.Pending() != 0 {
		t.Errorf("pending jobs != 0")
		return
	}
}