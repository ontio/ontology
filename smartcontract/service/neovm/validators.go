package neovm

import (
	vm "github.com/ontio/ontology/vm/neovm"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/core/types"
)

func validatorAttribute(engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 1 {
		return errors.NewErr("[validatorAttribute] Too few input parameters ")
	}
	d := vm.PeekInteropInterface(engine); if d == nil {
		return errors.NewErr("[validatorAttribute] Pop txAttribute nil!")
	}
	_, ok := d.(*types.TxAttribute); if ok == false {
		return errors.NewErr("[validatorAttribute] Wrong type!")
	}
	return nil
}

func validatorBlock(engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 1 {
		return errors.NewErr("[Block] Too few input parameters ")
	}
	if _, err := peekBlock(engine); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[validatorBlock] Validate block fail!")
	}
	return nil
}

func validatorBlockTransaction(engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 2 {
		return errors.NewErr("[validatorBlockTransaction] Too few input parameters ")
	}
	block, err := peekBlock(engine); if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[validatorBlockTransaction] Validate block fail!")
	}
	index := vm.PeekInt(engine); if index < 0 {
		return errors.NewErr("[validatorBlockTransaction] Pop index invalid!")
	}
	if index >= len(block.Transactions) {
		return errors.NewErr("[validatorBlockTransaction] index invalid!")
	}
	return nil
}

func peekBlock(engine *vm.ExecutionEngine) (*types.Block, error) {
	d := vm.PeekInteropInterface(engine); if d == nil {
		return nil, errors.NewErr("[Block] Pop blockdata nil!")
	}
	block, ok := d.(*types.Block); if !ok {
		return nil, errors.NewErr("[Block] Wrong type!")
	}
	return block, nil
}

