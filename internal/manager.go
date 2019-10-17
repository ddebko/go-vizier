package internal

import (
	"fmt"
	"math"
)

type Manager struct {
	_          struct{}
	name       string
	states     map[string]IState
	size       int
	run        bool
	stopWorker chan bool
}

func (m *Manager) Node(name string, f func(interface{}) map[string]interface{}) *Manager {
	if _, ok := m.states[name]; ok {
		detail := fmt.Sprintf("failed to create state %s.", name)
		panic(NewVizierError(ErrSourceManager, ErrMsgStateAlreadyExists, detail))
	}

	m.states[name] = NewState(name, f)

	return m
}

func (m *Manager) Edge(from, to, name string) *Manager {
	edgeName := fmt.Sprintf("%s_to_%s_%s", from, to, name)

	if _, ok := m.states[to]; !ok {
		detail := fmt.Sprintf("failed to create edge %s. destination state %s", name, to)
		panic(NewVizierError(ErrSourceManager, ErrMsgStateDoesNotExist, detail))
	}

	if _, ok := m.states[from]; !ok {
		detail := fmt.Sprintf("failed to create edge %s. source state %s", name, from)
		panic(NewVizierError(ErrSourceManager, ErrMsgStateDoesNotExist, detail))
	}

	if m.states[from].HasEdge(edgeName) {
		detail := fmt.Sprintf("failed to create edge %s. source state %s", name, from)
		panic(NewVizierError(ErrSourceManager, ErrMsgEdgeAlreadyExists, detail))
	}

	err := m.states[from].AttachEdge(edgeName, m.states[to].GetPipe())
	if err != nil {
		panic(err)
	}

	return m
}

func (m *Manager) Invoke(name string, payload interface{}) error {
	if _, ok := m.states[name]; !ok {
		return nil
	}

	m.states[name].Invoke(payload)

	return nil
}

func (m *Manager) Stop() error {
	if !m.run {
		return fmt.Errorf("")
	}
	m.run = false
	return nil
}

func (m *Manager) GetSize() int {
	return m.size
}

func (m *Manager) SetSize(size int) error {
	if !m.run {
		return fmt.Errorf("")
	}

	if size <= 0 {
		detail := fmt.Sprintf("%s. invalid size %d", m.name, size)
		return fmt.Errorf(detail)
	}

	delta := int(math.Abs(float64(m.size - size)))
	spawn := (size > m.size)
	for i := 0; i < delta; i++ {
		if spawn {
			m.spawnWorker()
			continue
		}
		m.stopWorker <- true
	}
	m.size = size

	return nil
}

func (m *Manager) spawnWorker() {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				m.spawnWorker()
			}
		}()
		for m.run {
			select {
			case <-m.stopWorker:
				return
			default:
				for _, state := range m.states {
					state.Poll()
				}
			}
		}
	}()
}

func NewManager(name string, size int) (*Manager, error) {
	states := make(map[string]IState)

	manager := Manager{
		name:       name,
		states:     states,
		size:       size,
		stopWorker: make(chan bool, size),
	}

	return &manager, nil
}
