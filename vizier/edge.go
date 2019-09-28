package vizier

type IEdge interface{}

type Edge struct {
	_    struct{}
	recv chan Packet
	send chan Packet
}
