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
