package neovm

import (
	vm "github.com/ontio/ontology/vm/neovm"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/common"
	vmtypes "github.com/ontio/ontology/vm/neovm/types"
)

// get height from blockchain
func BlockChainGetHeight(service *NeoVmService, engine *vm.ExecutionEngine) error {
	vm.PushData(engine, service.Store.GetCurrentBlockHeight())
	return nil
}

// get header from blockchain
func BlockChainGetHeader(service *NeoVmService, engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 1 {
		return errors.NewErr("[BlockChainGetHeader] Too few input parameters ")
	}
	data := vm.PopByteArray(engine)
	var (
		header *types.Header
		err error
	)
	l := len(data)
	if l <= 5 {
		b := vmtypes.ConvertBytesToBigInteger(data)
		height := uint32(b.Int64())
		hash := service.Store.GetBlockHash(height)
		header, err = service.Store.GetHeaderByHash(hash)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[BlockChainGetHeader] GetHeader error!.")
		}
	} else if l == 32 {
		hash, _ := common.Uint256ParseFromBytes(data)
		header, err = service.Store.GetHeaderByHash(hash)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[BlockChainGetHeader] GetHeader error!.")
		}
	} else {
		return errors.NewErr("[BlockChainGetHeader] data invalid.")
	}
	vm.PushData(engine, header)
	return nil
}

// get block from blockchain
func BlockChainGetBlock(service *NeoVmService, engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 1 {
		return errors.NewErr("[BlockChainGetBlock] Too few input parameters ")
	}
	data := vm.PopByteArray(engine)
	var block *types.Block
	l := len(data)
	if l <= 5 {
		b := vmtypes.ConvertBytesToBigInteger(data)
		height := uint32(b.Int64())
		var err error
		block, err = service.Store.GetBlockByHeight(height)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[BlockChainGetBlock] GetBlock error!.")
		}
	} else if l == 32 {
		hash, err := common.Uint256ParseFromBytes(data)
		if err != nil {
			return err
		}
		block, err = service.Store.GetBlockByHash(hash)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[BlockChainGetBlock] GetBlock error!.")
		}
	} else {
		return errors.NewErr("[BlockChainGetBlock] data invalid.")
	}
	vm.PushData(engine, block)
	return nil
}

// get transaction from blockchain
func BlockChainGetTransaction(service *NeoVmService, engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 1 {
		return errors.NewErr("[BlockChainGetTransaction] Too few input parameters ")
	}
	d := vm.PopByteArray(engine)
	hash, err := common.Uint256ParseFromBytes(d); if err != nil {
		return err
	}
	t, _, err := service.Store.GetTransaction(hash); if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[BlockChainGetTransaction] GetTransaction error!")
	}
	vm.PushData(engine, t)
	return nil
}

// get contract from blockchain
func BlockChainGetContract(service *NeoVmService, engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 1 {
		return errors.NewErr("[GetContract] Too few input parameters ")
	}
	address, err := common.AddressParseFromBytes(vm.PopByteArray(engine)); if err != nil {
		return err
	}
	item, err := service.Store.GetContractState(address); if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[GetContract] GetAsset error!")
	}
	vm.PushData(engine, item)
	return nil
}


