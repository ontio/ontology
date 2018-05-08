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
