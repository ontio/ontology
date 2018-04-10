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

package leb128

import (
	"bytes"
	"testing"
)

func TestReadVarUint32(t *testing.T) {
	n, err := ReadVarUint32(bytes.NewReader([]byte{0x80, 0x7f}))
	if err != nil {
		t.Fatal(err)
	}
	if n != uint32(16256) {
		t.Fatalf("got = %d; want = %d", n, 16256)
	}

}

func TestReadVarint32(t *testing.T) {
	n, err := ReadVarint32(bytes.NewReader([]byte{0xFF, 0x7e}))
	if err != nil {
		t.Fatal(err)
	}
	if n != int32(-129) {
		t.Fatalf("got = %d; want = %d", n, -129)
	}
}
