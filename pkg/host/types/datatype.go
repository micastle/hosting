package types

import (
	"reflect"
)

type DataType interface {
	Key() interface{}

	FullName() string
	Pkg() string
	Name() string
	IsAny() bool
	IsInterface() bool
	IsStruct() bool
	IsPtr() bool
	IsFunc() bool

	ElementType() DataType
	PointerType() DataType

	CheckCompatible(interfaceType DataType) bool
}

type DefaultDataType struct {
	rawType reflect.Type
}

// get the key for data type, for compare and indexing purpose
func (dt *DefaultDataType) Key() interface{} {
	return dt.rawType
}

// hosting.DataType
func (dt *DefaultDataType) FullName() string {
	return dt.rawType.String()
}

// goms.io/azureml/mir/mir-vmagent/pkg/hosting
func (dt *DefaultDataType) Pkg() string {
	if dt.IsPtr() {
		return dt.ElementType().Pkg()
	}
	return dt.rawType.PkgPath()
}

// DataType
func (dt *DefaultDataType) Name() string {
	if dt.IsPtr() {
		return dt.ElementType().Name()
	}
	return dt.rawType.Name()
}

func (dt *DefaultDataType) IsAny() bool {
	if dt.rawType == reflect.TypeOf(new(interface{})).Elem() {
		return true
	}
	return false
}
func (dt *DefaultDataType) IsInterface() bool {
	return dt.rawType.Kind() == reflect.Interface
}

func (dt *DefaultDataType) IsStruct() bool {
	return dt.rawType.Kind() == reflect.Struct
}

func (dt *DefaultDataType) IsPtr() bool {
	return dt.rawType.Kind() == reflect.Ptr
}

func (dt *DefaultDataType) IsFunc() bool {
	return dt.rawType.Kind() == reflect.Func
}

func (dt *DefaultDataType) ElementType() DataType {
	return &DefaultDataType{
		rawType: dt.rawType.Elem(),
	}
}

func (dt *DefaultDataType) PointerType() DataType {
	return &DefaultDataType{
		rawType: reflect.PtrTo(dt.rawType),
	}
}

func (dt *DefaultDataType) CheckCompatible(interfaceType DataType) bool {
	return checkCompatible(dt.rawType, interfaceType.(*DefaultDataType).rawType)
}

func checkCompatible(instanceType reflect.Type, componentType reflect.Type) bool {
	if instanceType == componentType {
		return true
	}

	actualType := componentType
	if componentType.Kind() == reflect.Ptr {
		actualType = componentType.Elem()
	}

	if actualType.Kind() == reflect.Interface {
		if instanceType.Implements(actualType) {
			return true
		}
	} else {
		if instanceType == actualType {
			return true
		}
	}

	return false
}
