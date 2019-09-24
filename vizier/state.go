package vizier

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

type IState interface {
	Run()
	Send(string, interface{})
}

type State struct {
	_       struct{}
	Name    string
	Process func(interface{}) interface{}
	edges   map[string]Edge
}

func (s *State) Run() {
	for name, edge := range s.edges {
		if payload, ok := <-edge.recv; ok {
			log.WithFields(log.Fields{
				"source": "state",
				"state":  s.Name,
				"edge":   name,
			}).Info("payload recieved")
			edge.send <- s.Process(payload)
		}
	}
}

func (s *State) Send(name string, payload interface{}) vizierErr {
	if edge, ok := s.edges[name]; ok {
		log.WithFields(log.Fields{
			"source": "state",
			"state":  s.Name,
			"edge":   name,
		}).Info("invoked state")
		edge.recv <- payload
		return nil
	}
	detail := fmt.Sprintf("failed to send to state: %s on edge: %s", s.Name, name)
	return NewVizierError(ErrSourceState, ErrMsgEdgeDoesNotExist, detail)
}

func (s *State) AttachEdge(name string, recv chan interface{}, send chan interface{}) vizierErr {
	if _, ok := s.edges[name]; ok {
		detail := fmt.Sprintf("failed to attach edge %s to state %s", name, s.Name)
		return NewVizierError(ErrSourceState, ErrMsgEdgeAlreadyExists, detail)
	}
	log.WithFields(log.Fields{
		"source": "state",
		"state":  s.Name,
		"edge":   name,
	}).Info("attached edge")
	s.edges[name] = Edge{
		recv: recv,
		send: send,
	}
	return nil
}

func (s *State) DetachEdge(name string) vizierErr {
	if _, ok := s.edges[name]; !ok {
		detail := fmt.Sprintf("failed to detach edge %s from state %s", name, s.Name)
		return NewVizierError(ErrSourceState, ErrMsgEdgeDoesNotExist, detail)
	}
	log.WithFields(log.Fields{
		"source": "state",
		"state":  s.Name,
		"edge":   name,
	}).Info("detached edge")
	delete(s.edges, name)
	return nil
}

func NewState(name string, process func(interface{}) interface{}) State {
	return State{
		Name:    name,
		Process: process,
		edges:   make(map[string]Edge),
	}
}
