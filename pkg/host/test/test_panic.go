package test

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
)

func AssertPanic(t *testing.T, errorMsg string) {
	if r := recover(); r != nil {
		t.Logf("panic raised as expected: %v", r)
	} else {
		t.Error(errorMsg)
	}
}

func AssertPanicContent(t *testing.T, contentSubStr string, errorMsg string) {
	if r := recover(); r != nil {
		content := fmt.Sprintf("%v", r)
		fmt.Printf("panic raised: %s\n", content)
		if !strings.Contains(content, contentSubStr) {
			t.Errorf("%s, should contains substring: %s", errorMsg, contentSubStr)

			printStack()
		} else {
			t.Logf("panic raised with expected content substring: %s", contentSubStr)
		}
	} else {
		t.Errorf("expect panic being raised with content: %s", contentSubStr)
	}
}

func AssertNoPanic(t *testing.T, errorMsg string) {
	if r := recover(); r == nil {
		t.Log("no panic was raised, expected")
	} else {
		content := fmt.Sprintf("%v", r)
		t.Logf("panic raised: %s", content)
		t.Error(errorMsg)

		printStack()
	}
}

func printStack() {
	var buf [4096]byte
	n := runtime.Stack(buf[:], false)
	fmt.Println(string(buf[:n]))
}
