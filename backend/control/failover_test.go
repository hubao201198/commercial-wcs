package control

import (
	"errors"
	"testing"
)

func TestPromotionFencesOldController(t *testing.T) {
	move, err := NewSegmentedMove([]string{"A", "B", "C"})
	if err != nil { t.Fatal(err) }
	lease := NewControlLease(7)
	old := lease.Generation()
	newGeneration, err := lease.Promote(move)
	if err != nil { t.Fatal(err) }
	if newGeneration != old+1 { t.Fatalf("generation=%d want %d", newGeneration, old+1) }
	if _, err := lease.DispatchNext(move, old); !errors.Is(err, ErrStaleControlGeneration) {
		t.Fatalf("stale controller dispatched: %v", err)
	}
	seg, err := lease.DispatchNext(move, newGeneration)
	if err != nil { t.Fatal(err) }
	if seg.From != "A" || seg.To != "B" { t.Fatalf("segment=%+v", seg) }
}

func TestPromotionRefusesInFlightPhysicalCommand(t *testing.T) {
	move, _ := NewSegmentedMove([]string{"A", "B", "C"})
	if _, err := move.DispatchNext(); err != nil { t.Fatal(err) }
	lease := NewControlLease(3)
	before := lease.Generation()
	if _, err := lease.Promote(move); !errors.Is(err, ErrTakeoverUnsafeState) {
		t.Fatalf("in-flight takeover accepted: %v", err)
	}
	if lease.Generation() != before { t.Fatal("failed takeover advanced generation") }
}

func TestPromotionRefusesUnknownPhysicalOutcome(t *testing.T) {
	move, _ := NewSegmentedMove([]string{"A", "B", "C"})
	if _, err := move.DispatchNext(); err != nil { t.Fatal(err) }
	move.MarkUnknown()
	lease := NewControlLease(4)
	if _, err := lease.Promote(move); !errors.Is(err, ErrTakeoverUnsafeState) {
		t.Fatalf("UNKNOWN takeover accepted: %v", err)
	}
	if move.ConfirmedNode() != "A" { t.Fatalf("takeover mutated position: %s", move.ConfirmedNode()) }
}

func TestStaleGenerationCannotMutateMove(t *testing.T) {
	move, _ := NewSegmentedMove([]string{"A", "B"})
	lease := NewControlLease(10)
	if _, err := lease.DispatchNext(move, 9); !errors.Is(err, ErrStaleControlGeneration) {
		t.Fatalf("stale generation accepted: %v", err)
	}
	if move.State() != MoveReady || move.ConfirmedNode() != "A" {
		t.Fatalf("stale dispatch mutated move: state=%s node=%s", move.State(), move.ConfirmedNode())
	}
}
