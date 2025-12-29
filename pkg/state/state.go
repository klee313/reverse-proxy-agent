// Package state defines agent lifecycle states and a state machine with transition rules.
// It is used by the agent to guard lifecycle transitions.

package state

import (
	"fmt"
	"sync"
)

type State int

const (
	StateStopped State = iota
	StateConnecting
	StateConnected
)

func (s State) String() string {
	switch s {
	case StateStopped:
		return "STOPPED"
	case StateConnecting:
		return "CONNECTING"
	case StateConnected:
		return "CONNECTED"
	default:
		return "UNKNOWN"
	}
}

type StateMachine struct {
	mu    sync.Mutex
	state State
}

func NewStateMachine() *StateMachine {
	return &StateMachine{state: StateStopped}
}

func (sm *StateMachine) State() State {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return sm.state
}

func (sm *StateMachine) Transition(next State) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if !allowedTransition(sm.state, next) {
		return fmt.Errorf("invalid transition: %s -> %s", sm.state, next)
	}

	sm.state = next
	return nil
}

func allowedTransition(from, to State) bool {
	switch from {
	case StateStopped:
		return to == StateConnecting || to == StateStopped
	case StateConnecting:
		return to == StateConnected || to == StateStopped
	case StateConnected:
		return to == StateConnecting || to == StateStopped || to == StateConnected
	default:
		return false
	}
}
