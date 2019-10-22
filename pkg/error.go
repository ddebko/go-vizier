package internal

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

type source string
type message string

// Common Defintions For VizierError Source & Message
const (
	ErrSourceManager         source  = "manager"
	ErrSourceState           source  = "state"
	ErrSourceEdge            source  = "edge"
	ErrSourceWorker          source  = "worker"
	ErrMsgManagerCycle       message = "cycle detected in graph"
	ErrMsgStateAlreadyExists message = "state already exists"
	ErrMsgStateDoesNotExist  message = "state does not exist"
	ErrNsgStateInvalidChan   message = "channel cannot be nil"
	ErrMsgEdgeAlreadyExists  message = "edge already exists"
	ErrMsgEdgeDoesNotExist   message = "edge does not exist"
	ErrMsgPoolNotRunning     message = "pool is not running"
	ErrMsgPoolIsRunning      message = "pool is already running"
	ErrMsgPoolSizeInvalid    message = "pool size must be greater than 0"
	ErrMsgPoolEmptyStates    message = "pool requires at least one state"
)

// Error is a interface that provides different levels of information.
// Source() provides information on where the error had occured: manager, state, edge, worker
// Message() provides information on the root cause of the error
// Details() provides additional information for debugging purposes
// Err() concats information from source, message, & details into a type error
type Error interface {
	error
	Source() string
	Message() string
	Details() string
	Err() error
}

type vizierErr Error

// VizierError implements the interace Error
type VizierError struct {
	vizierErr
	_       struct{}
	src     source
	msg     message
	details string
}

// Source casts the field src into a string & returns the value
func (v *VizierError) Source() string {
	return string(v.src)
}

// Message casts the field msg into a string & returns the value
func (v *VizierError) Message() string {
	return string(v.msg)
}

// Details returns the field details
func (v *VizierError) Details() string {
	return v.details
}

// Err returns the fields src, msg, & details into an error\
// format [VIZIER] source: %s. message: %s. details: %s.
func (v *VizierError) Err() error {
	return fmt.Errorf("[VIZIER] source: %s. message: %s. details: %s.", v.src, v.msg, v.details)
}

func newVizierError(src source, msg message, details string) *VizierError {
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
