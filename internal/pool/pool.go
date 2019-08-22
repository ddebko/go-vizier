package pool

import (
	"errors"
	"math"
	"runtime"
	"sync"
)

var (
	ErrPoolNotRunning    = errors.New("the pool is not running")
	ErrTooManyPanics     = errors.New("the pool is panicing")
	ErrInvalidPoolSize   = errors.New("the pool size must be greater than 0")
	ErrInvalidRetryValue = errors.New("the pool retries value cannot be negative")
)

type Pool struct {
	size     int
	retries  int
	status   bool
	run      bool
	stop     chan bool
	complete chan bool
	wg       sync.WaitGroup
	task     func(interface{}) interface{}
}

func (p *Pool) Create(size int, retries int, task func(interface{}) interface{}) error {
	if size <= 0 {
		return ErrInvalidPoolSize
	}

	if retries < 0 {
		return ErrInvalidRetryValue
	}

	p.run = true
	p.status = false
	p.size = size
	p.retries = retries
	p.stop = make(chan bool)
	p.task = task

	p.spawnWorkers(size)

	return nil
}

func (p *Pool) Wait() error {
	if !p.run {
		return ErrPoolNotRunning
	}
	p.wg.Wait()
	return nil
}

func (p *Pool) SetSize(size int) error {
	if !p.run {
		return ErrPoolNotRunning
	}

	if size <= 0 {
		return ErrInvalidPoolSize
	}

	delta := int(math.Abs(float64(p.size - size)))
	if size > p.size {
		p.spawnWorkers(delta)
	} else {
		p.releaseWorkers(delta)
	}

	return nil
}

func (p *Pool) Stop() error {
	if !p.run {
		return ErrPoolNotRunning
	}
	p.run = false
	return nil
}

func (p *Pool) spawnWorkers(size int) {
	for i := 0; i < size; i++ {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			for p.run {
				select {
				case <-p.stop:
					return
				}
			}
		}()
	}
}

func (p *Pool) releaseWorkers(size int) {
	for i := 0; i < size; i++ {
		p.stop <- true
	}
}

func OptimalPoolSize() int {
	return (runtime.NumCPU() * 2) - 1
}
