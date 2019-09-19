package vizier

type Manager struct {
	_      struct{}
	name   string
	states map[string]IState
	edges  map[string](chan interface{})
	Pool   *Pool
}

func (m *Manager) CreateState(name string, state IState) vizierErr {
	if _, ok := m.states[name]; ok {
		return NewVizierError(ErrCodeManager, ErrMsgStateAlreadyExists, name)
	}
	m.states[name] = state
	return nil
}

func (m *Manager) DeleteState(name string) vizierErr {
	if _, ok := m.states[name]; ok {
		delete(m.states, name)
		return nil
	}
	return NewVizierError(ErrCodeManager, ErrMsgStateDoesNotExist, name)
}

func (m *Manager) GetState(name string) (IState, vizierErr) {
	if s, ok := m.states[name]; ok {
		return s, nil
	}
	return nil, NewVizierError(ErrCodeManager, ErrMsgStateDoesNotExist, name)
}

func (m *Manager) CreateEdge(name string) vizierErr {
	if _, ok := m.edges[name]; ok {
		return NewVizierError(ErrCodeManager, ErrMsgEdgeAlreadyExists, name)
	}
	m.edges[name] = make(chan interface{})
	return nil
}

func (m *Manager) DeleteEdge(name string) vizierErr {
	if _, ok := m.edges[name]; ok {
		delete(m.edges, name)
		return nil
	}
	return NewVizierError(ErrCodeManager, ErrMsgEdgeDoesNotExist, name)
}

func (m *Manager) GetEdge(name string) (chan interface{}, vizierErr) {
	if e, ok := m.edges[name]; ok {
		return e, nil
	}
	return nil, NewVizierError(ErrCodeManager, ErrMsgEdgeDoesNotExist, name)
}

func NewManager(name string, poolSize int) (*Manager, error) {
	states := make(map[string]IState)
	pool, err := NewPool(name, poolSize, states)
	if err != nil {
		return nil, err
	}
	return &Manager{
		name:   name,
		states: states,
		edges:  make(map[string](chan interface{})),
		Pool:   pool,
	}, nil
}
