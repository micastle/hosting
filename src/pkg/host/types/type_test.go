package types

import (
	"fmt"
	"testing"

	"goms.io/azureml/mir/mir-vmagent/pkg/host/test"
)

type TestStruct struct {
	value int
}

func (ts *TestStruct) Print() {

}

type TestInterface interface {
	Print()
}

type StructFactory func (int) *TestStruct

func FuncToBeTest(value int) *TestStruct {
	return &TestStruct{value: value}
}

const types_pkg_path string = "goms.io/azureml/mir/mir-vmagent/pkg/host/types"

type GenericObjectCreator func() interface{}
type TestMethod interface {
    GenericObjectCreator
}

type MyInterface[T any] interface {
	Register(val T)
}
type MyStruct[T TestMethod] struct{
	val T
}

func (s *MyStruct[T]) Register(val int) {

}

func CreateStruct[T GenericObjectCreator](val T) *MyStruct[T] {
	return &MyStruct[T]{val : val}
}

func CreateInterface[T GenericObjectCreator](val T) MyInterface[int] {
	return &MyStruct[T]{val : val}
}

func MyCreateStruct() interface{} {
	return nil
}
func TestStructGenerics(t *testing.T) {
	intStruct := CreateStruct[GenericObjectCreator](MyCreateStruct)
	if intStruct.val() != nil {
		t.Error("not expected value")
	}
	intStruct.Register(123)

	intInterface := CreateInterface[GenericObjectCreator](MyCreateStruct)
	intInterface.Register(123)
}

type MyStruct2[T any, S any] struct {
	val1 T
	val2 S
}
func TestStruct2Generics(t *testing.T) {
	s2 := &MyStruct2[int, string]{
		val1: 123,
		val2: "hello",
	}
	if s2.val1 != 123 {
		t.Error("int val not expected")
	}
}

type FactoryMethod[R any] func() R
type Creator1[R any, T any] func(T) R
type Creator2[R any, T any, S any] func(T, S) R
type Creator3[R any, T any, S any, U any] func(T, S, U) R
type Creator4[R any, T any, S any, U any, V any] func(T, S, U, V) R

type FreeStyle1[R any, T any] interface {
	FactoryMethod[R] | Creator1[R, T]
}

type FreeStyleFactoryMethod[R any, T any, S any, U any, V any, W any, X any, Y any, Z any] interface {
	func() R |
	func(T) R |
	func(T, S) R |
	func(T, S, U) R |
	func(T, S, U, V) R |
	func(T, S, U, V, W) R |
	func(T, S, U, V, W, X) R |
	func(T, S, U, V, W, X, Y) R |
	func(T, S, U, V, W, X, Y, Z) R
	//FactoryMethod | Creator1[T] | Creator2[T,S] | Creator3[T, S, U] | Creator4[T, S, U, V]
}

func Create(val int, str string) *TestStruct {
	return nil
}

func RegisterSingleton[T any, S any, Func FreeStyleFactoryMethod[*TestStruct, T, S, any, any, any, any, any, any]](createInstance Func) {

}

func TestFreeStyleFactoryMethod(t *testing.T) {
	RegisterSingleton[int, string](Create)
}

func ToInterface(ptr *TestStruct) TestInterface {
	return ptr
}
func TestStructType(t *testing.T) {
	t_struct := &TestStruct{}
	t_struct.Print()

	structType := Of(t_struct)
	structType2 := Of(t_struct)
	if structType.Key() != structType2.Key() {
		t.Errorf("data type of struct pointer does not match, %v != %v", structType.FullName(), structType2.FullName())
	}

	if !structType.IsPtr() || !structType.ElementType().IsStruct() {
		t.Errorf("data type of struct pointer is not PTR or element type is not STRUCT, %v", structType.FullName())
	}

	if structType.Name() != "TestStruct" {
		t.Errorf("name of struct pointer data type is not expected, %v", structType.Name())
	}

	if structType.FullName() != "*types.TestStruct" {
		t.Errorf("full name of struct pointer data type is not expected, %v", structType.FullName())
	}

	if structType.Pkg() != types_pkg_path {
		t.Errorf("package of struct pointer data type is not expected, %v", structType.Pkg())
	}

	struct_name := Name(t_struct)
	if struct_name != structType.Name() {
		t.Errorf("struct instance name(%v) does not match struct type name: %v", struct_name, structType.Name())
	}

	t_interface := ToInterface(t_struct)

	interfaceType := Of(t_interface)
	interfaceType2 := Of(t_interface)
	if interfaceType.Key() != interfaceType2.Key() {
		t.Errorf("data type of interface does not match, %v != %v", interfaceType.FullName(), interfaceType2.FullName())
	}

	if interfaceType.IsInterface() {
		t.Errorf("data type of interface actually from struct pointer should not be INTERFACE, %v", interfaceType.FullName())
	}

	if interfaceType.Key() != structType.Key() {
		t.Errorf("data type of interface actually from struct pointer does not match original type, %v != %v", interfaceType.FullName(), structType.FullName())
	}

	if !interfaceType.IsPtr() || !interfaceType.ElementType().IsStruct() {
		t.Errorf("data type of struct pointer is not PTR or element type is not STRUCT, %v", structType.FullName())
	}

	if interfaceType.Name() != "TestStruct" {
		t.Errorf("name of struct pointer data type is not expected, %v", interfaceType.Name())
	}

	if interfaceType.FullName() != "*types.TestStruct" {
		t.Errorf("full name of struct pointer data type is not expected, %v", interfaceType.FullName())
	}

	if interfaceType.Pkg() != types_pkg_path {
		t.Errorf("package of struct pointer data type is not expected, %v", interfaceType.Pkg())
	}

	if t_interface == nil {
		t.Errorf("interface actually from struct pointer should not be nil")
	} else {
		t_interface.Print()
	}

	interface_name := Name(t_interface)
	if interface_name != interfaceType.Name() {
		t.Errorf("interface instance name(%v) does not match interface type name: %v", interface_name, interfaceType.Name())
	}
}

func TestInterfaceType(t *testing.T) {

	t_interface := new(TestInterface)

	interfacePtr := Of(t_interface)
	interfacePtr2 := Of(t_interface)
	if interfacePtr.Key() != interfacePtr2.Key() {
		t.Errorf("data type of interface pointer does not match, %v != %v", interfacePtr.FullName(), interfacePtr2.FullName())
	}

	if !interfacePtr.IsPtr() {
		t.Errorf("data type of interface pointer is not PTR, %v", interfacePtr.FullName())
	}

	interfaceType := interfacePtr.ElementType()
	if !interfaceType.IsInterface() {
		t.Errorf("data type of interface from empty is not INTERFACE, %v", interfaceType.FullName())
	}

	if interfaceType.Name() != "TestInterface" {
		t.Errorf("name of interface data type is not expected, %v", interfaceType.Name())
	}

	if interfaceType.FullName() != "types.TestInterface" {
		t.Errorf("full name of interface data type is not expected, %v", interfaceType.FullName())
	}

	if interfaceType.Pkg() != types_pkg_path {
		t.Errorf("package of interface data type is not expected, %v", interfaceType.Pkg())
	}

	if *t_interface != nil {
		(*t_interface).Print()
		t.Errorf("dereference result of interface pointer from empty is not nil")
	}
}

func TestGetType(t *testing.T) {
	interfaceType := Get[TestInterface]()
	if !interfaceType.IsInterface() {
		t.Errorf("get interface type is not actually interface: %s", interfaceType.FullName())
	}
	interfacePtrType := Get[*TestInterface]()
	if !interfacePtrType.IsPtr() {
		t.Errorf("get interface ptr type is not actually ptr: %s", interfacePtrType.FullName())
	} else {
		elementType := interfacePtrType.ElementType()
		if !elementType.IsInterface() {
			t.Errorf("element type of interface ptr type is not actually interface: %s", elementType.FullName())
		}
	}
	structType := Get[TestStruct]()
	if !structType.IsStruct() {
		t.Errorf("get struct type is not actually struct: %s", structType.FullName())
	}
	structPtrType := Get[*TestStruct]()
	if !structPtrType.IsPtr() {
		t.Errorf("get struct ptr type is not actually ptr: %s", structPtrType.FullName())
	} else {
		elementType := structPtrType.ElementType()
		if !elementType.IsStruct() {
			t.Errorf("element type of struct ptr type is not actually struct: %s", elementType.FullName())
		}
	}

	funcType := Get[StructFactory]()
	if !funcType.IsFunc() {
		t.Errorf("get func type is not actually func: %s", funcType.FullName())
	}
}

func TestTypeKey(t *testing.T) {
	t_struct := &TestStruct{}
	dt := Of(t_struct)
	key := dt.Key()
	dt2 := FromKey(key)
	if !dt2.IsPtr() || !dt2.ElementType().IsStruct() || dt2.FullName() != dt.FullName() {
		t.Errorf("data type from key of a struct pointer type is not expected: %v", dt2.FullName())
	}
}

func TestTypeRelations(t *testing.T) {
	// struct types
	t_struct := &TestStruct{}
	nil_interface := new(TestInterface)

	structPtrType := Of(t_struct)
	if !structPtrType.IsPtr() {
		t.Errorf("data type of struct pointer is not PTR, %v", structPtrType.FullName())
	}
	structType := structPtrType.ElementType()
	if !structType.IsStruct() {
		t.Errorf("element type of struct pointer is not STRUCT, %v", structType.FullName())
	}
	ptrStructType := structType.PointerType()
	if !ptrStructType.IsPtr() || !ptrStructType.ElementType().IsStruct() || ptrStructType.Key() != structPtrType.Key() {
		t.Errorf("pointer type of struct does not match data type of struct pointer, %v != %v", ptrStructType.FullName(), structPtrType.FullName())
	}

	// interface types
	interfacePtrType := Of(nil_interface)
	if !interfacePtrType.IsPtr() {
		t.Errorf("data type of nil interface is not PTR, %v", interfacePtrType.FullName())
	}
	interfaceType := interfacePtrType.ElementType()
	if !interfaceType.IsInterface() {
		t.Errorf("element type if nil interface is not INTERFACE, %v", interfaceType.FullName())
	}
	ptrInterfaceType := interfaceType.PointerType()
	if !ptrInterfaceType.IsPtr() || !ptrInterfaceType.ElementType().IsInterface() || ptrInterfaceType.Key() != interfacePtrType.Key() {
		t.Errorf("pointer type of interface does not match data type of interface pointer, %v != %v", ptrInterfaceType.FullName(), interfacePtrType.FullName())
	}

	// type compatibility
	if structType.CheckCompatible(interfaceType) {
		t.Errorf("struct type should not be compatible with interface type")
	}
	if !structPtrType.CheckCompatible(interfaceType) {
		t.Errorf("struct pointer type does not compatible with interface type")
	}
	if !interfaceType.CheckCompatible(interfaceType) {
		t.Errorf("same interface type should be compatible with each other")
	}

	if structType.CheckCompatible(interfacePtrType) {
		t.Errorf("struct type should not be compatible with interface pointer type")
	}
	if !structPtrType.CheckCompatible(interfacePtrType) {
		t.Errorf("struct pointer type does not compatible with interface pointer type")
	}
	if !interfaceType.CheckCompatible(interfacePtrType) {
		t.Errorf("interface type should be compatible with interface pointer type")
	}

	if !structType.CheckCompatible(structType) {
		t.Errorf("same struct type should be compatible with each other")
	}
	if structPtrType.CheckCompatible(structType) {
		t.Errorf("struct pointer type should not be compatible with struct type")
	}
	if interfaceType.CheckCompatible(structType) {
		t.Errorf("interface type should not be compatible with struct type")
	}

	if !structType.CheckCompatible(structPtrType) {
		t.Errorf("same struct type should be compatible with struct pointer type")
	}
	if !structPtrType.CheckCompatible(structPtrType) {
		t.Errorf("struct pointer type should be compatible with struct pointer type")
	}
	if interfaceType.CheckCompatible(structPtrType) {
		t.Errorf("interface type should not be compatible with struct pointer type")
	}

	// type any
	if structType.IsAny() {
		t.Errorf("struct type should not be any")
	}
	if structPtrType.IsAny() {
		t.Errorf("struct pointer type should not be any")
	}
	if interfaceType.IsAny() {
		t.Errorf("interface type should not be any")
	}
	if interfacePtrType.IsAny() {
		t.Errorf("interface pointer type should not be any")
	}
	anyPtrType := Of(new(interface{}))
	anyType := anyPtrType.ElementType()
	if !anyType.IsAny() {
		t.Errorf("type any is not ANY, %v", anyType.FullName())
	}

	if anyType.IsPtr() {
		t.Errorf("ANY type should not be a PTR")
	}
	if !anyType.IsInterface() {
		t.Errorf("ANY type should be a INTERFACE")
	}
	if anyType.IsStruct() {
		t.Errorf("ANY type should be a STRUCT")
	}
	if !anyType.CheckCompatible(anyType) {
		t.Errorf("any type should be compatible with type ANY")
	}
	if !interfaceType.CheckCompatible(anyType) {
		t.Errorf("interface type should be compatible with type ANY")
	}
	if !structType.CheckCompatible(anyType) {
		t.Errorf("struct type should be compatible with type ANY")
	}
	if !interfacePtrType.CheckCompatible(anyType) {
		t.Errorf("interface pointer type should be compatible with type ANY")
	}
	if !structPtrType.CheckCompatible(anyType) {
		t.Errorf("struct pointer type should be compatible with type ANY")
	}
}

func TestTypeAPI_negative(t *testing.T) {
	defer test.AssertPanicContent(t, "nil type specified, instance is nil", "panic content is not expected")

	var inst TestInterface = nil

	ty := Of(inst)
	if ty == nil {
		t.Error("type of nil interface should not return nil, raise panic instead")
	}
}

func TestFuncType(t *testing.T) {
	funcType := GetFuncType(FuncToBeTest)
	if funcType.Name() != "" {
		t.Errorf("name of func type is not expected, %v", funcType.Name())
	}
	if funcType.Pkg() != "" {
		t.Errorf("pkg of func type is not expected, %v", funcType.Pkg())
	}
	if funcType.FullName() != "func(int) *types.TestStruct" {
		t.Errorf("full func type name is not expected, %v", funcType.FullName())
	}

	outNum := funcType.GetNumOfOutput()
	fmt.Printf("output count: %v\n", outNum)

	outType := funcType.GetOutputType()
	if outType.Key() != Of(new(TestStruct)).Key() {
		t.Errorf("output type of func type is not expected, %v", outType.FullName())
	}

	funcInst := ToFunc(FuncToBeTest)
	if funcInst.GetType().Key() != funcType.Key() {
		t.Errorf("func type does not match, %v != %v", funcInst.GetType().FullName(), funcType.FullName())
	}

	result := funcInst.Call(func(index int, argType DataType) interface{} { return int(123) })
	if len(result) != 1 {
		t.Errorf("output count should be 1, actual - %v", len(result))
	}

	output := result[0].(*TestStruct)
	if output == nil {
		t.Error("output should be not nil")
	} else if output.value != 123 {
		t.Errorf("output value is not expected: %d", output.value)
	}
}

func FuncMultiOutput(value int) (*TestStruct, int) {
	return &TestStruct{value: value}, value
}
func TestFuncOutputType_negative(t *testing.T) {
	defer test.AssertPanicContent(t, "number of func outputs is not 1", "panic content is not expected")

	funcType := GetFuncType(FuncMultiOutput)
	ty := funcType.GetOutputType()
	if ty == nil {
		t.Error("func output type for FuncMultiOutput should not be nil")
	}
}

func TestFuncTypeAPI_negative(t *testing.T) {
	defer test.AssertPanicContent(t, "arg func instance is nil", "panic content is not expected")

	var inst interface{} = nil

	ty := GetFuncType(inst)
	if ty == nil {
		t.Error("type of nil interface should not return nil, raise panic instead")
	}
}

func TestFuncTypeAPI_negative2(t *testing.T) {
	defer test.AssertPanicContent(t, "type of argument instance is not func", "panic content is not expected")

	inst := &TestStruct{}

	ty := GetFuncType(inst)
	if ty == nil {
		t.Error("type of nil interface should not return nil, raise panic instead")
	}
}
