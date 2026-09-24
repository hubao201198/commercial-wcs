package control

import (
	"errors"
	"sync"
)

var (
	ErrAuthorityConflict = errors.New("control authority generation changed")
	ErrAuthorityRollback = errors.New("control authority generation rollback")
)

// AuthorityStore is the durability/consensus boundary for failover authority.
// A production implementation may use PostgreSQL or a consensus service, but it
// MUST provide atomic compare-and-swap semantics across controller processes.
type AuthorityStore interface {
	Load() (uint64, error)
	CompareAndSwap(expected, next uint64) error
}

// MemoryAuthorityStore is a deterministic implementation used by tests and
// single-process simulation. It deliberately implements the same CAS contract
// required from a production durable store.
type MemoryAuthorityStore struct {
	mu         sync.Mutex
	generation uint64
}

func NewMemoryAuthorityStore(initial uint64) *MemoryAuthorityStore {
	if initial == 0 {
		initial = 1
	}
	return &MemoryAuthorityStore{generation: initial}
}

func (s *MemoryAuthorityStore) Load() (uint64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.generation, nil
}

func (s *MemoryAuthorityStore) CompareAndSwap(expected, next uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.generation != expected {
		return ErrAuthorityConflict
	}
	if next <= expected {
		return ErrAuthorityRollback
	}
	s.generation = next
	return nil
}

// DurableControlLease restores its generation from AuthorityStore and advances
// authority with CAS before exposing the promoted generation. A controller with
// a stale cached generation therefore cannot successfully promote over a newer
// owner. Physical-state reconciliation remains mandatory through SegmentedMove.
type DurableControlLease struct {
	mu         sync.Mutex
	store      AuthorityStore
	generation uint64
}

func RestoreControlLease(store AuthorityStore) (*DurableControlLease, error) {
	if store == nil {
		return nil, errors.New("authority store required")
	}
	generation, err := store.Load()
	if err != nil {
		return nil, err
	}
	if generation == 0 {
		return nil, ErrAuthorityRollback
	}
	return &DurableControlLease{store: store, generation: generation}, nil
}

func (l *DurableControlLease) Generation() uint64 {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.generation
}

func (l *DurableControlLease) Authorize(generation uint64) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	current, err := l.store.Load()
	if err != nil {
		return err
	}
	if generation != l.generation || generation != current {
		return ErrStaleControlGeneration
	}
	return nil
}

func (l *DurableControlLease) Promote(move *SegmentedMove) (uint64, error) {
	if move == nil || move.State() != MoveReady || move.ConfirmedNode() == "" {
		return 0, ErrTakeoverUnsafeState
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	next := l.generation + 1
	if err := l.store.CompareAndSwap(l.generation, next); err != nil {
		return 0, err
	}
	l.generation = next
	return next, nil
}

func (l *DurableControlLease) DispatchNext(move *SegmentedMove, generation uint64) (MoveSegment, error) {
	if err := l.Authorize(generation); err != nil {
		return MoveSegment{}, err
	}
	if move == nil {
		return MoveSegment{}, ErrTakeoverUnsafeState
	}
	return move.DispatchNext()
}
