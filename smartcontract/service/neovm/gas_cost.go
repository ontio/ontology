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
	"github.com/ontio/ontology/errors"
	vm "github.com/ontio/ontology/vm/neovm"
)

func StoreGasCost(engine *vm.ExecutionEngine) (uint64, error) {
	key, err := vm.PeekNByteArray(1, engine)
	if err != nil {
		return 0, err
	}
	value, err := vm.PeekNByteArray(2, engine)
	if err != nil {
		return 0, err
	}
	if putCost, ok := GAS_TABLE.Load(STORAGE_PUT_NAME); ok {
		return uint64(((len(key)+len(value)-1)/1024 + 1)) * putCost.(uint64), nil
	} else {
		return uint64(0), errors.NewErr("[StoreGasCost] get STORAGE_PUT_NAME gas failed")
	}
}

func GasPrice(engine *vm.ExecutionEngine, name string) (uint64, error) {
	switch name {
	case STORAGE_PUT_NAME:
		return StoreGasCost(engine)
	default:
		if value, ok := GAS_TABLE.Load(name); ok {
			return value.(uint64), nil
		}
		return OPCODE_GAS, nil
	}
}
