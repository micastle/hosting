package types

import (
	"fmt"
	"reflect"
)

type FuncType interface {
	Key() interface{}

	FullName() string
	Pkg() string
	Name() string

	GetNumOfOutput() int
	GetOutput(index int) DataType

	GetOutputType() DataType
}

type DefaultFuncType struct {
	rawType reflect.Type
}

func newFuncType(rawType reflect.Type) *DefaultFuncType {
	return &DefaultFuncType{
		rawType: rawType,
	}
}

func (dt *DefaultFuncType) Key() interface{} {
	return dt.rawType
}

func (dt *DefaultFuncType) FullName() string {
	return dt.rawType.String()
}

func (dt *DefaultFuncType) Pkg() string {
	return dt.rawType.PkgPath()
}

func (dt *DefaultFuncType) Name() string {
	return dt.rawType.Name()
}

func (ft *DefaultFuncType) GetNumOfOutput() int {
	return ft.rawType.NumOut()
}
func (ft *DefaultFuncType) GetOutputType() DataType {
	return ft.GetOutput(0)
}
func (ft *DefaultFuncType) GetOutput(index int) DataType {
	if ft.rawType.NumOut() <= index {
		panic(fmt.Errorf("output index %d exceeded number of func outputs %d, %s", index, ft.rawType.NumOut(), ft.rawType.String()))
	}
	return from(ft.rawType.Out(index))
}

type InputProvider func(index int, argType DataType) interface{}

type Func interface {
	GetType() FuncType

	Call(inputs InputProvider) []interface{}
}

type DefaultFunc struct {
	rawType  reflect.Type
	instance interface{}
}

func newFunc(rawType reflect.Type, instance interface{}) *DefaultFunc {
	return &DefaultFunc{
		rawType:  rawType,
		instance: instance,
	}
}
func (f *DefaultFunc) GetType() FuncType {
	return newFuncType(f.rawType)
}

func (f *DefaultFunc) Call(getArg InputProvider) []interface{} {
	funcValue := reflect.ValueOf(f.instance)

	// get args for the function call
	inputs := make([]reflect.Value, 0, f.rawType.NumIn())
	for i := 0; i < f.rawType.NumIn(); i++ {
		arg := getArg(i, from(f.rawType.In(i)))
		inputs = append(inputs, reflect.ValueOf(arg))
	}

	// make call
	outputs := funcValue.Call(inputs)

	// return the outputs
	results := make([]interface{}, 0, len(outputs))
	for _, out := range outputs {
		results = append(results, out.Interface())
	}
	return results
}
