package pool

import (
	"sync"
)

// HackPool ...
// TODO: implement tests
type HackPool struct {
	numGo    int
	messages chan interface{}
	function func(interface{})
}

// New ...
// TODO: implement tests
func New(numGoroutine int, function func(interface{})) *HackPool {
	return &HackPool{
		numGo:    numGoroutine,
		messages: make(chan interface{}),
		function: function,
	}
}

// Push ...
// TODO: implement tests
func (c *HackPool) Push(data interface{}) {
	c.messages <- data
}

// CloseQueue ...
// TODO: implement tests
func (c *HackPool) CloseQueue() {
	close(c.messages)
}

// Run ...
// TODO: implement tests
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
