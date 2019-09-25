package vizier

import (
	"fmt"
	"math"
	"sync"

	log "github.com/sirupsen/logrus"
)

type Pool struct {
	_          struct{}
	wg         sync.WaitGroup
	states     map[string]IState
	name       string
	size       int
	run        bool
	stopWorker chan bool
}

func (p *Pool) Create() error {
	if len(p.states) <= 0 {
		return NewVizierError(ErrSourcePool, ErrMsgPoolEmptyStates, p.name)
	}

	for i := 0; i < p.size; i++ {
		p.spawnWorker()
	}

	return nil
}

func (p *Pool) spawnWorker() {
	log.WithFields(log.Fields{
		"source": "pool",
		"name":   p.name,
	}).Info("worker spawned")
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		defer func() {
			if err := recover(); err != nil {
				log.WithFields(log.Fields{
					"source": "pool",
					"name":   p.name,
					"err":    err,
				}).Warn("worker panic")
				p.spawnWorker()
			}
		}()
		for p.run {
			select {
			case <-p.stopWorker:
				return
			default:
				for _, state := range p.states {
					state.Run()
				}
			}
		}
	}()
}

func (p *Pool) Stop() vizierErr {
	if !p.run {
		return NewVizierError(ErrSourcePool, ErrMsgPoolNotRunning, p.name)
	}
	log.WithFields(log.Fields{
		"source": "pool",
		"name":   p.name,
	}).Info("stop pool")
	p.run = false
	return nil
}

func (p *Pool) Wait() vizierErr {
	if !p.run {
		return NewVizierError(ErrSourcePool, ErrMsgPoolNotRunning, p.name)
	}
	p.wg.Wait()
	return nil
}

func (p *Pool) SetSize(size int) vizierErr {
	if !p.run {
		return NewVizierError(ErrSourcePool, ErrMsgPoolNotRunning, p.name)
	}

	if size <= 0 {
		detail := fmt.Sprintf("%s. invalid size %d", p.name, size)
		return NewVizierError(ErrSourcePool, ErrMsgPoolSizeInvalid, detail)
	}

	log.WithFields(log.Fields{
		"source":  "pool",
		"name":    p.name,
		"oldsize": p.size,
		"newsize": size,
	}).Info("set pool size")

	delta := int(math.Abs(float64(p.size - size)))
	spawn := (size > p.size)
	for i := 0; i < delta; i++ {
		if spawn {
			p.spawnWorker()
			continue
		}
		p.stopWorker <- true
	}
	p.size = size

	return nil
}

func NewPool(name string, size int, states map[string]IState) (*Pool, vizierErr) {
	if size <= 0 {
		return nil, NewVizierError(ErrSourcePool, ErrMsgPoolSizeInvalid, name)
	}

	log.WithFields(log.Fields{
		"source": "pool",
		"name":   name,
		"size":   size,
	}).Info("created pool")

	pool := Pool{
		states:     states,
		name:       name,
		size:       size,
		run:        true,
		stopWorker: make(chan bool),
	}

	return &pool, nil
}
