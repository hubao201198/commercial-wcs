package control

import (
	"errors"
	"testing"
)

func TestRestoredLeaseFencesPreRestartGeneration(t *testing.T) {
	store := NewMemoryAuthorityStore(20)
	first, err := RestoreControlLease(store)
	if err != nil { t.Fatal(err) }
	move, _ := NewSegmentedMove([]string{"A", "B"})
	promoted, err := first.Promote(move)
	if err != nil { t.Fatal(err) }
	if promoted != 21 { t.Fatalf("generation=%d", promoted) }

	restarted, err := RestoreControlLease(store)
	if err != nil { t.Fatal(err) }
	if restarted.Generation() != 21 { t.Fatalf("restored=%d", restarted.Generation()) }
	if _, err := restarted.DispatchNext(move, 20); !errors.Is(err, ErrStaleControlGeneration) {
		t.Fatalf("pre-restart authority dispatched: %v", err)
	}
}

func TestConcurrentPromotionUsesCompareAndSwap(t *testing.T) {
	store := NewMemoryAuthorityStore(30)
	a, _ := RestoreControlLease(store)
	b, _ := RestoreControlLease(store)
	moveA, _ := NewSegmentedMove([]string{"A", "B"})
	moveB, _ := NewSegmentedMove([]string{"A", "C"})

	if generation, err := a.Promote(moveA); err != nil || generation != 31 {
		t.Fatalf("first promotion generation=%d err=%v", generation, err)
	}
	if _, err := b.Promote(moveB); !errors.Is(err, ErrAuthorityConflict) {
		t.Fatalf("stale competing promotion accepted: %v", err)
	}
	if _, err := b.DispatchNext(moveB, 30); !errors.Is(err, ErrStaleControlGeneration) {
		t.Fatalf("stale controller dispatched after competing promotion: %v", err)
	}
}

func TestDurablePromotionStillRequiresTrustedQuiescentState(t *testing.T) {
	store := NewMemoryAuthorityStore(40)
	lease, _ := RestoreControlLease(store)
	move, _ := NewSegmentedMove([]string{"A", "B"})
	if _, err := move.DispatchNext(); err != nil { t.Fatal(err) }
	if _, err := lease.Promote(move); !errors.Is(err, ErrTakeoverUnsafeState) {
		t.Fatalf("in-flight durable takeover accepted: %v", err)
	}
	generation, _ := store.Load()
	if generation != 40 { t.Fatalf("unsafe takeover persisted generation=%d", generation) }
}

func TestAuthorizeDetectsAuthorityAdvancedByAnotherController(t *testing.T) {
	store := NewMemoryAuthorityStore(50)
	lease, _ := RestoreControlLease(store)
	if err := store.CompareAndSwap(50, 51); err != nil { t.Fatal(err) }
	if err := lease.Authorize(50); !errors.Is(err, ErrStaleControlGeneration) {
		t.Fatalf("cached stale authority remained valid: %v", err)
	}
}
