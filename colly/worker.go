// worker pool
package colly

import (
	"sync"
)

type Worker interface {
	Runner()
}

type Pool struct {
	work chan Worker
	wg   sync.WaitGroup
}

func NewWorkPool(poolSize int) *Pool {
	p := Pool{
		work: make(chan Worker),
	}
	p.wg.Add(poolSize)
	for i := 0; i < poolSize; i++ {
		go func() {
			for task := range p.work {
				task.Runner()
			}
			p.wg.Done()
		}()
	}
	return &p
}

func (p *Pool) Run(worker Worker) {
	p.work <- worker
}

func (p *Pool) ShutDown() {
	close(p.work)
	p.wg.Wait()
}
