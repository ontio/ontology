package errors

import (
	"errors"
	"fmt"
	"testing"
)

var (
	TestRootError = errors.New("Test Root Error Msg.")
)

func TestNewDetailErr(t *testing.T) {
	e := NewDetailErr(TestRootError, ErrUnknown, "Test New Detail Error")
	if e == nil {
		t.Fatal("NewDetailErr should not return nil.")
	}
	fmt.Println(e.Error())

	msg := CallStacksString(GetCallStacks(e))

	fmt.Println(msg)

	if msg == "" {
		t.Errorf("CallStacksString should not return empty msg.")
	}

	rooterr := RootErr(e)
	fmt.Println("Root: ", rooterr.Error())

	code := ErrerCode(e)
	fmt.Println("Code: ", code.Error())

	fmt.Println("TestNewDetailErr End.")
}
