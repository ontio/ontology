package neovm

import (
	vm "github.com/ontio/ontology/vm/neovm"
	"github.com/ontio/ontology/core/types"
	vmtypes "github.com/ontio/ontology/vm/neovm/types"
)

// get hash from transaction
func TransactionGetHash(service *NeoVmService, engine *vm.ExecutionEngine) error {
	txn := vm.PopInteropInterface(engine).(*types.Transaction)
	txHash := txn.Hash()
	vm.PushData(engine, txHash.ToArray())
	return nil
}

// get type from transaction
func TransactionGetType(service *NeoVmService, engine *vm.ExecutionEngine) error {
	txn := vm.PopInteropInterface(engine).(*types.Transaction)
	vm.PushData(engine, int(txn.TxType))
	return nil
}

// get attributes from transaction
func TransactionGetAttributes(service *NeoVmService, engine *vm.ExecutionEngine) error {
	txn := vm.PopInteropInterface(engine).(*types.Transaction)
	attributes := txn.Attributes
	attributList := make([]vmtypes.StackItems, 0)
	for _, v := range attributes {
		attributList = append(attributList, vmtypes.NewInteropInterface(v))
	}
	vm.PushData(engine, attributList)
	return nil
}


