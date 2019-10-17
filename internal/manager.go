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

func (m *Manager) Output(from, name string) chan Packet {
	if _, ok := m.states[from]; !ok {
		detail := fmt.Sprintf("failed to create output edge %s. source state %s", name, from)
		panic(NewVizierError(ErrSourceManager, ErrMsgStateDoesNotExist, detail))
	}

	if m.states[from].HasEdge(name) {
		detail := fmt.Sprintf("failed to create output edge %s. source state %s", name, from)
		panic(NewVizierError(ErrSourceManager, ErrMsgEdgeAlreadyExists, detail))
	}

	output := make(chan Packet)
	err := m.states[from].AttachEdge(name, output)
	if err != nil {
		panic(err)
	}

	return output
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

func (m *Manager) Invoke(name string, payload interface{}) vizierErr {
	if _, ok := m.states[name]; !ok {
		detail := fmt.Sprintf("failed to invoke state %s.", name)
		return NewVizierError(ErrSourceManager, ErrMsgStateDoesNotExist, detail)
	}

	m.states[name].Invoke(payload)

	return nil
}

func (m *Manager) Start() vizierErr {
	if len(m.states) < 1 {
		return NewVizierError(ErrSourceManager, ErrMsgPoolEmptyStates, m.name)
	}

	if m.run {
		return NewVizierError(ErrSourceManager, ErrMsgPoolIsRunning, m.name)
	}

	m.run = true
	for i := 0; i < m.size; i++ {
		m.spawnWorker()
	}
	
	return nil
}

func (m *Manager) Stop() vizierErr {
	if !m.run {
		return NewVizierError(ErrSourceManager, ErrMsgPoolNotRunning, "")
	}
	m.run = false
	return nil
}

func (m *Manager) GetSize() int {
	return m.size
}

func (m *Manager) SetSize(size int) vizierErr {
	if !m.run {
		return NewVizierError(ErrSourceManager, ErrMsgPoolNotRunning, "")
	}

	if size <= 0 {
		detail := fmt.Sprintf("%s. invalid size %d", m.name, size)
		return NewVizierError(ErrSourceManager, ErrMsgPoolSizeInvalid, detail)
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
