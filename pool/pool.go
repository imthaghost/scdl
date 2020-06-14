package pool

import (
	"sync"
)

// TODO implement tests
// HackPool ...
type HackPool struct {
	numGo    int
	messages chan interface{}
	function func(interface{})
}

// TODO implement tests
// New ...
func New(numGoroutine int, function func(interface{})) *HackPool {
	return &HackPool{
		numGo:    numGoroutine,
		messages: make(chan interface{}),
		function: function,
	}
}

// TODO implement tests
// Push ...
func (c *HackPool) Push(data interface{}) {
	c.messages <- data
}

// TODO implement tests
// CloseQueue ...
func (c *HackPool) CloseQueue() {
	close(c.messages)
}

// TODO implement tests
// Run ...
func (c *HackPool) Run() {
	var wg sync.WaitGroup

	wg.Add(c.numGo)

	for i := 0; i < c.numGo; i++ {
		go func() {
			for v := range c.messages {
				c.function(v)
			}
			wg.Done()
		}()
	}

	wg.Wait()
}
