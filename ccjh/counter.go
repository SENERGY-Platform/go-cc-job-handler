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
