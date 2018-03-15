// Copyright 2017 The go-interpreter Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package operators

import (
	"testing"
)

func TestNew(t *testing.T) {
	op1, err := New(Unreachable)
	if err != nil {
		t.Fatalf("unexpected error from New: %v", err)
	}
	if op1.Name != "unreachable" {
		t.Fatalf("0x00: unexpected Op name. got=%s, want=unrechable", op1.Name)
	}
	if !op1.IsValid() {
		t.Fatalf("0x00: operator %v is invalid (should be valid)", op1)
	}

	op2, err := New(0xff)
	if err == nil {
		t.Fatalf("0xff: expected error while getting Op value")
	}
	if op2.IsValid() {
		t.Fatalf("0xff: operator %v is valid (should be invalid)", op2)
	}
}
