package domain

import (
	"testing"
	"time"
)

func TestDeadlineFailsClosed(t *testing.T) {
	now := time.Now()
	task := Task{ID: "t1", State: TaskPending, Deadline: now}
	if err := task.Start(now); err == nil { t.Fatal("expected expired deadline rejection") }
	if task.State != TaskFailed { t.Fatalf("state=%s", task.State) }
}

func TestUnknownMoveNeverBlindlyRetries(t *testing.T) {
	task := Task{ID: "t1", State: TaskRunning, Retry: RetryPolicy{MaxAttempts: 3}}
	task.RecordAttempt(CommandMove, ResultUnknown)
	if task.State != TaskUnknown { t.Fatalf("state=%s", task.State) }
	if task.CanRetry() { t.Fatal("UNKNOWN MOVE must never be blindly retried") }
}

func TestFailedNonPhysicalCommandCanRetryWithinBudget(t *testing.T) {
	task := Task{ID: "t1", State: TaskRunning, Retry: RetryPolicy{MaxAttempts: 3}}
	task.RecordAttempt(CommandNonPhysical, ResultFailed)
	if !task.CanRetry() { t.Fatal("safe failed non-physical command should be retryable within budget") }
}

func TestCancelRequiresTrustedPositionBeforeCompensation(t *testing.T) {
	task := Task{ID: "t1", State: TaskRunning}
	if err := task.RequestCancel(); err != nil { t.Fatal(err) }
	if err := task.BeginCompensation(Device{ID: "d1", Position: "N1", PositionTrusted: false}); err == nil {
		t.Fatal("untrusted physical position must not authorize compensation")
	}
	if task.State != TaskCancelRequested { t.Fatalf("failed recovery mutated state: %s", task.State) }
	if err := task.BeginCompensation(Device{ID: "d1", Position: "N1", PositionTrusted: true}); err != nil { t.Fatal(err) }
	if task.State != TaskCompensating { t.Fatalf("state=%s", task.State) }
}

func TestUnknownMoveRecoveryRequiresTrustedPosition(t *testing.T) {
	task := Task{ID: "t1", State: TaskUnknown, LastCommand: CommandMove, LastResult: ResultUnknown}
	if err := task.BeginCompensation(Device{ID: "d1"}); err == nil { t.Fatal("unknown move recovery must fail closed") }
	if err := task.BeginCompensation(Device{ID: "d1", Position: "N2", PositionTrusted: true}); err != nil { t.Fatal(err) }
}

func TestCompensationCompletionRequiresFreshTrustedPosition(t *testing.T) {
	task := Task{ID: "t1", State: TaskCompensating}
	if err := task.CompleteCompensation(Device{ID: "d1", Position: "SAFE", PositionTrusted: false}); err == nil {
		t.Fatal("untrusted post-compensation position must not complete cancellation")
	}
	if task.State != TaskCompensating { t.Fatalf("failed completion mutated state: %s", task.State) }
	if err := task.CompleteCompensation(Device{ID: "d1", PositionTrusted: true}); err == nil {
		t.Fatal("empty position must not complete cancellation")
	}
	if err := task.CompleteCompensation(Device{ID: "d1", Position: "SAFE", PositionTrusted: true}); err != nil { t.Fatal(err) }
	if task.State != TaskCancelled { t.Fatalf("state=%s", task.State) }
}

func TestCompensationCannotCompleteFromWrongLifecycleState(t *testing.T) {
	task := Task{ID: "t1", State: TaskUnknown}
	if err := task.CompleteCompensation(Device{ID: "d1", Position: "N2", PositionTrusted: true}); err == nil {
		t.Fatal("UNKNOWN task must first enter compensation")
	}
	if task.State != TaskUnknown { t.Fatalf("state mutated: %s", task.State) }
}
