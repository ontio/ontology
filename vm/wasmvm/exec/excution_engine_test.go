package exec

import "testing"

/***
 * execution engine testing in engine_test.go
 */

func TestStackPush(t *testing.T) {
	vs := newStack(0)
	err := vs.push(&VM{})
	if err == nil {
		t.Error("empty stack should raise error while pushing")
	}

	vs = newStack(1)
	err = vs.push(&VM{})
	if err != nil {
		t.Error("Push should be succeed!")
	}
	err = vs.push(&VM{})
	if err == nil {
		t.Error("should raise error while pushing")
	}

	vs = newStack(10)
	for i := 0; i < 10; i++ {
		err = vs.push(&VM{})
		if err != nil {
			t.Error("Push should be succeed!")
		}
	}
	err = vs.push(&VM{})
	if err == nil {
		t.Error("should raise error while pushing")
	}
}

func TestStackPop(t *testing.T) {

	vs := newStack(1)
	_, err := vs.pop()
	if err == nil {
		t.Error("empty stack should raise error while poping")
	}

	vs.push(&VM{})
	_, err = vs.pop()
	if err != nil {
		t.Error("pop should be succeed!")
	}
	_, err = vs.pop()
	if err == nil {
		t.Error("empty stack should raise error while poping")
	}

	vs = newStack(10)
	for i := 0; i < 10; i++ {
		err = vs.push(&VM{})
		if err != nil {
			t.Error("Push should be succeed!")
		}
	}

	for i := 0; i < 10; i++ {
		_, err = vs.pop()
		if err != nil {
			t.Error("Pop should be succeed!")
		}
	}
	_, err = vs.pop()
	if err == nil {
		t.Error("empty stack should raise error while poping")
	}
}
