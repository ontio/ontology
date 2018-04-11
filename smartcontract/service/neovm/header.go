package neovm

import (
	vm "github.com/ontio/ontology/vm/neovm"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/core/types"
)

// get hash from header
func HeaderGetHash(service *NeoVmService, engine *vm.ExecutionEngine) error {
	d := vm.PopInteropInterface(engine)
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

// get version from header
func HeaderGetVersion(service *NeoVmService, engine *vm.ExecutionEngine) error {
	d := vm.PopInteropInterface(engine)
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

// get prevhash from header
func HeaderGetPrevHash(service *NeoVmService, engine *vm.ExecutionEngine) error {
	d := vm.PopInteropInterface(engine)
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

// get merkle root from header
func HeaderGetMerkleRoot(service *NeoVmService, engine *vm.ExecutionEngine) error {
	d := vm.PopInteropInterface(engine)
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

// get height from header
func HeaderGetIndex(service *NeoVmService, engine *vm.ExecutionEngine) error {
	d := vm.PopInteropInterface(engine)
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

// get timestamp from header
func HeaderGetTimestamp(service *NeoVmService, engine *vm.ExecutionEngine) error {
	d := vm.PopInteropInterface(engine)
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

// get consensus data from header
func HeaderGetConsensusData(service *NeoVmService, engine *vm.ExecutionEngine) error {
	d := vm.PopInteropInterface(engine)
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

// get next consensus address from header
func HeaderGetNextConsensus(service *NeoVmService, engine *vm.ExecutionEngine) error {
	d := vm.PopInteropInterface(engine)
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









