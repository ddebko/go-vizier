package vizier

type IEdge interface {}

type Edge struct {
	_    struct{}
	recv chan interface{}
	send chan interface{}
}