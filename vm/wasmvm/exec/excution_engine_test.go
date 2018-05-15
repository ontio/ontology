/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */
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
