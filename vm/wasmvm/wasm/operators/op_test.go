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
