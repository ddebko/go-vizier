package vizier

import (
	"fmt"
)

// TODO: Create Custom Error Type
// TODO: Create Logger
// TODO: Worker Pool Observer To Record Statistics

type Manager struct {
	_      struct{}
	states map[string]IState
	edges  map[string](chan interface{})
	Pool   *Pool
}

func (m *Manager) CreateState(name string, state IState) error {
	if _, ok := m.states[name]; ok {
		return fmt.Errorf("state name already exists: %s", name)
	}

	m.states[name] = state
	return nil
}

func (m *Manager) CreateEdge(name string) error {
	if _, ok := m.edges[name]; ok {
		return fmt.Errorf("edge name already exists: %s", name)
	}

	m.edges[name] = make(chan interface{})
	return nil
}

func (m *Manager) GetEdge(name string) (chan interface{}, error) {
	if e, ok := m.edges[name]; ok {
		return e, nil
	}
	return nil, fmt.Errorf("edge does not exist: %s", name)
}

func NewManager(poolSize int) (*Manager, error) {
	states := make(map[string]IState)
	pool, err := NewPool(poolSize, states)
	if err != nil {
		return nil, err
	}
	return &Manager{
		states: states,
		edges:  make(map[string](chan interface{})),
		Pool:   pool,
	}, nil
}
