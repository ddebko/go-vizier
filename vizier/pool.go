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
	p.log(log.Fields{}).Info("worker spawned")
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		defer func() {
			if err := recover(); err != nil {
				p.log(log.Fields{"err": err}).Info("worker panic")
				p.spawnWorker()
			}
		}()
		for p.run {
			select {
			case <-p.stopWorker:
				return
			default:
				for _, state := range p.states {
					state.Poll()
				}
			}
		}
	}()
}

func (p *Pool) Stop() vizierErr {
	if !p.run {
		return NewVizierError(ErrSourcePool, ErrMsgPoolNotRunning, p.name)
	}
	p.log(log.Fields{}).Info("stop pool")
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

func (p *Pool) GetSize() int {
	return p.size
}

func (p *Pool) SetSize(size int) vizierErr {
	if !p.run {
		return NewVizierError(ErrSourcePool, ErrMsgPoolNotRunning, p.name)
	}

	if size <= 0 {
		detail := fmt.Sprintf("%s. invalid size %d", p.name, size)
		return NewVizierError(ErrSourcePool, ErrMsgPoolSizeInvalid, detail)
	}

	p.log(log.Fields{"new_size": size}).Info("set pool size")

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

func (p *Pool) log(fields log.Fields) *log.Entry {
	fields["source"] = "pool"
	fields["name"] = p.name
	fields["size"] = p.size
	return log.WithFields(fields)
}

func NewPool(name string, size int, states map[string]IState) (*Pool, vizierErr) {
	if size <= 0 {
		return nil, NewVizierError(ErrSourcePool, ErrMsgPoolSizeInvalid, name)
	}

	pool := Pool{
		states:     states,
		name:       name,
		size:       size,
		run:        true,
		stopWorker: make(chan bool),
	}

	pool.log(log.Fields{}).Info("created pool")

	return &pool, nil
}
