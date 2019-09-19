package vizier

import "fmt"

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
	for _, edge := range s.edges {
		if payload, ok := <-edge.recv; ok {
			edge.send <- s.Process(payload)
		}
	}
}

func (s *State) Send(name string, payload interface{}) vizierErr {
	if edge, ok := s.edges[name]; ok {
		edge.recv <- payload
		return nil
	}
	detail := fmt.Sprintf("failed to send to state: %s on edge: %s", s.Name, name)
	return NewVizierError(ErrCodeState, ErrMsgEdgeDoesNotExist, detail)
}

func (s *State) AttachEdge(name string, from chan interface{}, to chan interface{}) vizierErr {
	if _, ok := s.edges[name]; ok {
		detail := fmt.Sprintf("failed to attach edge %s to state %s", name, s.Name)
		return NewVizierError(ErrCodeState, ErrMsgEdgeAlreadyExists, detail)
	}
	s.edges[name] = Edge{
		recv: from,
		send: to,
	}
	return nil
}

func (s *State) DetachEdge(name string) vizierErr {
	if _, ok := s.edges[name]; !ok {
		detail := fmt.Sprintf("failed to detach edge %s from state %s", name, s.Name)
		return NewVizierError(ErrCodeState, ErrMsgEdgeDoesNotExist, detail)
	}
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
