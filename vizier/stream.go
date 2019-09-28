package vizier

type Stream struct {
	_         struct{}
	TraceID   string
	Processed bool
	Payload   interface{}
}
