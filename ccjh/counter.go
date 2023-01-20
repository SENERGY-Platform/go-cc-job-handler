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

import "sync"

type counter struct {
	c  int
	mu sync.RWMutex
}

func (c *counter) Increase() {
	c.mu.Lock()
	c.c++
	c.mu.Unlock()
}

func (c *counter) Decrease() {
	c.mu.Lock()
	c.c--
	c.mu.Unlock()
}

func (c *counter) Reset() {
	c.mu.Lock()
	c.c = 0
	c.mu.Unlock()
}

func (c *counter) Value() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.c
}
