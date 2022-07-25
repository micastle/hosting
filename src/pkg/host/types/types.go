package types

import (
	"fmt"
	"reflect"
)

// create data type from raw type by reflect
func from(rawType reflect.Type) *DefaultDataType {
	if rawType == nil {
		panic(fmt.Errorf("nil type specified, instance is nil?"))
	}
	return &DefaultDataType{
		rawType: rawType,
	}
}

func gettype(instance interface{}) *DefaultDataType {
	return from(reflect.TypeOf(instance))
}

// get type of the instance
func Of(instance interface{}) DataType {
	return gettype(instance)
}

// get type name of the instance
func Name(instance interface{}) string {
	return Of(instance).Name()
}

// get type
func Get[T any]() DataType {
	return gettype((*T)(nil)).ElementType()
}

// get type from type key
func FromKey(key interface{}) DataType {
	return from(key.(reflect.Type))
}

// func related
func getAndValidateFuncType(instance interface{}) reflect.Type {
	if instance == nil {
		panic(fmt.Errorf("arg func instance is nil"))
	}

	instanceType := gettype(instance)
	if !instanceType.IsFunc() {
		panic(fmt.Errorf("type of argument instance is not func: %v", instanceType.FullName()))
	}

	return instanceType.rawType
}
func GetFuncType(instance interface{}) FuncType {
	return newFuncType(getAndValidateFuncType(instance))
}
func ToFunc(instance interface{}) Func {
	return newFunc(getAndValidateFuncType(instance), instance)
}
