package neovm

import (
	vm "github.com/ontio/ontology/vm/neovm"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/core/types"
)

func HeaderGetHash(service *NeoVmService, engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 1 {
		return errors.NewErr("[HeaderGetHash] Too few input parameters ")
	}
	d := vm.PopInteropInterface(engine); if d == nil {
		return errors.NewErr("[HeaderGetHash] Pop blockdata nil!")
	}
	var data *types.Header
	if b, ok := d.(*types.Block); ok {
		data = b.Header
	} else if h, ok := d.(*types.Header); ok {
		data = h
	} else {
		return errors.NewErr("[HeaderGetHash] Wrong type!")
	}
	h := data.Hash()
	vm.PushData(engine, h.ToArray())
	return nil
}

func HeaderGetVersion(service *NeoVmService, engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 1 {
		return errors.NewErr("[HeaderGetVersion] Too few input parameters ")
	}
	d := vm.PopInteropInterface(engine); if d == nil {
		return errors.NewErr("[HeaderGetVersion] Pop blockdata nil!")
	}
	var data *types.Header
	if b, ok := d.(*types.Block); ok {
		data = b.Header
	} else if h, ok := d.(*types.Header); ok {
		data = h
	} else {
		return errors.NewErr("[HeaderGetVersion] Wrong type!")
	}
	vm.PushData(engine, data.Version)
	return nil
}

func HeaderGetPrevHash(service *NeoVmService, engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 1 {
		return errors.NewErr("[HeaderGetPrevHash] Too few input parameters ")
	}
	d := vm.PopInteropInterface(engine); if d == nil {
		return errors.NewErr("[HeaderGetPrevHash] Pop blockdata nil!")
	}
	var data *types.Header
	if b, ok := d.(*types.Block); ok {
		data = b.Header
	} else if h, ok := d.(*types.Header); ok {
		data = h
	} else {
		return errors.NewErr("[HeaderGetPrevHash] Wrong type!")
	}
	vm.PushData(engine, data.PrevBlockHash.ToArray())
	return nil
}

func HeaderGetMerkleRoot(service *NeoVmService, engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 1 {
		return errors.NewErr("[HeaderGetMerkleRoot] Too few input parameters ")
	}
	d := vm.PopInteropInterface(engine); if d == nil {
		return errors.NewErr("[HeaderGetMerkleRoot] Pop blockdata nil!")
	}
	var data *types.Header
	if b, ok := d.(*types.Block); ok {
		data = b.Header
	} else if h, ok := d.(*types.Header); ok {
		data = h
	} else {
		return errors.NewErr("[HeaderGetMerkleRoot] Wrong type!")
	}
	vm.PushData(engine, data.TransactionsRoot.ToArray())
	return nil
}

func HeaderGetIndex(service *NeoVmService, engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 1 {
		return errors.NewErr("[HeaderGetIndex] Too few input parameters ")
	}
	d := vm.PopInteropInterface(engine); if d == nil {
		return errors.NewErr("[HeaderGetIndex] Pop blockdata nil!")
	}
	var data *types.Header
	if b, ok := d.(*types.Block); ok {
		data = b.Header
	} else if h, ok := d.(*types.Header); ok {
		data = h
	} else {
		return errors.NewErr("[HeaderGetIndex] Wrong type!")
	}
	vm.PushData(engine, data.Height)
	return nil
}

func HeaderGetTimestamp(service *NeoVmService, engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 1 {
		return errors.NewErr("[HeaderGetTimestamp] Too few input parameters ")
	}
	d := vm.PopInteropInterface(engine); if d == nil {
		return errors.NewErr("[HeaderGetTimestamp] Pop blockdata nil!")
	}
	var data *types.Header
	if b, ok := d.(*types.Block); ok {
		data = b.Header
	} else if h, ok := d.(*types.Header); ok {
		data = h
	} else {
		return errors.NewErr("[HeaderGetTimestamp] Wrong type!")
	}
	vm.PushData(engine, data.Timestamp)
	return nil
}

func HeaderGetConsensusData(service *NeoVmService, engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 1 {
		return errors.NewErr("[HeaderGetConsensusData] Too few input parameters ")
	}
	d := vm.PopInteropInterface(engine); if d == nil {
		return errors.NewErr("[HeaderGetConsensusData] Pop blockdata nil!")
	}
	var data *types.Header
	if b, ok := d.(*types.Block); ok {
		data = b.Header
	} else if h, ok := d.(*types.Header); ok {
		data = h
	} else {
		return errors.NewErr("[HeaderGetConsensusData] Wrong type!")
	}
	vm.PushData(engine, data.ConsensusData)
	return nil
}

func HeaderGetNextConsensus(service *NeoVmService, engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 1 {
		return errors.NewErr("[HeaderGetNextConsensus] Too few input parameters ")
	}
	d := vm.PopInteropInterface(engine); if d == nil {
		return errors.NewErr("[HeaderGetNextConsensus] Pop blockdata nil!")
	}
	var data *types.Header
	if b, ok := d.(*types.Block); ok {
		data = b.Header
	} else if h, ok := d.(*types.Header); ok {
		data = h
	} else {
		return errors.NewErr("[HeaderGetNextConsensus] Wrong type!")
	}
	vm.PushData(engine, data.NextBookkeeper[:])
	return nil
}









