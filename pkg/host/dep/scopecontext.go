package dep

import (
	"goms.io/azureml/mir/mir-vmagent/pkg/host/types"
)

type DefaultScopeContext struct {
	debug       bool
	concurrency bool
	parent      ScopeContextEx
	depDict     DepDict[ComponentGetter]
	data        ScopeDataEx
}

func NewScopeContext(parent ScopeContextEx, props Properties) *DefaultScopeContext {
	inheritProps := parent.GetScope().CopyProperties()
	inheritProps.Update(props)
	return &DefaultScopeContext{
		debug:       parent.IsDebug(),
		concurrency: parent.IsConcurrencyEnabled(),
		parent:      parent,
		depDict:     NewDependencyDictionary[ComponentGetter](),
		data:        NewScopeData(parent.IsConcurrencyEnabled(), inheritProps),
	}
}

func NewGlobalScopeContext(debug bool, props Properties) *DefaultScopeContext {
	copyProps := NewPropertiesFrom(props)
	return &DefaultScopeContext{
		debug:       debug,
		concurrency: false, // default value, overwrite by EnableConcurrency immediately
		parent:      nil,
		depDict:     NewDependencyDictionary[ComponentGetter](),
		data:        NewScopeData(false, copyProps),
	}
}

// EnableConcurrency is only called on Global context which does not have a parent to inherit necessary data
func (sc *DefaultScopeContext) EnableConcurrency(concurrency bool) {
	sc.concurrency = concurrency
	sc.data.EnableConcurrency(concurrency)
}

func (sc *DefaultScopeContext) Initialize(scopeType types.DataType, scopeInst Scopable) {
	sc.data.Initialize(scopeType, scopeInst)

	// if scopeType.Key() != ScopeType_Global.Key() {
	sc.AddDependency(scopeType, CompGetter(scopeInst))
	// }
}
func (sc *DefaultScopeContext) IsDebug() bool {
	return sc.debug
}
func (sc *DefaultScopeContext) IsConcurrencyEnabled() bool {
	return sc.concurrency
}
func (sc *DefaultScopeContext) GetParent() ScopeContextEx {
	return sc.parent
}
func (sc *DefaultScopeContext) GetScope() ScopeDataEx {
	return sc.data
}
func (sc *DefaultScopeContext) AddDependency(depType types.DataType, depGetter ComponentGetter) {
	if sc.depDict != nil {
		//fmt.Printf("scope %s: add dependency: %s@%p\n", sc.Name(), depType.Name(), inst)
		sc.depDict.AddDependencies(Dep[ComponentGetter](depType, depGetter))
	}
}
func (sc *DefaultScopeContext) getFromCurrentScope(depType types.DataType) interface{} {
	if sc.depDict != nil {
		if sc.depDict.ExistDependency(depType) {
			inst := sc.depDict.GetDependency(depType)()
			//fmt.Printf("scope %s: found dependency: %s@%p\n", sc.Name(), depType.Name(), inst)
			return inst
		}
	}

	return nil
}
func (sc *DefaultScopeContext) GetDependency(depType types.DataType) interface{} {
	// match contextual deps in ancester scopes recursively
	//fmt.Printf("scope %s: get dependency: %s\n", sc.Name(), depType.Name())
	inst := sc.getFromCurrentScope(depType)
	if inst != nil {
		return inst
	} else {
		if sc.parent != nil {
			return sc.parent.GetDependency(depType)
		} else {
			return nil
		}
	}
}
func (sc *DefaultScopeContext) IsGlobal() bool {
	ty, _ := sc.data.GetInstance()
	return ty.Key() == ScopeType_Global.Key()
}

func (sc *DefaultScopeContext) Type() string { return ContextType_Scope }
func (sc *DefaultScopeContext) Name() string {
	return sc.data.GetTypeName()
}
func (sc *DefaultScopeContext) ScopeId() string {
	return sc.data.GetScopeId()
}
