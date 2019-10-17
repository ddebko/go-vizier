package vizier

import (
	"fmt"
	"time"

	"github.com/golang-collections/go-datastructures/queue"
	log "github.com/sirupsen/logrus"
)

var (
	CHANNEL_SIZE = 1000
)

type Manager struct {
	_      struct{}
	name   string
	states map[string]IState
	Pool   *Pool
}

func (m *Manager) Node(name string, f func(interface{}) interface{}) *Manager {
	if _, ok := m.states[name]; ok {
		panic(NewVizierError(ErrSourceManager, ErrMsgStateAlreadyExists, name))
	}

	m.log(log.Fields{"state": name}).Info("created state")

	m.states[name] = State{
		Name:    name,
		Process: f,
		edges:   make(map[string]Edge),
		buffers: make(map[string]*queue.Queue),
	}

	return m
}

func (m *Manager) Edge(from, to, name string) *Manager {
	edgeName := fmt.Sprintf("%s_to_%s_%s", from, to, name)

	if _, ok := m.states[to]; !ok {
		details := fmt.Sprintf("failed to create edge. destination state %s", to)
		panic(NewVizierError(ErrSourceManager, ErrMsgStateDoesNotExist, details))
	}

	if _, ok := m.states[from]; !ok {
		details := fmt.Sprintf("failed to create edge. source state %s", from)
		panic(NewVizierError(ErrSourceManager, ErrMsgStateDoesNotExist, details))
	}

	if m.states[from].HasEdge(edgeName) {
		panic(NewVizierError(ErrSourceManager, ErrMsgEdgeAlreadyExists, edgeName))
	}

	m.log(log.Fields{"edge": edgeName}).Info("created edge")

	edge := Edge{
		recv: make(chan Packet, CHANNEL_SIZE),
		send: make(chan Packet, CHANNEL_SIZE),
	}

	m.states[from].AttachEdge(edgeName, edge)

	return m
}

func (m *Manager) Invoke(state, edge string, payload interface{}) vizierErr {
	if state, ok := m.states[state]; !ok {
		if !state.HasEdge(edge) {
			details := fmt.Sprintf("failed to invoke state %s. edge %s", state, edge)
			panic(NewVizierError(ErrSourceManager, ErrMsgEdgeDoesNotExist, details))
		}

		details := fmt.Sprintf("failed to invoke state. %s", state)
		panic(NewVizierError(ErrSourceManager, ErrMsgStateDoesNotExist, details))
	}

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
		Pool:   pool,
	}

	manager.log(log.Fields{}).Info("created manager")

	return &manager, nil
}
