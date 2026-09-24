package control

import (
	"errors"
	"sync"
)

var (
	ErrStaleControlGeneration = errors.New("stale control generation")
	ErrTakeoverUnsafeState    = errors.New("takeover requires quiescent trusted physical state")
)

// ControlLease fences command authority across controller failover. Every command
// producer must hold the current monotonically increasing generation. A promoted
// controller advances the generation; commands from the old controller then fail
// closed instead of racing the new owner.
type ControlLease struct {
	mu         sync.Mutex
	generation uint64
}

func NewControlLease(initial uint64) *ControlLease {
	if initial == 0 {
		initial = 1
	}
	return &ControlLease{generation: initial}
}

func (l *ControlLease) Generation() uint64 {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.generation
}

func (l *ControlLease) Authorize(generation uint64) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if generation != l.generation {
		return ErrStaleControlGeneration
	}
	return nil
}

// Promote transfers command authority to a new controller generation. Promotion
// is permitted only when physical execution is quiescent and the controller has
// a trusted confirmed node. IN_FLIGHT/UNKNOWN execution must be reconciled first.
func (l *ControlLease) Promote(move *SegmentedMove) (uint64, error) {
	if move == nil || move.State() != MoveReady || move.ConfirmedNode() == "" {
		return 0, ErrTakeoverUnsafeState
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.generation++
	return l.generation, nil
}

// DispatchNext authorizes the generation before issuing a physical segment.
// This keeps stale controllers from dispatching after failover.
func (l *ControlLease) DispatchNext(move *SegmentedMove, generation uint64) (Segment, error) {
	if err := l.Authorize(generation); err != nil {
		return Segment{}, err
	}
	if move == nil {
		return Segment{}, ErrTakeoverUnsafeState
	}
	return move.DispatchNext()
}
