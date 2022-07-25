package hosting

import (
	"fmt"
	"testing"
)

type AnotherInterface interface {
	Another()
}

type AnotherStruct struct {
	value int
}

func NewAnotherStruct() *AnotherStruct {
	return &AnotherStruct{
		value: 0,
	}
}

func (as *AnotherStruct) Another() {
	fmt.Println("Another", as.value)
}

//misc tests
func Test_RunningMode(t *testing.T) {
	modes := []RunningMode{Debug, Release, RunningMode(2), RunningMode(3)}
	names := []string{"Debug", "Release", "RunningMode(2)", "RunningMode(3)"}
	for index, mode := range modes {
		name := mode.String()
		if names[index] != name {
			t.Errorf("RunningMode name not expected: %s", name)
		}
	}
}

func Test_EventType(t *testing.T) {
	types := []EventType{EVENT_TYPE_SIGNAL, EVENT_TYPE_WINSVC, EventType(2), EventType(3)}
	names := []string{"Signal", "WinSvc", "EventType(2)", "EventType(3)"}
	for index, ty := range types {
		name := ty.String()
		if names[index] != name {
			t.Errorf("EventType name not expected: %s", name)
		}
	}
}
