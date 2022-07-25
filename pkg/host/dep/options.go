package dep

import (
	"bytes"
	"fmt"

	"goms.io/azureml/mir/mir-vmagent/pkg/host/types"
)

type TypeConstraint int8

const (
	InterfaceType TypeConstraint = iota
	StructType
	PointerType
	InterfacePtrType
	StructPtrType

	_minType = InterfaceType
	_maxType = StructPtrType
)

func (tc TypeConstraint) String() string {
	switch tc {
	case InterfaceType:
		return "Interface"
	case StructType:
		return "Struct"
	case StructPtrType:
		return "StructPointer"
	case PointerType:
		return "Pointer"
	case InterfacePtrType:
		return "InterfacePointer"
	default:
		return fmt.Sprintf("Constraint(%d)", tc)
	}
}

func matchTypeConstraint(componentType types.DataType, typeConstraint TypeConstraint) bool {
	switch typeConstraint {
	case InterfaceType:
		return componentType.IsInterface()
	case StructType:
		return componentType.IsStruct()
	case PointerType:
		return componentType.IsPtr()
	case InterfacePtrType:
		return componentType.IsPtr() && componentType.ElementType().IsInterface()
	case StructPtrType:
		return componentType.IsPtr() && componentType.ElementType().IsStruct()
	default:
		panic(fmt.Errorf("unexpected type constraint: %v", typeConstraint))
	}
}

type ComponentProviderOptions struct {
	// component types that are allowed to be added into the component collection
	AllowedComponentTypes []TypeConstraint
	// configuration types that are allowed to be added into the collection
	AllowedConfigurationTypes []TypeConstraint
	// allow interface{} as return type of factory method
	AllowTypeAnyFromFactoryMethod bool

	// enable printing diagnostic info
	EnableDiagnostics bool
	// enable concurrency control during creating singleton instance
	EnableSingletonConcurrency bool
	// enable tracking recurrence when creating transient component instances
	TrackTransientRecurrence bool
	// max allowed recurrence when creating transient component instances
	MaxAllowedRecurrence uint32

	// enable properties pass-over
	PropertiesPassOver bool
}

func NewComponentProviderOptions(allowedComponentTypes ...TypeConstraint) *ComponentProviderOptions {
	options := &ComponentProviderOptions{
		AllowedComponentTypes:         make([]TypeConstraint, 0),
		AllowedConfigurationTypes:     make([]TypeConstraint, 0),
		AllowTypeAnyFromFactoryMethod: false,
		EnableDiagnostics:             false,
		EnableSingletonConcurrency:    true,
		TrackTransientRecurrence:      false,
		MaxAllowedRecurrence:          2,
		PropertiesPassOver:            false,
	}
	options.AllowedComponentTypes = append(options.AllowedComponentTypes, allowedComponentTypes...)
	options.AllowedConfigurationTypes = append(options.AllowedConfigurationTypes, StructType)
	return options
}

func (cpo *ComponentProviderOptions) ToString(constraints []TypeConstraint) string {
	var csv bytes.Buffer
	for index, item := range constraints {
		csv.WriteString(item.String())
		if index < (len(constraints) - 1) {
			csv.WriteString(",")
		}
	}
	return csv.String()
}

func (cpo *ComponentProviderOptions) ValidateConfigurationTypeAllowed(configType types.DataType) {
	for _, allowedType := range cpo.AllowedConfigurationTypes {
		if matchTypeConstraint(configType, allowedType) {
			return
		}
	}
	panic(fmt.Errorf("configuration type not allowed: %v, allowed types: %v", configType.FullName(), cpo.ToString(cpo.AllowedConfigurationTypes)))
}
func (cpo *ComponentProviderOptions) ValidateComponentTypeAllowed(componentType types.DataType) {
	for _, allowedType := range cpo.AllowedComponentTypes {
		if matchTypeConstraint(componentType, allowedType) {
			return
		}
	}
	panic(fmt.Errorf("component type not allowed: %v, allowed types: %v", componentType.FullName(), cpo.ToString(cpo.AllowedComponentTypes)))
}
