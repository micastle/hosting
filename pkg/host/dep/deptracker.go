package dep

import "sync"

type DefaultDependencyTracker struct {
	parent ScopeContextEx

	mutex      sync.Mutex
	dependents []ContextEx
}

func NewDependencyTracker(scopeCtxt ScopeContextEx) *DefaultDependencyTracker {
	return &DefaultDependencyTracker{
		parent:     scopeCtxt,
		dependents: make([]ContextEx, 0),
	}
}

func (dt *DefaultDependencyTracker) AddDependents(deps ...ContextEx) {
	defer dt.mutex.Unlock()
	dt.mutex.Lock()
	dt.dependents = append(dt.dependents, deps...)
}
func (dt *DefaultDependencyTracker) GetParent() ScopeContextEx {
	return dt.parent
}
func (dt *DefaultDependencyTracker) GetDependents() []ContextEx {
	defer dt.mutex.Unlock()
	dt.mutex.Lock()
	result := dt.dependents
	return result
}
