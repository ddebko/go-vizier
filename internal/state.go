package internal

import (
	"fmt"
	"sync"
	"time"

	"github.com/golang-collections/go-datastructures/queue"
	"github.com/google/uuid"

	log "github.com/sirupsen/logrus"
)

var (
	BUFFER_SIZE_WARNING int64       = 1000
	WARNING_INCREMENTS  int64       = 100
	CHANNEL_SIZE        int         = 1000
	STOP_STATE          interface{} = nil
)

type Packet struct {
	_         struct{}
	wg        *sync.WaitGroup
	TraceID   string
	Processed bool
	Payload   interface{}
}

type IState interface {
	Poll()
	Invoke(interface{}, *sync.WaitGroup)
	AttachEdge(string, chan Packet, bool) vizierErr
	HasEdge(string) bool
	GetPipe() chan Packet
}

type Edge struct {
	_        struct{}
	pipe     chan Packet
	isOutput bool
}

type State struct {
	_       struct{}
	Name    string
	Process func(interface{}) map[string]interface{}
	pipe    chan Packet
	edges   map[string]Edge
	buffers map[string]*queue.Queue
}

func (s State) Poll() {
	select {
	case packet := <-s.pipe:
		s.consumePacket(packet)
	default:
		s.consumeBuffers()
	}
}

func (s State) Invoke(payload interface{}, wg *sync.WaitGroup) {
	if wg != nil {
		wg.Add(1)
	}

	packet := Packet{
		wg:      wg,
		TraceID: uuid.New().String(),
		Payload: payload,
	}

	select {
	case s.pipe <- packet:
	default:
		s.buffers[s.Name].Put(packet)
	}
}

func (s State) AttachEdge(name string, pipe chan Packet, isOutput bool) vizierErr {
	if _, ok := s.edges[name]; ok {
		detail := fmt.Sprintf("failed to attach edge %s to state %s", name, s.Name)
		return NewVizierError(ErrSourceState, ErrMsgEdgeAlreadyExists, detail)
	}

	if pipe == nil {
		detail := fmt.Sprintf("channel Packet is nil for edge %s in state %s", name, s.Name)
		return NewVizierError(ErrSourceState, ErrNsgStateInvalidChan, detail)
	}

	s.edges[name] = Edge{
		pipe:     pipe,
		isOutput: isOutput,
	}
	s.buffers[name] = queue.New(1)

	return nil
}

func (s State) HasEdge(name string) bool {
	_, ok := s.edges[name]
	return ok
}

func (s State) GetPipe() chan Packet {
	return s.pipe
}

func (s State) consumeBuffers() {
	for edge, buffer := range s.buffers {
		if buffer.Len() > 0 {
			items, err := buffer.Get(1)
			if err != nil {
				s.log(log.Fields{
					"err":         err,
					"buffer_name": edge,
				}).Warn("failed to read from buffer")
				continue
			}

			if len(items) != 0 {
				if packet, ok := items[0].(Packet); ok {
					if packet.Processed {
						s.sendPacket(edge, packet)
						continue
					}

					s.consumePacket(packet)
				}
			}
		}
	}
}

func (s State) consumePacket(packet Packet) {
	for edge, payload := range s.Process(packet.Payload) {
		if _, ok := s.edges[edge]; ok {
			s.sendPacket(edge, Packet{
				wg:        packet.wg,
				TraceID:   packet.TraceID,
				Payload:   payload,
				Processed: true,
			})
			continue
		}
		s.log(log.Fields{
			"edge":    edge,
			"trace":   packet.TraceID,
			"payload": packet.Payload,
		}).Warn("edge not attached to state")
	}
}

func (s State) sendPacket(name string, packet Packet) {
	traceDetails := s.log(log.Fields{
		"edge":    name,
		"trace":   packet.TraceID,
		"payload": packet.Payload,
	})

	if packet.Payload == STOP_STATE {
		traceDetails.Info("packet returned STOP_STATE")
		return
	}

	select {
	case s.edges[name].pipe <- packet:
		if s.edges[name].isOutput {
			packet.wg.Done()
		}
		traceDetails.Info("packet sent to edge")
	default:
		s.buffers[name].Put(packet)
		traceDetails.Info("packet pushed to buffer")
	}
}

func (s State) log(fields log.Fields) *log.Entry {
	fields["source"] = "state"
	fields["state"] = s.Name
	fields["time"] = time.Now().UTC().String()
	return log.WithFields(fields)
}

func NewState(name string, process func(interface{}) map[string]interface{}) State {
	buffers := make(map[string]*queue.Queue)
	buffers[name] = queue.New(1)
	return State{
		Name:    name,
		Process: process,
		pipe:    make(chan Packet, CHANNEL_SIZE),
		edges:   make(map[string]Edge),
		buffers: buffers,
	}
}
