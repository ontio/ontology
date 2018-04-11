package neovm

import (
	vm "github.com/ontio/ontology/vm/neovm"
	"github.com/ontio/ontology/core/types"
	vmtypes "github.com/ontio/ontology/vm/neovm/types"
)

// get transaction count from block
func BlockGetTransactionCount(service *NeoVmService, engine *vm.ExecutionEngine) error {
	vm.PushData(engine, len(vm.PopInteropInterface(engine).(*types.Block).Transactions))
	return nil
}

// get transactions from block
func BlockGetTransactions(service *NeoVmService, engine *vm.ExecutionEngine) error {
	transactions := vm.PopInteropInterface(engine).(*types.Block).Transactions
	transactionList := make([]vmtypes.StackItems, 0)
	for _, v := range transactions {
		transactionList = append(transactionList, vmtypes.NewInteropInterface(v))
	}
	vm.PushData(engine, transactionList)
	return nil
}

// get transaction from block
func BlockGetTransaction(service *NeoVmService, engine *vm.ExecutionEngine) error {
	vm.PushData(engine, vm.PopInteropInterface(engine).(*types.Block).Transactions[vm.PopInt(engine)])
	return nil
}



