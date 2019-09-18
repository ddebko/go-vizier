package vizier

type IState interface {
	Run()
	Send(interface{})
}

type State struct {
	_       struct{}
	Process func(interface{}) interface{}
	recv    chan interface{}
	send    chan interface{}
}

func (s *State) Run() {
	if payload, ok := <-s.recv; ok {
		s.send <- s.Process(payload)
	}
}

func (s *State) Send(payload interface{}) {
	s.recv <- payload
}
