package vizier

type Stream struct {
	_       struct{}
	TraceID string
	Payload interface{}
}