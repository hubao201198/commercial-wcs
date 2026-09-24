package control

import (
	"errors"
	"sync"
)

var (
	ErrMoveAwaitingConfirmation = errors.New("move segment awaiting node confirmation")
	ErrMoveUnknown              = errors.New("move segment outcome unknown")
	ErrUntrustedPosition        = errors.New("trusted physical position required")
)

type MoveState string

const (
	MoveReady    MoveState = "READY"
	MoveInFlight MoveState = "IN_FLIGHT"
	MoveUnknown  MoveState = "UNKNOWN"
	MoveDone     MoveState = "DONE"
)

type MoveSegment struct {
	From string
	To   string
}

// SegmentedMove coordinates physical movement one confirmed segment at a time.
// It deliberately separates dispatch from node confirmation: a successful send
// never proves physical arrival. UNKNOWN outcomes stop progression and are not
// replayable without an external recovery decision based on trusted position.
type SegmentedMove struct {
	mu sync.Mutex
	segments []MoveSegment
	next int
	inFlight int
	state MoveState
	confirmedNode string
}

func NewSegmentedMove(path []string) (*SegmentedMove, error) {
	if len(path) < 2 { return nil, errors.New("move path requires at least two nodes") }
	segments := make([]MoveSegment, 0, len(path)-1)
	for i := 0; i < len(path)-1; i++ {
		if path[i] == "" || path[i+1] == "" || path[i] == path[i+1] { return nil, errors.New("invalid move path") }
		segments = append(segments, MoveSegment{From: path[i], To: path[i+1]})
	}
	return &SegmentedMove{segments: segments, inFlight: -1, state: MoveReady, confirmedNode: path[0]}, nil
}

// Lookahead returns at most two upcoming destinations for reservation. The
// caller may atomically reserve these resources before DispatchNext.
func (m *SegmentedMove) Lookahead() []string {
	m.mu.Lock(); defer m.mu.Unlock()
	if m.state == MoveUnknown || m.state == MoveDone { return nil }
	start := m.next
	end := start + 2
	if end > len(m.segments) { end = len(m.segments) }
	out := make([]string, 0, end-start)
	for i := start; i < end; i++ { out = append(out, m.segments[i].To) }
	return out
}

func (m *SegmentedMove) DispatchNext() (MoveSegment, error) {
	m.mu.Lock(); defer m.mu.Unlock()
	if m.state == MoveUnknown { return MoveSegment{}, ErrMoveUnknown }
	if m.state == MoveInFlight { return MoveSegment{}, ErrMoveAwaitingConfirmation }
	if m.next >= len(m.segments) { m.state = MoveDone; return MoveSegment{}, errors.New("move complete") }
	seg := m.segments[m.next]
	if m.confirmedNode != seg.From { return MoveSegment{}, ErrUntrustedPosition }
	m.inFlight = m.next
	m.state = MoveInFlight
	return seg, nil
}

// MarkUnknown is fail-closed. It does not advance the path, release occupancy,
// or permit DispatchNext; physical outcome must first be reconciled externally.
func (m *SegmentedMove) MarkUnknown() {
	m.mu.Lock(); defer m.mu.Unlock()
	if m.state == MoveInFlight { m.state = MoveUnknown }
}

// ConfirmNode is the only normal path that advances a physical MOVE. A caller
// should release the previous physical resource only after this succeeds.
func (m *SegmentedMove) ConfirmNode(node string, trusted bool) (string, error) {
	m.mu.Lock(); defer m.mu.Unlock()
	if !trusted || node == "" { return "", ErrUntrustedPosition }
	if m.state != MoveInFlight || m.inFlight < 0 { return "", errors.New("no move segment awaiting confirmation") }
	seg := m.segments[m.inFlight]
	if node != seg.To { return "", errors.New("confirmation does not match commanded destination") }
	previous := m.confirmedNode
	m.confirmedNode = node
	m.next = m.inFlight + 1
	m.inFlight = -1
	if m.next == len(m.segments) { m.state = MoveDone } else { m.state = MoveReady }
	return previous, nil
}

func (m *SegmentedMove) State() MoveState { m.mu.Lock(); defer m.mu.Unlock(); return m.state }
func (m *SegmentedMove) ConfirmedNode() string { m.mu.Lock(); defer m.mu.Unlock(); return m.confirmedNode }
