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

func (vm *VM) i32Wrapi64() {
	vm.pushUint32(uint32(vm.popUint64()))
}

func (vm *VM) i32TruncSF32() {
	vm.pushInt32(int32(math.Trunc(float64(vm.popFloat32()))))
}

func (vm *VM) i32TruncUF32() {
	vm.pushUint32(uint32(math.Trunc(float64(vm.popFloat32()))))
}

func (vm *VM) i32TruncSF64() {
	vm.pushInt32(int32(math.Trunc(vm.popFloat64())))
}

func (vm *VM) i32TruncUF64() {
	vm.pushUint32(uint32(math.Trunc(vm.popFloat64())))
}

func (vm *VM) i64ExtendSI32() {
	vm.pushInt64(int64(vm.popInt32()))
}

func (vm *VM) i64ExtendUI32() {
	vm.pushUint64(uint64(vm.popUint32()))
}

func (vm *VM) i64TruncSF32() {
	vm.pushInt64(int64(math.Trunc(float64(vm.popFloat32()))))
}

func (vm *VM) i64TruncUF32() {
	vm.pushUint64(uint64(math.Trunc(float64(vm.popFloat32()))))
}

func (vm *VM) i64TruncSF64() {
	vm.pushInt64(int64(math.Trunc(vm.popFloat64())))
}

func (vm *VM) i64TruncUF64() {
	vm.pushUint64(uint64(math.Trunc(vm.popFloat64())))
}

func (vm *VM) f32ConvertSI32() {
	vm.pushFloat32(float32(vm.popInt32()))
}

func (vm *VM) f32ConvertUI32() {
	vm.pushFloat32(float32(vm.popUint32()))
}

func (vm *VM) f32ConvertSI64() {
	vm.pushFloat32(float32(vm.popInt64()))
}

func (vm *VM) f32ConvertUI64() {
	vm.pushFloat32(float32(vm.popUint64()))
}

func (vm *VM) f32DemoteF64() {
	vm.pushFloat32(float32(vm.popFloat64()))
}

func (vm *VM) f64ConvertSI32() {
	vm.pushFloat64(float64(vm.popInt32()))
}

func (vm *VM) f64ConvertUI32() {
	vm.pushFloat64(float64(vm.popUint32()))
}

func (vm *VM) f64ConvertSI64() {
	vm.pushFloat64(float64(vm.popInt64()))
}

func (vm *VM) f64ConvertUI64() {
	vm.pushFloat64(float64(vm.popUint64()))
}

func (vm *VM) f64PromoteF32() {
	vm.pushFloat64(float64(vm.popFloat32()))
}
