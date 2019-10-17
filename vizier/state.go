package vizier

import (
	"fmt"
	"time"

	"github.com/golang-collections/go-datastructures/queue"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

var (
	BUFFER_SIZE_WARNING int64       = 1000
	WARNING_INCREMENTS  int64       = 100
	STOP_STATE          interface{} = nil
)

type IState interface {
	Poll()
	Invoke(string, interface{}) vizierErr
	AttachEdge(string, chan Packet) vizierErr
	DetachEdge(string) vizierErr
}

type State struct {
	_       struct{}
	Name    string
	Process func(interface{}) interface{}
	Pipe    chan Packet
	edges   map[string](chan Packet)
	buffers map[string]*queue.Queue
}

func (s State) Poll() {
	select {
	case stream := <-s.pipe:
		fields["trace"] = stream.TraceID
		defer func() {
			if err := recover(); err != nil {
				fields["payload"] = stream.Payload
				fields["err"] = err
				s.log(fields).Warn("process failed")
			}
		}()
		s.log(fields).Info("payload recieved")
		s.send(name, Packet{
			TraceID:   stream.TraceID,
			Payload:   s.Process(stream.Payload),
			Processed: true,
		})
	default:
		err := s.consumeBuffer(name)
		if err != nil {
			fields["err"] = err
			s.log(fields).Warn("consuming buffer")
		}
	}
}

func (s State) Invoke(name string, payload interface{}) vizierErr {
	if edge, ok := s.edges[name]; ok {
		stream := Packet{
			TraceID: uuid.New().String(),
			Payload: payload,
		}
		s.log(log.Fields{
			"edge":  name,
			"trace": stream.TraceID,
		}).Info("invoked state")
		select {
		case edge.recv <- stream:
		default:
			s.buffers[name].Put(stream)
			bufferSize := s.buffers[name].Len()
			if bufferSize > BUFFER_SIZE_WARNING && bufferSize%WARNING_INCREMENTS == 0 {
				s.log(log.Fields{
					"edge": name,
					"size": bufferSize,
				}).Warn("buffer size")
			}
		}
		return nil
	}
	detail := fmt.Sprintf("failed to send to state: %s on edge: %s", s.Name, name)
	return NewVizierError(ErrSourceState, ErrMsgEdgeDoesNotExist, detail)
}

func (s State) AttachEdge(name string, recv chan Packet, send chan Packet) vizierErr {
	if _, ok := s.edges[name]; ok {
		detail := fmt.Sprintf("failed to attach edge %s to state %s", name, s.Name)
		return NewVizierError(ErrSourceState, ErrMsgEdgeAlreadyExists, detail)
	}

	if recv == nil {
		return NewVizierError(ErrSourceState, ErrMsgStateInvalidEdge, "channel recv cannot be nil")
	}

	if send == nil {
		return NewVizierError(ErrSourceState, ErrMsgStateInvalidEdge, "channel send cannot be nil")
	}

	s.log(log.Fields{"edge": name}).Info("attached edge")
	s.edges[name] = Edge{
		recv: recv,
		send: send,
	}
	s.buffers[name] = queue.New(1)
	return nil
}

func (s State) DetachEdge(name string) vizierErr {
	if _, ok := s.edges[name]; !ok {
		detail := fmt.Sprintf("failed to detach edge %s from state %s", name, s.Name)
		return NewVizierError(ErrSourceState, ErrMsgEdgeDoesNotExist, detail)
	}
	s.log(log.Fields{"edge": name}).Info("detached edge")
	delete(s.edges, name)
	delete(s.buffers, name)
	return nil
}

func (s State) consumeBuffer(name string) error {
	if s.buffers[name].Len() > 0 {
		items, err := s.buffers[name].Get(1)
		if err != nil {
			return err
		}

		if len(items) != 0 {
			if stream, ok := items[0].(Packet); ok {
				if stream.Processed {
					s.send(name, stream)
				}

				s.send(name, Packet{
					TraceID:   stream.TraceID,
					Payload:   s.Process(stream.Payload),
					Processed: true,
				})
			}
		}
	}
	return nil
}

func (s State) send(name string, stream Packet) {
	if stream.Payload == STOP_STATE {
		return
	}

	select {
	case s.edges[name].send <- stream:
	default:
		s.buffers[name].Put(stream)
	}
}

func (s State) log(fields log.Fields) *log.Entry {
	fields["source"] = "state"
	fields["state"] = s.Name
	fields["time"] = time.Now().UTC().String()
	return log.WithFields(fields)
}

func NewState(name string, process func(interface{}) interface{}) State {
	return State{
		Name:    name,
		Process: process,
		Pipe:    make(chan Packet, 1000),
		edges:   make(map[string](chan Packet)),
		buffers: make(map[string]*queue.Queue),
	}
}
