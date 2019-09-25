package vizier

type IEdge interface {}

type Edge struct {
	_    struct{}
	recv chan Stream
	send chan Stream
}