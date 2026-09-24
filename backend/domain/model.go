package domain

import (
	"errors"
	"time"
)

type DeviceMode string

const (
	ModeAuto DeviceMode = "AUTO"
	ModeManual DeviceMode = "MANUAL"
	ModeMaintenance DeviceMode = "MAINTENANCE"
)

type Device struct {
	ID string
	Kind string
	Mode DeviceMode
	Online bool
	Position string
	PositionTrusted bool
	UpdatedAt time.Time
}

type AlarmState string

const (
	AlarmActive AlarmState = "ACTIVE"
	AlarmAcknowledged AlarmState = "ACKNOWLEDGED"
	AlarmRecovered AlarmState = "RECOVERED"
)

type Alarm struct {
	ID string
	DeviceID string
	Code string
	Message string
	State AlarmState
	RaisedAt time.Time
}

type TaskState string

const (
	TaskPending TaskState = "PENDING"
	TaskRunning TaskState = "RUNNING"
	TaskCancelRequested TaskState = "CANCEL_REQUESTED"
	TaskCompensating TaskState = "COMPENSATING"
	TaskCancelled TaskState = "CANCELLED"
	TaskFailed TaskState = "FAILED"
	TaskCompleted TaskState = "COMPLETED"
	TaskUnknown TaskState = "UNKNOWN"
)

type CommandKind string

const (
	CommandMove CommandKind = "MOVE"
	CommandNonPhysical CommandKind = "NON_PHYSICAL"
)

type CommandResult string

const (
	ResultSucceeded CommandResult = "SUCCEEDED"
	ResultFailed CommandResult = "FAILED"
	ResultUnknown CommandResult = "UNKNOWN"
)

type RetryPolicy struct {
	MaxAttempts int
}

type Task struct {
	ID string
	DeviceID string
	State TaskState
	Deadline time.Time
	Retry RetryPolicy
	Attempts int
	LastCommand CommandKind
	LastResult CommandResult
}

func (t *Task) Start(now time.Time) error {
	if t.State != TaskPending { return errors.New("task is not pending") }
	if !t.Deadline.IsZero() && !now.Before(t.Deadline) {
		t.State = TaskFailed
		return errors.New("task deadline exceeded")
	}
	t.State = TaskRunning
	return nil
}

func (t *Task) RecordAttempt(kind CommandKind, result CommandResult) {
	t.Attempts++
	t.LastCommand = kind
	t.LastResult = result
	if result == ResultUnknown {
		t.State = TaskUnknown
	} else if result == ResultFailed {
		t.State = TaskFailed
	} else if result == ResultSucceeded {
		t.State = TaskCompleted
	}
}

// CanRetry deliberately fails closed for an UNKNOWN MOVE: the physical device may
// have executed the command even when the controller did not observe the result.
func (t Task) CanRetry() bool {
	if t.LastCommand == CommandMove && t.LastResult == ResultUnknown { return false }
	if t.Retry.MaxAttempts <= 0 { return false }
	return t.Attempts < t.Retry.MaxAttempts && t.State == TaskFailed
}

func (t *Task) RequestCancel() error {
	if t.State != TaskRunning { return errors.New("only running tasks can request cancellation") }
	t.State = TaskCancelRequested
	return nil
}

// BeginCompensation requires trusted physical position before changing a task
// that may already have moved equipment. This prevents cancellation/recovery
// from inventing a safe physical state.
func (t *Task) BeginCompensation(device Device) error {
	if t.State != TaskCancelRequested && t.State != TaskUnknown {
		return errors.New("task is not awaiting recovery")
	}
	if !device.PositionTrusted || device.Position == "" {
		return errors.New("trusted physical position required")
	}
	t.State = TaskCompensating
	return nil
}

// CompleteCompensation is a second physical-state gate. A compensation command
// being issued or acknowledged is not enough to declare cancellation complete:
// the controller must reconcile a trusted, non-empty device position after the
// compensation action. This keeps occupied-resource release decisions separate
// from transport/command acknowledgements.
func (t *Task) CompleteCompensation(device Device) error {
	if t.State != TaskCompensating {
		return errors.New("task is not compensating")
	}
	if !device.PositionTrusted || device.Position == "" {
		return errors.New("trusted physical position required to complete compensation")
	}
	t.State = TaskCancelled
	return nil
}
