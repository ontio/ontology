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

func StoreGasCost(gasTable map[string]uint64, engine *vm.Executor) (uint64, error) {
	key, err := engine.EvalStack.PeekAsBytes(1)
	if err != nil {
		return 0, err
	}
	value, err := engine.EvalStack.PeekAsBytes(2)
	if err != nil {
		return 0, err
	}
	if putCost, ok := gasTable[STORAGE_PUT_NAME]; ok {
		return uint64((len(key)+len(value)-1)/1024+1) * putCost, nil
	} else {
		return uint64(0), errors.NewErr("[StoreGasCost] get STORAGE_PUT_NAME gas failed")
	}
}

func GasPrice(gasTable map[string]uint64, engine *vm.Executor, name string) (uint64, error) {
	switch name {
	case STORAGE_PUT_NAME:
		return StoreGasCost(gasTable, engine)
	default:
		if value, ok := gasTable[name]; ok {
			return value, nil
		}
		return OPCODE_GAS, nil
	}
}
