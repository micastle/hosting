package hosting

import (
	"fmt"
)

type LoopRunContext interface {
	LooperName() string
	IsStopped() bool
	SetStopped()
}

type ScopeContextBase interface {
	HasVariable(key string, localScope bool) bool
	GetVariable(key string) interface{}
	SetVariable(key string, value interface{})
}
func GetVariable[T any](scope ScopeContextBase, key string) T {
	return scope.GetVariable(key).(T)
}
func SetVariable[T any](scope ScopeContextBase, key string, value T) {
	scope.SetVariable(key, value)
}

type LoopGlobalContext interface {
	ScopeContextBase

	LooperName() string
	GetLooperContext() ServiceContext
}

type ScopeOption int8

const (
	Current ScopeOption = iota
	TopLevel
	Global

	_minType = Current
	_maxType = Global
)

type ScopeContext interface {
	ScopeContextBase

	GetLoopRunContext() LoopRunContext
	GetLooperContext() ServiceContext

	ExitScope(ScopeOption)
	IsExit() bool
}

type VariableSet struct {
	Entries map[string]interface{}
}

func NewVariableSet() *VariableSet {
	return &VariableSet{
		Entries: make(map[string]interface{}),
	}
}
func (vs *VariableSet) Get(key string) interface{} {
	return vs.Entries[key]
}
func (vs *VariableSet) Set(key string, value interface{}) {
	vs.Entries[key] = value
}
func (vs *VariableSet) Exist(key string) bool {
	_, exist := vs.Entries[key]
	return exist
}

type DefaultLoopGlobalContext struct {
	Looper    *DefaultLooper
	Variables *VariableSet
}

func NewLoopGlobalContext(looper *DefaultLooper) *DefaultLoopGlobalContext {
	return &DefaultLoopGlobalContext{
		Looper:    looper,
		Variables: NewVariableSet(),
	}
}

func (lrc *DefaultLoopGlobalContext) LooperName() string {
	return lrc.Looper.Name()
}
func (lrc *DefaultLoopGlobalContext) GetLooperContext() ServiceContext {
	return lrc.Looper.context
}

func (lrc *DefaultLoopGlobalContext) HasVariable(key string, localScope bool) bool {
	return lrc.Variables.Exist(key)
}
func (lrc *DefaultLoopGlobalContext) GetVariable(key string) interface{} {
	if !lrc.Variables.Exist(key) {
		panic(fmt.Errorf("variable not exist: %s", key))
	}
	return lrc.Variables.Get(key)
}
func (lrc *DefaultLoopGlobalContext) SetVariable(key string, value interface{}) {
	lrc.Variables.Set(key, value)
}

type DefaultLoopRunContext struct {
	parent LoopGlobalContext

	Stopped   bool
	Variables *VariableSet
}

func NewLoopRunContext(globalCtxt LoopGlobalContext) *DefaultLoopRunContext {
	return &DefaultLoopRunContext{
		parent:    globalCtxt,
		Stopped:   false,
		Variables: NewVariableSet(),
	}
}

func (gsc *DefaultLoopRunContext) LooperName() string {
	return gsc.parent.LooperName()
}
func (gsc *DefaultLoopRunContext) IsStopped() bool {
	return gsc.Stopped
}
func (gsc *DefaultLoopRunContext) SetStopped() {
	gsc.Stopped = true
}

func (gsc *DefaultLoopRunContext) GetLoopRunContext() LoopRunContext {
	return gsc
}
func (gsc *DefaultLoopRunContext) GetLooperContext() ServiceContext {
	return gsc.parent.GetLooperContext()
}

func (gsc *DefaultLoopRunContext) HasVariable(key string, localScope bool) bool {
	exist := gsc.Variables.Exist(key)
	if exist {
		return true
	}
	if localScope {
		return false
	}
	return gsc.parent.HasVariable(key, true)
}
func (gsc *DefaultLoopRunContext) GetVariable(key string) interface{} {
	if gsc.Variables.Exist(key) {
		return gsc.Variables.Get(key)
	} else {
		return gsc.parent.GetVariable(key)
	}
}
func (gsc *DefaultLoopRunContext) SetVariable(key string, value interface{}) {
	gsc.Variables.Set(key, value)
}

func (gsc *DefaultLoopRunContext) ExitScope(option ScopeOption) {
	gsc.SetStopped()
}

func (gsc *DefaultLoopRunContext) IsExit() bool {
	return gsc.IsStopped()
}

type GroupScopeContext struct {
	group  ProcessorGroup
	parent ScopeContext

	//Entries  map[string]interface{}
	Variables *VariableSet

	complete bool
}

func NewGroupScopeContext(pg ProcessorGroup, parentCtxt ScopeContext) *GroupScopeContext {
	return &GroupScopeContext{
		group:     pg,
		parent:    parentCtxt,
		complete:  false,
		Variables: NewVariableSet(),
	}
}

func (gsc *GroupScopeContext) GetLoopRunContext() LoopRunContext {
	return gsc.parent.GetLoopRunContext()
}

func (gsc *GroupScopeContext) GetLooperContext() ServiceContext {
	return gsc.parent.GetLooperContext()
}

func (gsc *GroupScopeContext) HasVariable(key string, localScope bool) bool {
	exist := gsc.Variables.Exist(key)
	if exist {
		return true
	}
	if localScope {
		return false
	}
	return gsc.parent.HasVariable(key, false)
}
func (gsc *GroupScopeContext) GetVariable(key string) interface{} {
	if gsc.Variables.Exist(key) {
		return gsc.Variables.Get(key)
	} else {
		return gsc.parent.GetVariable(key)
	}
}

func (gsc *GroupScopeContext) SetVariable(key string, value interface{}) {
	gsc.Variables.Set(key, value)
}

func (gsc *GroupScopeContext) ExitScope(option ScopeOption) {
	gsc.complete = true
	if option >= TopLevel {
		gsc.GetLoopRunContext().SetStopped()
	}
}

func (gsc *GroupScopeContext) IsExit() bool {
	return gsc.complete
}
