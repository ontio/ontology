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

func (vm *VM) getLocal() {
	index := vm.fetchUint32()
	vm.pushUint64(vm.ctx.locals[int(index)])
}

func (vm *VM) setLocal() {
	index := vm.fetchUint32()
	vm.ctx.locals[int(index)] = vm.popUint64()
}

func (vm *VM) teeLocal() {
	index := vm.fetchUint32()
	val := vm.ctx.stack[len(vm.ctx.stack)-1]
	vm.ctx.locals[int(index)] = val
}

func (vm *VM) getGlobal() {
	index := vm.fetchUint32()
	vm.pushUint64(vm.globals[int(index)])
}

func (vm *VM) setGlobal() {
	index := vm.fetchUint32()
	vm.globals[int(index)] = vm.popUint64()
}
