package neovm

import (
	vm "github.com/ontio/ontology/vm/neovm"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/core/types"
	vmtypes "github.com/ontio/ontology/vm/neovm/types"
)

func TransactionGetHash(service *NeoVmService, engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 1 {
		return errors.NewErr("[TransactionGetHash] Too few input parameters ")
	}
	d := vm.PopInteropInterface(engine); if d == nil {
		return errors.NewErr("[TransactionGetHash] Pop transaction nil!")
	}

	txn, ok := d.(*types.Transaction); if ok == false {
		return errors.NewErr("[TransactionGetHash] Wrong type!")
	}
	txHash := txn.Hash()
	vm.PushData(engine, txHash.ToArray())
	return nil
}

func TransactionGetType(service *NeoVmService, engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 1 {
		return errors.NewErr("[TransactionGetType] Too few input parameters ")
	}
	d := vm.PopInteropInterface(engine); if d == nil {
		return errors.NewErr("[TransactionGetType] Pop transaction nil!")
	}
	txn, ok := d.(*types.Transaction); if ok == false {
		return errors.NewErr("[TransactionGetHash] Wrong type!")
	}
	vm.PushData(engine, int(txn.TxType))
	return nil
}

func TransactionGetAttributes(service *NeoVmService, engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 1 {
		return errors.NewErr("[TransactionGetAttributes] Too few input parameters ")
	}
	d := vm.PopInteropInterface(engine); if d == nil {
		return errors.NewErr("[TransactionGetAttributes] Pop transaction nil!")
	}
	txn, ok := d.(*types.Transaction); if ok == false {
		return errors.NewErr("[TransactionGetAttributes] Wrong type!")
	}
	attributes := txn.Attributes
	attributList := make([]vmtypes.StackItems, 0)
	for _, v := range attributes {
		attributList = append(attributList, vmtypes.NewInteropInterface(v))
	}
	vm.PushData(engine, attributList)
	return nil
}


