package vizier

import "fmt"

type IState interface {
	Run()
	Send(interface{})
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

func (s *State) Send(name string, payload interface{}) error {
	if edge, ok := s.edges[name]; ok {
		edge.recv <- payload
		return nil
	}
	return fmt.Errorf("edge %s not found in state %s", name, s.Name)
}

func (s *State) AttachEdge(name string, from chan interface{}, to chan interface{}) error {
	if _, ok := s.edges[name]; ok {
		return fmt.Errorf("edge %s already exists in state %s", name, s.Name)
	}
	s.edges[name] = Edge{
		recv: from,
		send: to,
	}
	return nil
}

func (s *State) DetachEdge(name string) error {
	if _, ok := s.edges[name]; !ok {
		return fmt.Errorf("edge %s does not exist in state %s", name, s.Name)
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
