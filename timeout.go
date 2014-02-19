package mongoproxy

import (
	"errors"
	"time"
)

var errTimeout error

func init() {
	errTimeout = errors.New("operation timed out")
}

type response struct {
	value interface{}
	err   error
}

// Runs the block and waits at most timeout for the block to finish.  If
// the block does not finish in time, returns a timeout error, which can
// be checked by IsTimeout(err).
func TimeoutIn(block func() (interface{}, error), timeout time.Duration) (
	interface{}, error) {
	ch := make(chan *response, 1)
	go func() {
		value, err := block()
		ch <- &response{value: value, err: err}
	}()
	timer := time.After(timeout)
	select {
	case value := <-ch:
		return value.value, value.err
	case <-timer:
		return nil, errTimeout
	}
}

// Returns whether or not this error is a timeout.
func IsTimeout(err error) bool {
	return err == errTimeout
}
