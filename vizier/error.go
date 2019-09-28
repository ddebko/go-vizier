package vizier

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

type source string
type message string

const (
	ErrSourceManager         source  = "manager"
	ErrSourceState           source  = "state"
	ErrSourceEdge            source  = "edge"
	ErrSourcePool            source  = "pool"
	ErrSourceWorker          source  = "worker"
	ErrMsgStateAlreadyExists message = "state already exists"
	ErrMsgStateDoesNotExist  message = "state does not exist"
	ErrMsgStateInvalidEdge   message = "edge cannot be nil"
	ErrMsgEdgeAlreadyExists  message = "edge already exists"
	ErrMsgEdgeDoesNotExist   message = "edge does not exist"
	ErrMsgPoolNotRunning     message = "pool is not running"
	ErrMsgPoolSizeInvalid    message = "pool size must be greater than 0"
	ErrMsgPoolEmptyStates    message = "pool requires at least one state"
)

type Error interface {
	error
	Source() string
	Message() string
	Details() string
	Err() error
}

type vizierErr Error

type VizierError struct {
	vizierErr
	_       struct{}
	src     source
	msg     message
	details string
}

func (v *VizierError) Source() string {
	return string(v.src)
}

func (v *VizierError) Message() string {
	return string(v.msg)
}

func (v *VizierError) Details() string {
	return v.details
}

func (v *VizierError) Err() error {
	return fmt.Errorf("[VIZIER] code: %s. message: %s. details: %s.", v.src, v.msg, v.details)
}

func NewVizierError(src source, msg message, details string) *VizierError {
	log.WithFields(log.Fields{
		"source":  src,
		"details": details,
	}).Warn(msg)
	return &VizierError{
		src:     src,
		msg:     msg,
		details: details,
	}
}
