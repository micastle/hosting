package dep

import (
	"fmt"
	"sync"

	"goms.io/azureml/mir/mir-vmagent/pkg/host/types"
)

type DefaultScopeData struct {
	scopeType types.DataType
	scopeInst Scopable

	concurrency bool
	mutex       sync.Mutex
	records     map[interface{}]ScopedCompRecord

	properties Properties
}

func NewScopeData(concurrency bool, props Properties) *DefaultScopeData {
	return &DefaultScopeData{
		concurrency: concurrency,
		records:     make(map[interface{}]ScopedCompRecord),
		properties:  props,
	}
}

func (sd *DefaultScopeData) EnableConcurrency(concurrency bool) {
	sd.concurrency = concurrency
}
func (sd *DefaultScopeData) Initialize(scopeType types.DataType, scopeInst Scopable) {
	sd.scopeType = scopeType
	sd.scopeInst = scopeInst
}
func (sd *DefaultScopeData) GetInstance() (types.DataType, Scopable) {
	return sd.scopeType, sd.scopeInst
}
func (sd *DefaultScopeData) IsTypedScope() bool {
	return sd.scopeType.Key() != ScopeType_None.Key()
}
func (sd *DefaultScopeData) GetType() types.DataType {
	return sd.scopeType
}
func (sd *DefaultScopeData) GetTypeName() string {
	return sd.scopeType.Name()
}
func (sd *DefaultScopeData) GetScopeId() string {
	if sd.IsTypedScope() {
		if sd.scopeInst == nil {
			return sd.GetTypeName()
		} else {
			return fmt.Sprintf("%s@%p", sd.GetTypeName(), sd.scopeInst)
		}
	} else {
		return fmt.Sprintf("%p", sd.scopeInst)
	}
}
func (sd *DefaultScopeData) Match(reqType types.DataType) bool {
	if reqType.Key() == ScopeType_Any.Key() {
		return true
	}

	return sd.scopeType.Key() == reqType.Key()
	//return sd.scopeType.CheckCompatible(reqType)
}
func (sd *DefaultScopeData) Clear() {
	defer sd.mutex.Unlock()
	sd.mutex.Lock()

	sd.records = make(map[interface{}]ScopedCompRecord)
}
func (sd *DefaultScopeData) getRecord(compType types.DataType) ScopedCompRecord {
	defer sd.mutex.Unlock()
	sd.mutex.Lock()

	record, exist := sd.records[compType.Key()]
	if exist {
		return record
	}

	// not exist, create new record and store into records
	record = NewScopedCompRecord(compType.Name(), sd.concurrency)
	sd.records[compType.Key()] = record

	return record
}
func (sd *DefaultScopeData) GetCompRecord(compType types.DataType) ScopedCompRecord {
	return sd.getRecord(compType)
}

func (sd *DefaultScopeData) CopyProperties() Properties {
	return NewPropertiesFrom(sd.properties)
}

// methods for ScopedCompRecord
func NewScopedCompRecord(name string, concurrency bool) ScopedCompRecord {
	if concurrency {
		return &ScopedCompRecord_Lock{
			lock:     NewThreadSafeLock(name),
			creating: false,
			instance: nil,
			compCtxt: nil,
		}
	} else {
		return &ScopedCompRecord_NoLock{
			creating: false,
			instance: nil,
			compCtxt: nil,
		}
	}
}

type ScopedCompRecord_Lock struct {
	lock     SingletonLock
	creating bool
	instance interface{}
	compCtxt ContextEx
}

func (scr *ScopedCompRecord_Lock) Execute(createComponent FactoryAction) (interface{}, ContextEx, bool, bool) {
	if scr.instance == nil {
		defer scr.lock.Unlock()
		scr.lock.Lock()
		if scr.instance == nil {
			if !scr.creating {
				defer func() { scr.creating = false }()
				scr.creating = true
				scr.instance, scr.compCtxt = createComponent()
				return scr.instance, scr.compCtxt, false, false
			} else {
				return nil, nil, false, true
			}
		}
	}
	return scr.instance, scr.compCtxt, true, false
}

type ScopedCompRecord_NoLock struct {
	creating bool
	instance interface{}
	compCtxt ContextEx
}

func (scr *ScopedCompRecord_NoLock) Execute(createComponent FactoryAction) (interface{}, ContextEx, bool, bool) {
	if scr.instance == nil {
		if !scr.creating {
			defer func() { scr.creating = false }()
			scr.creating = true
			scr.instance, scr.compCtxt = createComponent()
			return scr.instance, scr.compCtxt, false, false
		} else {
			return nil, nil, false, true
		}
	}
	return scr.instance, scr.compCtxt, true, false
}
