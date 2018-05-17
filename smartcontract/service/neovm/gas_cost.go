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

package neovm

import (
	vm "github.com/ontio/ontology/vm/neovm"
)

func StoreGasCost(engine *vm.ExecutionEngine) uint64 {
	key := vm.PeekNByteArray(0, engine)
	value := vm.PeekNByteArray(1, engine)
	return uint64(((len(key)+len(value)-1)/1024 + 1)) * GAS_TABLE[STORAGE_PUT_NAME]
}

func GasPrice(engine *vm.ExecutionEngine, name string) uint64 {
	switch name {
	case STORAGE_PUT_NAME:
		return StoreGasCost(engine)
	default:
		if value, ok := GAS_TABLE[name]; ok {
			return value
		}
		return OPCODE_GAS
	}
}
