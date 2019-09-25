package vizier

import (
	"fmt"
	"time"

	"github.com/google/uuid"
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
		if stream, ok := <-edge.recv; ok {
			defer func() {
				if err := recover(); err != nil {
					log.WithFields(log.Fields{
						"source":  "state",
						"state":   s.Name,
						"edge":    name,
						"trace":   stream.TraceID,
						"time":    time.Now().UTC().String(),
						"payload": stream.Payload,
					}).Warn("process failed")
				}
			}()
			log.WithFields(log.Fields{
				"source": "state",
				"state":  s.Name,
				"edge":   name,
				"trace":  stream.TraceID,
				"time":   time.Now().UTC().String(),
			}).Info("payload recieved")
			stream.Payload = s.Process(stream.Payload)
			edge.send <- stream
		}
	}
}

func (s *State) Send(name string, payload interface{}) vizierErr {
	if edge, ok := s.edges[name]; ok {
		traceID := uuid.New().String()
		log.WithFields(log.Fields{
			"source": "state",
			"state":  s.Name,
			"edge":   name,
			"trace":  traceID,
			"time":   time.Now().UTC().String(),
		}).Info("invoked state")
		edge.recv <- Stream{
			TraceID: traceID,
			Payload: payload,
		}
		return nil
	}
	detail := fmt.Sprintf("failed to send to state: %s on edge: %s", s.Name, name)
	return NewVizierError(ErrSourceState, ErrMsgEdgeDoesNotExist, detail)
}

func (s *State) AttachEdge(name string, recv chan Stream, send chan Stream) vizierErr {
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
