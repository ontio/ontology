// Copyright 2017 The go-interpreter Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package validate

import (
	"github.com/Ontology/vm/wasmvm/wasm"
)

type operand struct {
	Type wasm.ValueType
}
