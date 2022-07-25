package dep

import (
	"fmt"
	"sync"

	"goms.io/azureml/mir/mir-vmagent/pkg/host/types"
)

type RecurrenceTracker interface {
	Execute(createComponent FactoryAction) (interface{}, ContextEx, bool)
}

// implement RecurrenceTracker but track nothing
type NoneTracker struct{}

func newNoneTracker() *NoneTracker {
	return &NoneTracker{}
}
func (nt *NoneTracker) Execute(createComponent FactoryAction) (interface{}, ContextEx, bool) {
	instance, compCtxt := createComponent()
	return instance, compCtxt, false
}

var s_nontracker *NoneTracker
var once sync.Once

func GetNonTracker() *NoneTracker {
	once.Do(func() {
		s_nontracker = newNoneTracker()
	})
	return s_nontracker
}

// implement RecurrenceTracker, tracking the recurrence for transient component
type DefaultRecurrenceTracker struct {
	mgr  RecurrenceManager
	data *TrackingData
}
type TrackingData struct {
	gid            int32
	recursiveDepth uint32
}

func NewRecurrenceTracker(mgr RecurrenceManager, data *TrackingData) *DefaultRecurrenceTracker {
	return &DefaultRecurrenceTracker{
		mgr:  mgr,
		data: data,
	}
}
func (rt *DefaultRecurrenceTracker) Execute(createComponent FactoryAction) (interface{}, ContextEx, bool) {
	if rt.IsNextRecurrenceAllowed() {
		defer func() { rt.DecreaseRecurrence() }()
		rt.IncreaseRecurrence()
		instance, compCtxt := createComponent()
		return instance, compCtxt, false
	} else {
		return nil, nil, true
	}
}
func (rt *DefaultRecurrenceTracker) IsNextRecurrenceAllowed() bool {
	return rt.data.recursiveDepth < rt.mgr.GetMaxAllowedRecursive()
}
func (rt *DefaultRecurrenceTracker) IncreaseRecurrence() {
	rt.data.recursiveDepth++
	fmt.Printf("goroutine %d: recurrence on %s - %d\n", rt.data.gid, rt.mgr.GetComponentType().FullName(), rt.data.recursiveDepth)
}
func (rt *DefaultRecurrenceTracker) DecreaseRecurrence() {
	rt.data.recursiveDepth--
	if rt.data.recursiveDepth <= 0 {
		rt.mgr.RemoveTracker(rt.data.gid)
	}
}

type RecurrenceManager interface {
	GetMaxAllowedRecursive() uint32
	GetComponentType() types.DataType

	GetTracker() RecurrenceTracker
	RemoveTracker(gid int32)
}
type DefaultRecurrenceManager struct {
	maxRecursiveDepth uint32
	trackRecurrence   bool
	compType          types.DataType

	mutex       sync.Mutex
	recurrences map[int32]*TrackingData
}

func NewRecurrenceManager(options *LifecycleOptions, compType types.DataType) *DefaultRecurrenceManager {
	return &DefaultRecurrenceManager{
		maxRecursiveDepth: options.MaxAllowedRecurrence,
		trackRecurrence:   options.TrackTransientRecurrence,
		compType:          compType,
		recurrences:       make(map[int32]*TrackingData),
	}
}
func (rt *DefaultRecurrenceManager) GetMaxAllowedRecursive() uint32 {
	return rt.maxRecursiveDepth
}
func (rt *DefaultRecurrenceManager) GetComponentType() types.DataType {
	return rt.compType
}
func (rt *DefaultRecurrenceManager) GetTracker() RecurrenceTracker {
	if rt.trackRecurrence {
		gid := goroutine_id()
		data := rt.getTrackingData(gid)
		return NewRecurrenceTracker(rt, data)
	} else {
		return GetNonTracker()
	}
}
func (rt *DefaultRecurrenceManager) RemoveTracker(gid int32) {
	defer func() { rt.mutex.Unlock() }()
	rt.mutex.Lock()

	delete(rt.recurrences, gid)
}
func (rt *DefaultRecurrenceManager) getTrackingData(gid int32) *TrackingData {
	defer func() { rt.mutex.Unlock() }()
	rt.mutex.Lock()

	data, exist := rt.recurrences[gid]
	if !exist {
		data = &TrackingData{
			gid:            gid,
			recursiveDepth: 0,
		}
		rt.recurrences[gid] = data
	}

	return data
}
