package control

import (
	"errors"
	"sort"
	"sync"
)

var ErrConflict = errors.New("resource already reserved")

// ReservationManager owns logical movement resources. A reservation is released
// only by explicit trusted confirmation from the device-position layer.
type ReservationManager struct {
	mu      sync.Mutex
	owners  map[string]string
	waiting map[string]map[string]struct{}
}

func NewReservationManager() *ReservationManager {
	return &ReservationManager{owners: map[string]string{}, waiting: map[string]map[string]struct{}{}}
}

// Reserve atomically acquires all resources or none. Sorted acquisition makes
// behavior deterministic; conflicts are recorded in the wait-for graph.
func (m *ReservationManager) Reserve(taskID string, resources []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	uniq := uniqueSorted(resources)
	blockers := map[string]struct{}{}
	for _, r := range uniq {
		if owner := m.owners[r]; owner != "" && owner != taskID {
			blockers[owner] = struct{}{}
		}
	}
	if len(blockers) > 0 {
		m.waiting[taskID] = blockers
		return ErrConflict
	}
	for _, r := range uniq {
		m.owners[r] = taskID
	}
	delete(m.waiting, taskID)
	return nil
}

// ConfirmPositionAndRelease is intentionally the only release API. Callers
// must supply trustedPosition=true only after authoritative device feedback.
func (m *ReservationManager) ConfirmPositionAndRelease(taskID string, resources []string, trustedPosition bool) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !trustedPosition {
		return false
	}
	for _, r := range uniqueSorted(resources) {
		if m.owners[r] == taskID {
			delete(m.owners, r)
		}
	}
	delete(m.waiting, taskID)
	for waiter, deps := range m.waiting {
		delete(deps, taskID)
		if len(deps) == 0 {
			delete(m.waiting, waiter)
		}
	}
	return true
}

func (m *ReservationManager) Owner(resource string) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.owners[resource]
}

// DeadlockedTasks returns every task participating in a wait-for cycle.
func (m *ReservationManager) DeadlockedTasks() []string {
	m.mu.Lock()
	defer m.mu.Unlock()

	state := map[string]uint8{}
	stack := []string{}
	index := map[string]int{}
	dead := map[string]struct{}{}
	var visit func(string)
	visit = func(n string) {
		state[n] = 1
		index[n] = len(stack)
		stack = append(stack, n)
		for dep := range m.waiting[n] {
			if state[dep] == 0 {
				visit(dep)
			} else if state[dep] == 1 {
				for _, member := range stack[index[dep]:] {
					dead[member] = struct{}{}
				}
			}
		}
		stack = stack[:len(stack)-1]
		delete(index, n)
		state[n] = 2
	}
	for n := range m.waiting {
		if state[n] == 0 { visit(n) }
	}
	out := make([]string, 0, len(dead))
	for n := range dead { out = append(out, n) }
	sort.Strings(out)
	return out
}

func uniqueSorted(in []string) []string {
	set := map[string]struct{}{}
	for _, v := range in { if v != "" { set[v] = struct{}{} } }
	out := make([]string, 0, len(set))
	for v := range set { out = append(out, v) }
	sort.Strings(out)
	return out
}
