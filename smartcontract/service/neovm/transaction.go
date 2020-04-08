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
	"fmt"

	"github.com/ontio/ontology/core/types"
	vm "github.com/ontio/ontology/vm/neovm"
	vmtypes "github.com/ontio/ontology/vm/neovm/types"
)

// GetExecutingAddress push transaction's hash to vm stack
func TransactionGetHash(service *NeoVmService, engine *vm.Executor) error {
	txn, err := engine.EvalStack.PopAsInteropValue()
	if err != nil {
		return fmt.Errorf("[TransactionGetHash] PopAsInteropValue error:%s", err)
	}
	if tx, ok := txn.Data.(*types.Transaction); ok {
		txHash := tx.Hash()
		return engine.EvalStack.PushBytes(txHash.ToArray())
	}
	return fmt.Errorf("[TransactionGetHash] Type error")
}

// TransactionGetType push transaction's type to vm stack
func TransactionGetType(service *NeoVmService, engine *vm.Executor) error {
	txn, err := engine.EvalStack.PopAsInteropValue()
	if err != nil {
		return fmt.Errorf("[TransactionGetType] PopAsInteropValue error:%s", err)
	}
	if tx, ok := txn.Data.(*types.Transaction); ok {
		return engine.EvalStack.PushInt64(int64(tx.TxType))
	}
	return fmt.Errorf("[TransactionGetType] Type error")
}

// TransactionGetAttributes push transaction's attributes to vm stack
func TransactionGetAttributes(service *NeoVmService, engine *vm.Executor) error {
	_, err := engine.EvalStack.PopAsInteropValue()
	if err != nil {
		return fmt.Errorf("[TransactionGetAttributes] PopAsInteropValue error: %s", err)
	}
	attributList := make([]vmtypes.VmValue, 0)
	return engine.EvalStack.PushAsArray(attributList)
}
