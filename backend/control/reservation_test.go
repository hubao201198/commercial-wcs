package control

import (
	"errors"
	"reflect"
	"testing"
)

func TestReservationIsAtomicOnConflict(t *testing.T) {
	m := NewReservationManager()
	if err := m.Reserve("task-a", []string{"segment-2"}); err != nil { t.Fatal(err) }
	if err := m.Reserve("task-b", []string{"segment-1", "segment-2"}); !errors.Is(err, ErrConflict) { t.Fatalf("want conflict, got %v", err) }
	if got := m.Owner("segment-1"); got != "" { t.Fatalf("partial reservation leaked: %q", got) }
	if got := m.Owner("segment-2"); got != "task-a" { t.Fatalf("owner changed: %q", got) }
}

func TestUntrustedPositionCannotReleasePhysicalResource(t *testing.T) {
	m := NewReservationManager()
	if err := m.Reserve("move-1", []string{"aisle-A"}); err != nil { t.Fatal(err) }
	if m.ConfirmPositionAndRelease("move-1", []string{"aisle-A"}, false) { t.Fatal("untrusted confirmation released resource") }
	if got := m.Owner("aisle-A"); got != "move-1" { t.Fatalf("resource released without trusted position: %q", got) }
	if !m.ConfirmPositionAndRelease("move-1", []string{"aisle-A"}, true) { t.Fatal("trusted confirmation rejected") }
	if got := m.Owner("aisle-A"); got != "" { t.Fatalf("resource still held: %q", got) }
}

func TestDeadlockCycleDetected(t *testing.T) {
	m := NewReservationManager()
	if err := m.Reserve("A", []string{"r1"}); err != nil { t.Fatal(err) }
	if err := m.Reserve("B", []string{"r2"}); err != nil { t.Fatal(err) }
	if err := m.Reserve("A", []string{"r2"}); !errors.Is(err, ErrConflict) { t.Fatal("A should wait for B") }
	if err := m.Reserve("B", []string{"r1"}); !errors.Is(err, ErrConflict) { t.Fatal("B should wait for A") }
	if got, want := m.DeadlockedTasks(), []string{"A", "B"}; !reflect.DeepEqual(got, want) { t.Fatalf("deadlock=%v want=%v", got, want) }
}

func TestTrustedReleaseClearsWaitDependency(t *testing.T) {
	m := NewReservationManager()
	_ = m.Reserve("A", []string{"r1"})
	_ = m.Reserve("B", []string{"r2"})
	_ = m.Reserve("A", []string{"r2"})
	_ = m.Reserve("B", []string{"r1"})
	m.ConfirmPositionAndRelease("A", []string{"r1"}, true)
	if got := m.DeadlockedTasks(); len(got) != 0 { t.Fatalf("stale wait edge retained: %v", got) }
}
