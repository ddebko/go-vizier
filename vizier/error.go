package vizier

import "fmt"

type vizierCode string
type vizierMessage string

const (
	ErrCodeManager vizierCode = "manager"
	ErrCodeState   vizierCode = "state"
	ErrCodeEdge    vizierCode = "edge"
	ErrCodePool    vizierCode = "pool"
	ErrCodeWorker  vizierCode = "worker"
	ErrMsgStateAlreadyExists vizierMessage = "state already exists"
	ErrMsgStateDoesNotExist  vizierMessage = "state does not exist"
	ErrMsgEdgeAlreadyExists  vizierMessage = "edge already exists"
	ErrMsgEdgeDoesNotExist   vizierMessage = "edge does not exist"
	ErrMsgPoolNotRunning     vizierMessage = "pool is not running"
	ErrMsgPoolSizeInvalid    vizierMessage = "pool size must be greater than 0"
	ErrMsgPoolEmptyStates    vizierMessage = "pool requires at least one state"
)

type Error interface {
	error
	Code() string
	Message() string
	Details() string
	Err() error
}

type vizierErr Error

type VizierError struct {
	vizierErr
	_       struct{}
	code    vizierCode
	message vizierMessage
	details string
}

func (v *VizierError) Code() string {
	return string(v.code)
}

func (v *VizierError) Message() string {
	return string(v.message)
}

func (v *VizierError) Details() string {
	return v.details
}

func (v *VizierError) Err() error {
	return fmt.Errorf("[VIZIER] code: %s. message: %s. details: %s.", v.code, v.message, v.details)
}

func NewVizierError(code vizierCode, message vizierMessage, details string) *VizierError {
	return &VizierError{
		code:    code,
		message: message,
		details: details,
	}
}
