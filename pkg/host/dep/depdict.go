package dep

import (
	"fmt"

	"goms.io/azureml/mir/mir-vmagent/pkg/host/types"
)

type ComponentGetter func() any

type TypedGetter[T any] func() T

type FactoryConstraint interface {
	FactoryMethod | ComponentGetter
}
type Dependency[T FactoryConstraint] struct {
	depType    types.DataType
	depFactory T
}

func getter[T any](instance T) TypedGetter[T] {
	return func() T { return instance }
}

func CompGetter(instance any) ComponentGetter {
	return func() any { return instance }
}
func Getter[T any](instance T) TypedGetter[T] {
	return getter[T](instance)
}
func Factorize[T any](instance T) FactoryMethod {
	return func(Context, types.DataType, Properties) any { return instance }
}

func Dep[C FactoryConstraint](depType types.DataType, depFactory C) *Dependency[C] {
	return &Dependency[C]{
		depType:    depType,
		depFactory: depFactory,
	}
}
func DepInst[T any](instance T) *Dependency[ComponentGetter] {
	return Dep[ComponentGetter](types.Get[T](), func() any { return getter[T](instance)() })
}
func DepFact[T any, C FactoryConstraint](depFactory C) *Dependency[C] {
	return Dep[C](types.Get[T](), depFactory)
}

type DepDictReader[T FactoryConstraint] interface {
	GetAllDeps() []types.DataType
	GetDependency(depType types.DataType) T
	ExistDependency(depType types.DataType) bool
	Count() int
}
type DepDictWriter[T FactoryConstraint] interface {
	AddDependency(factory T, depType types.DataType)
	AddDependencies(deps ...*Dependency[T])
}
type DepDict[T FactoryConstraint]interface {
	DepDictReader[T]
	DepDictWriter[T]
}

type DefaultDepDict[T FactoryConstraint] struct {
	dependencies map[interface{}]T
}

func NewDependencyDictionary[T FactoryConstraint]() *DefaultDepDict[T] {
	return &DefaultDepDict[T]{
		dependencies: make(map[interface{}]T),
	}
}

func (dd *DefaultDepDict[T]) AddDependencies(deps ...*Dependency[T]) {
	for _, dep := range deps {
		dd.AddDependency(dep.depFactory, dep.depType)
	}
}

func (dd *DefaultDepDict[T]) Count() int {
	return len(dd.dependencies)
}

func (dd *DefaultDepDict[T]) GetAllDeps() []types.DataType {
	deps := make([]types.DataType, 0, len(dd.dependencies))
	for key, _ := range dd.dependencies {
		deps = append(deps, types.FromKey(key))
	}
	return deps
}

func (dd *DefaultDepDict[T]) ExistDependency(depType types.DataType) bool {
	_, exist := dd.dependencies[depType.Key()]
	return exist
}

func (dd *DefaultDepDict[T]) AddDependency(getInstance T, depType types.DataType) {
	if dd.ExistDependency(depType) {
		panic(fmt.Errorf("dependency type (%v) already exist", depType.FullName()))
	}

	//fmt.Printf("Registered dependency type: %v\n", depType.FullName())
	dd.dependencies[depType.Key()] = getInstance
}

func (dd *DefaultDepDict[T]) GetDependency(depType types.DataType) T {
	factory, exist := dd.dependencies[depType.Key()]
	if !exist {
		panic(fmt.Errorf("dependency not configured, type: %v, quit", depType.FullName()))
	}
	return factory
}
