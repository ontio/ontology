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
