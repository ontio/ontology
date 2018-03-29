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
	"github.com/ontio/ontology/vm/wasmvm/wasm"
)

var (
	I32Const = newOp(0x41, "i32.const", nil, wasm.ValueTypeI32)
	I64Const = newOp(0x42, "i64.const", nil, wasm.ValueTypeI64)
	F32Const = newOp(0x43, "f32.const", nil, wasm.ValueTypeF32)
	F64Const = newOp(0x44, "f64.const", nil, wasm.ValueTypeF64)
)
