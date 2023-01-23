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
	mu       sync.RWMutex
	tFunc    func()
	Result   int
	Error    error
	Canceled bool
}

func newTestJob(ctx context.Context, t int) (*testJob, context.CancelFunc) {
	c, cf := context.WithCancel(ctx)
	j := &testJob{}
	j.tFunc = func() {
		defer cf()
		r, e := testFunc(c, t)
		if e == nil {
			e = c.Err()
		}
		j.Result = r
		j.Error = e
	}
	return j, cf
}

func (j *testJob) CallTarget(cbk func()) {
	j.tFunc()
	cbk()
}

func (j *testJob) IsCanceled() bool {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return j.Canceled
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

func TestHandlerIO(t *testing.T) {
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

func TestHandlerJobs(t *testing.T) {
	ctx := context.Background()
	jh := New(3)
	j1, _ := newTestJob(ctx, 1)
	err := jh.Add(j1)
	if err != nil {
		t.Error(err)
		return
	}
	go func() {
		err = jh.Run(2, 50*time.Millisecond)
		if err != nil {
			t.Error(err)
			return
		}
	}()
	j2, j2cf := newTestJob(ctx, 10)
	err = jh.Add(j2)
	if err != nil {
		t.Error(err)
		return
	}
	j2cf()
	j2.mu.Lock()
	j2.Canceled = true
	j2.mu.Unlock()
	j3, _ := newTestJob(ctx, 1)
	err = jh.Add(j3)
	if err != nil {
		t.Error(err)
		return
	}
	time.Sleep(250 * time.Millisecond)
	jh.Stop()
	time.Sleep(250 * time.Millisecond)
	if jh.Running() {
		t.Errorf("running == true")
		return
	}
	if jh.Pending() != 0 {
		t.Errorf("pending jobs != 0")
		return
	}
	jh.Wait()
	if j1.Result != 1 {
		t.Errorf("j1 result != 1")
		return
	}
	if j2.Result != 0 {
		t.Errorf("j2 result != 0")
		return
	}
	if j3.Result != 1 {
		t.Errorf("j3 result != 0")
		return
	}
}
