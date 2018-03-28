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

package exec

import (
	"math"
)

// these operations are essentially no-ops.
// TODO(vibhavp): Add optimisations to package compiles that
// removes them from the original bytecode.

func (vm *VM) i32ReinterpretF32() {
	vm.pushUint32(math.Float32bits(vm.popFloat32()))
}

func (vm *VM) i64ReinterpretF64() {
	vm.pushUint64(math.Float64bits(vm.popFloat64()))
}

func (vm *VM) f32ReinterpretI32() {
	vm.pushFloat32(math.Float32frombits(vm.popUint32()))
}

func (vm *VM) f64ReinterpretI64() {
	vm.pushFloat64(math.Float64frombits(vm.popUint64()))
}
