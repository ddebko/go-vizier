package internal

import (
	"fmt"
)

type Manager struct {
	_      struct{}
	name   string
	states map[string]IState
	size   int
	run    bool
}

func (m *Manager) Node(name string, f func(interface{}) map[string]interface{}) *Manager {
	if _, ok := m.states[name]; ok {
		panic("")
	}

	m.states[name] = NewState(name, f)

	return m
}

func (m *Manager) Edge(from, to, name string) *Manager {
	edgeName := fmt.Sprintf("%s_to_%s_%s", from, to, name)

	if _, ok := m.states[to]; !ok {
		details := fmt.Sprintf("failed to create edge. destination state %s", to)
		panic(details)
	}

	if _, ok := m.states[from]; !ok {
		details := fmt.Sprintf("failed to create edge. source state %s", from)
		panic(details)
	}

	if m.states[from].HasEdge(edgeName) {
		panic("")
	}

	m.states[from].AttachEdge(edgeName, m.states[to].GetPipe())

	return m
}

func (m *Manager) Invoke(name string, payload interface{}) error {
	if _, ok := m.states[name]; !ok {
		return nil
	}

	m.states[name].Invoke(payload)

	return nil
}

func NewManager(name string, size int) (*Manager, error) {
	states := make(map[string]IState)

	manager := Manager{
		name:   name,
		states: states,
		size:   size,
	}

	return &manager, nil
}

// TODO: Merge Pool.go Into Manager
