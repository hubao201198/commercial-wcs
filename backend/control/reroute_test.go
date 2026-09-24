package control

import (
	"errors"
	"reflect"
	"testing"
)

func TestSafeRerouteFromConfirmedQuiescentNode(t *testing.T) {
	move, err := NewSegmentedMove([]string{"A", "B", "C"})
	if err != nil { t.Fatal(err) }
	if _, err := move.DispatchNext(); err != nil { t.Fatal(err) }
	if _, err := move.ConfirmNode("B", true); err != nil { t.Fatal(err) }

	rerouted, err := SafeReroute(move, []string{"B", "D", "E"})
	if err != nil { t.Fatal(err) }
	if got := rerouted.ConfirmedNode(); got != "B" { t.Fatalf("confirmed node=%q", got) }
	if got := rerouted.Lookahead(); !reflect.DeepEqual(got, []string{"D", "E"}) { t.Fatalf("lookahead=%v", got) }
	seg, err := rerouted.DispatchNext()
	if err != nil { t.Fatal(err) }
	if seg.From != "B" || seg.To != "D" { t.Fatalf("segment=%+v", seg) }
}

func TestSafeRerouteRefusesInFlightMove(t *testing.T) {
	move, _ := NewSegmentedMove([]string{"A", "B", "C"})
	if _, err := move.DispatchNext(); err != nil { t.Fatal(err) }
	if _, err := SafeReroute(move, []string{"A", "D"}); !errors.Is(err, ErrRerouteUnsafeState) {
		t.Fatalf("in-flight reroute accepted: %v", err)
	}
}

func TestSafeRerouteRefusesUnknownMove(t *testing.T) {
	move, _ := NewSegmentedMove([]string{"A", "B", "C"})
	if _, err := move.DispatchNext(); err != nil { t.Fatal(err) }
	move.MarkUnknown()
	if _, err := SafeReroute(move, []string{"A", "D"}); !errors.Is(err, ErrRerouteUnsafeState) {
		t.Fatalf("UNKNOWN move reroute accepted: %v", err)
	}
	if move.ConfirmedNode() != "A" { t.Fatalf("failed reroute mutated position: %s", move.ConfirmedNode()) }
}

func TestSafeRerouteRequiresExactTrustedOrigin(t *testing.T) {
	move, _ := NewSegmentedMove([]string{"A", "B", "C"})
	if _, err := SafeReroute(move, []string{"B", "D"}); !errors.Is(err, ErrReroutePositionMismatch) {
		t.Fatalf("mismatched reroute origin accepted: %v", err)
	}
	if _, err := SafeReroute(move, []string{"A"}); !errors.Is(err, ErrReroutePositionMismatch) {
		t.Fatalf("short replacement path accepted: %v", err)
	}
}
