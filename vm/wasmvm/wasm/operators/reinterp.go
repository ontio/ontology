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
	I32ReinterpretF32 = newOp(0xbc, "i32.reinterpret/f32", []wasm.ValueType{wasm.ValueTypeF32}, wasm.ValueTypeI32)
	I64ReinterpretF64 = newOp(0xbd, "i64.reinterpret/f64", []wasm.ValueType{wasm.ValueTypeF64}, wasm.ValueTypeI64)
	F32ReinterpretI32 = newOp(0xbe, "f32.reinterpret/i32", []wasm.ValueType{wasm.ValueTypeI32}, wasm.ValueTypeF32)
	F64ReinterpretI64 = newOp(0xbf, "f64.reinterpret/i64", []wasm.ValueType{wasm.ValueTypeI64}, wasm.ValueTypeF64)
)
