package internal

import (
	"fmt"
	"math"
	"sync"
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
		panic(newVizierError(ErrSourceManager, ErrMsgStateAlreadyExists, detail))
	}

	m.states[name] = newState(name, f)

	return m
}

func (m *Manager) Output(from, name string) chan Packet {
	if _, ok := m.states[from]; !ok {
		detail := fmt.Sprintf("failed to create output edge %s. source state %s", name, from)
		panic(newVizierError(ErrSourceManager, ErrMsgStateDoesNotExist, detail))
	}

	if m.states[from].HasEdge(name) {
		detail := fmt.Sprintf("failed to create output edge %s. source state %s", name, from)
		panic(newVizierError(ErrSourceManager, ErrMsgEdgeAlreadyExists, detail))
	}

	output := make(chan Packet)
	err := m.states[from].AttachEdge(name, output, true)
	if err != nil {
		panic(err)
	}

	return output
}

func (m *Manager) Edge(from, to, name string) *Manager {
	edgeName := fmt.Sprintf("%s_to_%s_%s", from, to, name)

	if _, ok := m.states[to]; !ok {
		detail := fmt.Sprintf("failed to create edge %s. destination state %s", name, to)
		panic(newVizierError(ErrSourceManager, ErrMsgStateDoesNotExist, detail))
	}

	if _, ok := m.states[from]; !ok {
		detail := fmt.Sprintf("failed to create edge %s. source state %s", name, from)
		panic(newVizierError(ErrSourceManager, ErrMsgStateDoesNotExist, detail))
	}

	if m.states[from].HasEdge(edgeName) {
		detail := fmt.Sprintf("failed to create edge %s. source state %s", name, from)
		panic(newVizierError(ErrSourceManager, ErrMsgEdgeAlreadyExists, detail))
	}

	err := m.states[from].AttachEdge(edgeName, m.states[to].GetPipe(), false)
	if err != nil {
		panic(err)
	}

	return m
}

func (m *Manager) BatchInvoke(name string, batch []interface{}) (*sync.WaitGroup, vizierErr) {
	var wg sync.WaitGroup

	if _, ok := m.states[name]; !ok {
		detail := fmt.Sprintf("failed to invoke state %s.", name)
		return nil, newVizierError(ErrSourceManager, ErrMsgStateDoesNotExist, detail)
	}

	for _, payload := range batch {
		m.states[name].Invoke(payload, &wg)
	}

	return &wg, nil
}

func (m *Manager) Invoke(name string, payload interface{}) (*sync.WaitGroup, vizierErr) {
	var wg sync.WaitGroup

	if _, ok := m.states[name]; !ok {
		detail := fmt.Sprintf("failed to invoke state %s.", name)
		return nil, newVizierError(ErrSourceManager, ErrMsgStateDoesNotExist, detail)
	}

	m.states[name].Invoke(payload, &wg)

	return &wg, nil
}

func (m *Manager) Start() vizierErr {
	if len(m.states) < 1 {
		return newVizierError(ErrSourceManager, ErrMsgPoolEmptyStates, m.name)
	}

	if m.run {
		return newVizierError(ErrSourceManager, ErrMsgPoolIsRunning, m.name)
	}

	m.run = true
	for i := 0; i < m.size; i++ {
		m.spawnWorker()
	}

	return nil
}

func (m *Manager) Stop() vizierErr {
	if !m.run {
		return newVizierError(ErrSourceManager, ErrMsgPoolNotRunning, "")
	}
	m.run = false
	return nil
}

func (m *Manager) GetSize() int {
	return m.size
}

func (m *Manager) SetSize(size int) vizierErr {
	if !m.run {
		return newVizierError(ErrSourceManager, ErrMsgPoolNotRunning, "")
	}

	if size <= 0 {
		detail := fmt.Sprintf("%s. invalid size %d", m.name, size)
		return newVizierError(ErrSourceManager, ErrMsgPoolSizeInvalid, detail)
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

func (m *Manager) GetResults(wg *sync.WaitGroup, size int, pipe chan Packet) []interface{} {
	results := make([]interface{}, size)
	isWaiting := true
	go func() {
		index := 0
		for isWaiting || index < size {
			packet := <-pipe
			results[index] = packet.Payload
			index++
		}
	}()
	wg.Wait()
	isWaiting = false
	return results
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
