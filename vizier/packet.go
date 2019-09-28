package vizier

type Packet struct {
	_         struct{}
	TraceID   string
	Processed bool
	Payload   interface{}
}
