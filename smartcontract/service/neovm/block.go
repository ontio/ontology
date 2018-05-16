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
	"github.com/ontio/ontology/core/types"
	vm "github.com/ontio/ontology/vm/neovm"
	vmtypes "github.com/ontio/ontology/vm/neovm/types"
)

// BlockGetTransactionCount put block's transactions count to vm stack
func BlockGetTransactionCount(service *NeoVmService, engine *vm.ExecutionEngine) error {
	vm.PushData(engine, len(vm.PopInteropInterface(engine).(*types.Block).Transactions))
	return nil
}

// BlockGetTransactions put block's transactions to vm stack
func BlockGetTransactions(service *NeoVmService, engine *vm.ExecutionEngine) error {
	transactions := vm.PopInteropInterface(engine).(*types.Block).Transactions
	transactionList := make([]vmtypes.StackItems, 0)
	for _, v := range transactions {
		transactionList = append(transactionList, vmtypes.NewInteropInterface(v))
	}
	vm.PushData(engine, transactionList)
	return nil
}

// BlockGetTransaction put block's transaction to vm stack
func BlockGetTransaction(service *NeoVmService, engine *vm.ExecutionEngine) error {
	vm.PushData(engine, vm.PopInteropInterface(engine).(*types.Block).Transactions[vm.PopInt(engine)])
	return nil
}
