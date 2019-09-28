package vizier

import (
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	CHANNEL_SIZE = 1000
)

type Manager struct {
	_      struct{}
	name   string
	states map[string]IState
	edges  map[string](chan Packet)
	Pool   *Pool
}

func (m *Manager) CreateState(name string, state IState) vizierErr {
	if _, ok := m.states[name]; ok {
		return NewVizierError(ErrSourceManager, ErrMsgStateAlreadyExists, name)
	}
	m.log(log.Fields{"state": name}).Info("created state")
	m.states[name] = state
	return nil
}

func (m *Manager) DeleteState(name string) vizierErr {
	if _, ok := m.states[name]; ok {
		m.log(log.Fields{"state": name}).Info("deleted state")
		delete(m.states, name)
		return nil
	}
	return NewVizierError(ErrSourceManager, ErrMsgStateDoesNotExist, name)
}

func (m *Manager) GetState(name string) (IState, vizierErr) {
	if s, ok := m.states[name]; ok {
		m.log(log.Fields{"state": name}).Info("get state")
		return s, nil
	}
	return nil, NewVizierError(ErrSourceManager, ErrMsgStateDoesNotExist, name)
}

func (m *Manager) CreateEdge(name string) (chan Packet, vizierErr) {
	if _, ok := m.edges[name]; ok {
		return nil, NewVizierError(ErrSourceManager, ErrMsgEdgeAlreadyExists, name)
	}
	m.log(log.Fields{"edge": name}).Info("created edge")
	edge := make(chan Packet, CHANNEL_SIZE)
	m.edges[name] = edge
	return edge, nil
}

func (m *Manager) DeleteEdge(name string) vizierErr {
	if _, ok := m.edges[name]; ok {
		m.log(log.Fields{"edge": name}).Info("delete edge")
		delete(m.edges, name)
		return nil
	}
	return NewVizierError(ErrSourceManager, ErrMsgEdgeDoesNotExist, name)
}

func (m *Manager) GetEdge(name string) (chan Packet, vizierErr) {
	if e, ok := m.edges[name]; ok {
		m.log(log.Fields{"edge": name}).Info("get edge")
		return e, nil
	}
	return nil, NewVizierError(ErrSourceManager, ErrMsgEdgeDoesNotExist, name)
}

func (m *Manager) detectCycle() vizierErr {
	return nil
}

func (m *Manager) log(fields log.Fields) *log.Entry {
	fields["source"] = "manager"
	fields["manager"] = m.name
	fields["time"] = time.Now().UTC().String()
	return log.WithFields(fields)
}

func NewManager(name string, poolSize int) (*Manager, error) {
	states := make(map[string]IState)
	pool, err := NewPool(name, poolSize, states)
	if err != nil {
		return nil, err
	}
	manager := Manager{
		name:   name,
		states: states,
		edges:  make(map[string](chan Packet)),
		Pool:   pool,
	}
	manager.log(log.Fields{}).Info("created manager")
	return &manager, nil
}
