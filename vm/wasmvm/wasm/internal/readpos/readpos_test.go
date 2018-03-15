// Copyright 2017 The go-interpreter Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package readpos_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/Ontology/vm/wasmvm/wasm/internal/readpos"
)

func TestRead(t *testing.T) {
	data := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	for i, test := range []struct {
		r    io.Reader
		data []byte
		want int
		err  error
	}{
		{
			r:    bytes.NewReader(data),
			data: nil,
			want: 0,
			err:  nil,
		},
		{
			r:    bytes.NewReader(nil),
			data: nil,
			want: 0,
			err:  io.EOF,
		},
		{
			r:    bytes.NewReader(nil),
			data: make([]byte, 2),
			want: 0,
			err:  io.EOF,
		},
		{
			r:    bytes.NewReader(data),
			data: data,
			want: len(data),
			err:  nil,
		},
		{
			r:    bytes.NewReader(data[:1]),
			data: make([]byte, 2),
			want: 1,
			err:  nil,
		},
	} {
		r := readpos.ReadPos{R: test.r}
		n, err := r.Read(test.data)
		switch {
		case err != test.err:
			t.Errorf("test-#%d: got err=%v. want=%v", i, err, test.err)
			continue
		case n != test.want:
			t.Errorf("test-#%d: got n=%v. want=%v", i, n, test.want)
		case int(r.CurPos) != test.want:
			t.Errorf("test-#%d: got pos=%v. want=%v", i, r.CurPos, test.want)
		}
	}
}
