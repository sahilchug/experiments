package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

type State int

const (
	Closed State = iota
	HalfOpen
	Open
)

var ErrCircuitOpen = errors.New("circuit is open")

type CircuitBreaker struct {
	state            State
	successCount     int
	failureCount     int
	lastFailureTime  time.Time
	failureThreshold int           // Number of failures to trigger the circuit to open
	timeout          time.Duration // Duration to wait before transitioning from Open to Half-Open.
	mutex            sync.Mutex
}

func NewCircuitBreaker(failureThreshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:            Closed,
		failureThreshold: failureThreshold,
		timeout:          timeout,
	}
}

func (cb *CircuitBreaker) Execute(req func() (interface{}, error)) (interface{}, error) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	// Check the current state.
	switch cb.state {
	case Open:
		// If timeout has passed, allow a test call in Half-Open state.
		if time.Since(cb.lastFailureTime) > cb.timeout {
			cb.state = HalfOpen
		} else {
			return nil, ErrCircuitOpen
		}
	default:
		// Closed or Half-Open, continue.
	}

	// Execute the request.
	resp, err := req()
	if err != nil {
		cb.failureCount += 1
		// If failure count exceeds the threshold, trip the breaker.
		if cb.failureCount >= cb.failureThreshold {
			cb.trip()
		}

		return nil, err
	}

	// On success:
	// If we're in Half-Open, a successful call can reset the breaker.
	if cb.state == HalfOpen {
		cb.reset()
	} else if cb.state == Closed {
		// Optionally, you could reset the failure count on success.
		cb.failureCount = 0
	}

	return resp, nil
}

// trip transitions the breaker to the Open state.
func (cb *CircuitBreaker) trip() {
	cb.state = Open
	cb.lastFailureTime = time.Now()
	// Optionally, log or perform additional actions here.
}

// reset transitions the breaker back to the Closed state.
func (cb *CircuitBreaker) reset() {
	cb.state = Closed
	cb.failureCount = 0
	cb.successCount = 0
	// Optionally, log or perform additional actions here.
}
