package pkg

import (
	"fmt"
	"sync"
	"time"

	"github.com/golang-collections/go-datastructures/queue"
	"github.com/google/uuid"

	log "github.com/sirupsen/logrus"
)

//
var (
	BufferSizeWarning = 1000
	WarningIncrements = 100
	ChannelSize       = 1000
	StopState         interface{}
)

// Packet is sent between States through edges or output edges & contain the payloads that are created from the previous State's task function.
type Packet struct {
	_         struct{}
	wg        *sync.WaitGroup
	TraceID   string
	Processed bool
	Payload   interface{}
}

// IState is a interface that provides the basic requirements for Nodes to function together in the manager graph.
type IState interface {
	Poll()
	Invoke(interface{}, *sync.WaitGroup)
	AttachEdge(string, chan Packet, bool) VizierErr
	HasEdge(string) bool
	GetPipe() chan Packet
}

// Edge uses channels to either connect two States or is an output that allows information to be extracted out of the graph.
type Edge struct {
	_        struct{}
	pipe     chan Packet
	isOutput bool
}

// State implements the interface IState
type State struct {
	_       struct{}
	Name    string
	Process func(interface{}) map[string]interface{}
	pipe    chan Packet
	edges   map[string]Edge
	buffers map[string]*queue.Queue
}

// Poll either tries to listen for an incoming packet from State in the graph or consumes from a buffer of Packets waiting in a queue.
func (s State) Poll() {
	select {
	case packet := <-s.pipe:
		s.consumePacket(packet)
	default:
		s.consumeBuffers()
	}
}

// Invoke creates a new Packet & either sends the Packet to the State's channel or buffer.
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
		err := s.buffers[s.Name].Put(packet)
		if err != nil {
			details := fmt.Sprintf("packet: %+v, state %s, err %s", packet, s.Name, err.Error())
			fmt.Println(newVizierError(ErrSourceState, ErrMsgStateBufferErr, details).Err())
		}
	}
}

// AttachEdge creates a new edge that connects to a different State in the graph or creates an output edge
func (s State) AttachEdge(name string, pipe chan Packet, isOutput bool) VizierErr {
	if _, ok := s.edges[name]; ok {
		detail := fmt.Sprintf("failed to attach edge %s to state %s", name, s.Name)
		return newVizierError(ErrSourceState, ErrMsgEdgeAlreadyExists, detail)
	}

	if pipe == nil {
		detail := fmt.Sprintf("channel Packet is nil for edge %s in state %s", name, s.Name)
		return newVizierError(ErrSourceState, ErrMsgStateInvalidChan, detail)
	}

	s.edges[name] = Edge{
		pipe:     pipe,
		isOutput: isOutput,
	}
	s.buffers[name] = queue.New(1)

	return nil
}

// HasEdge returns true if the State contains a edge 'name'
func (s State) HasEdge(name string) bool {
	_, ok := s.edges[name]
	return ok
}

// GetPipe returns the channel Packet associated to the State
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

	if packet.Payload == StopState {
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
		err := s.buffers[name].Put(packet)
		if err != nil {
			details := fmt.Sprintf("packet: %+v, state %s, edge %s, error %s", packet, s.Name, name, err.Error())
			fmt.Println(newVizierError(ErrSourceState, ErrMsgStateBufferErr, details).Err())
		}
		traceDetails.Info("packet pushed to buffer")
	}
}

func (s State) log(fields log.Fields) *log.Entry {
	fields["source"] = "state"
	fields["state"] = s.Name
	fields["time"] = time.Now().UTC().String()
	return log.WithFields(fields)
}

func newState(name string, process func(interface{}) map[string]interface{}) State {
	buffers := make(map[string]*queue.Queue)
	buffers[name] = queue.New(1)
	return State{
		Name:    name,
		Process: process,
		pipe:    make(chan Packet, ChannelSize),
		edges:   make(map[string]Edge),
		buffers: buffers,
	}
}
