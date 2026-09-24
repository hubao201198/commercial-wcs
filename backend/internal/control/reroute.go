package control

import "errors"

var (
	ErrRerouteUnsafeState = errors.New("reroute requires a quiescent confirmed position")
	ErrReroutePositionMismatch = errors.New("reroute origin must match trusted confirmed position")
)

// SafeReroute replaces only the not-yet-dispatched remainder of a MOVE.
// It deliberately refuses to reinterpret an in-flight or UNKNOWN command:
// those states require physical reconciliation before any new route may be issued.
// The replacement path must begin at the controller's trusted confirmed node.
func SafeReroute(current *SegmentedMove, replacementPath []string) (*SegmentedMove, error) {
	if current == nil {
		return nil, ErrRerouteUnsafeState
	}
	if current.State() != MoveReady {
		return nil, ErrRerouteUnsafeState
	}
	confirmed := current.ConfirmedNode()
	if confirmed == "" || len(replacementPath) < 2 || replacementPath[0] != confirmed {
		return nil, ErrReroutePositionMismatch
	}

	replacement, err := NewSegmentedMove(replacementPath)
	if err != nil {
		return nil, err
	}
	return replacement, nil
}
