// Copyright 2017 The go-interpreter Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package operators

import (
	"reflect"
	"testing"

	"github.com/Ontology/vm/wasmvm/wasm"
)

func TestNewConversionOp(t *testing.T) {
	origOps := ops
	defer func() {
		ops = origOps
	}()

	ops = [256]Op{}
	testCases := []struct {
		name    string
		args    []wasm.ValueType
		returns wasm.ValueType
	}{
		{"i32.wrap/i64", []wasm.ValueType{wasm.ValueTypeI64}, wasm.ValueTypeI32},
		{"i32.trunc_s/f32", []wasm.ValueType{wasm.ValueTypeF32}, wasm.ValueTypeI32},
	}

	for i, testCase := range testCases {
		op, err := New(newConversionOp(byte(i), testCase.name))
		if err != nil {
			t.Fatalf("%s: unexpected error from New: %v", testCase.name, err)
		}

		if !reflect.DeepEqual(op.Args, testCase.args) {
			t.Fatalf("%s: unexpected param types: got=%v, want=%v", testCase.name, op.Args, testCase.args)
		}

		if op.Returns != testCase.returns {
			t.Fatalf("%s: unexpected return type: got=%v, want=%v", testCase.name, op.Returns, testCase.returns)
		}
	}
}
