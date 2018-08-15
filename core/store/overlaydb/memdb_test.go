// Copyright (c) 2014, Suryandaru Triandana <syndtr@gmail.com>
// All rights reserved.
//
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package overlaydb

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIter(t *testing.T) {
	db := NewMemDB(0)
	db.Put([]byte("aaa"), []byte("bbb"))
	iter := db.NewIterator(nil)
	assert.Equal(t, iter.First(), true)
	assert.Equal(t, iter.Last(), true)
	db.Delete([]byte("aaa"))
	assert.Equal(t, iter.First(), false)
	assert.Equal(t, iter.Last(), false)
}
