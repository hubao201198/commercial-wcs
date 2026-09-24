package control

import (
	"errors"
	"reflect"
	"testing"
)

func TestSegmentedMoveTwoSegmentLookaheadAndConfirmation(t *testing.T) {
	m, err := NewSegmentedMove([]string{"A", "B", "C", "D"})
	if err != nil { t.Fatal(err) }
	if got := m.Lookahead(); !reflect.DeepEqual(got, []string{"B", "C"}) { t.Fatalf("lookahead=%v", got) }
	seg, err := m.DispatchNext(); if err != nil { t.Fatal(err) }
	if seg.From != "A" || seg.To != "B" { t.Fatalf("segment=%+v", seg) }
	if _, err := m.DispatchNext(); !errors.Is(err, ErrMoveAwaitingConfirmation) { t.Fatalf("expected confirmation gate, got %v", err) }
	previous, err := m.ConfirmNode("B", true); if err != nil { t.Fatal(err) }
	if previous != "A" { t.Fatalf("previous=%s", previous) }
	if got := m.Lookahead(); !reflect.DeepEqual(got, []string{"C", "D"}) { t.Fatalf("lookahead=%v", got) }
}

func TestUnknownSegmentNeverBlindlyReplays(t *testing.T) {
	m, _ := NewSegmentedMove([]string{"A", "B", "C"})
	if _, err := m.DispatchNext(); err != nil { t.Fatal(err) }
	m.MarkUnknown()
	if m.State() != MoveUnknown { t.Fatalf("state=%s", m.State()) }
	if _, err := m.DispatchNext(); !errors.Is(err, ErrMoveUnknown) { t.Fatalf("UNKNOWN MOVE replayed: %v", err) }
	if m.ConfirmedNode() != "A" { t.Fatalf("unknown outcome advanced physical position to %s", m.ConfirmedNode()) }
	if got := m.Lookahead(); got != nil { t.Fatalf("unknown move exposed reservation lookahead: %v", got) }
}

func TestUntrustedOrWrongNodeCannotAdvanceOrRelease(t *testing.T) {
	m, _ := NewSegmentedMove([]string{"A", "B", "C"})
	if _, err := m.DispatchNext(); err != nil { t.Fatal(err) }
	if previous, err := m.ConfirmNode("B", false); !errors.Is(err, ErrUntrustedPosition) || previous != "" { t.Fatalf("untrusted confirmation accepted: previous=%q err=%v", previous, err) }
	if previous, err := m.ConfirmNode("C", true); err == nil || previous != "" { t.Fatalf("wrong node confirmation accepted: previous=%q err=%v", previous, err) }
	if m.ConfirmedNode() != "A" { t.Fatalf("failed confirmation mutated physical position: %s", m.ConfirmedNode()) }
	previous, err := m.ConfirmNode("B", true); if err != nil { t.Fatal(err) }
	if previous != "A" { t.Fatalf("release authority should identify previous trusted node, got %s", previous) }
}

func TestPathValidationFailsClosed(t *testing.T) {
	for _, path := range [][]string{{"A"}, {"A", ""}, {"A", "A"}} {
		if _, err := NewSegmentedMove(path); err == nil { t.Fatalf("accepted unsafe path %v", path) }
	}
}
