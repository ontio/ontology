package actor

import (
	"time"
	"github.com/Ontology/eventbus/actor"
	"github.com/Ontology/core/types"
	"github.com/Ontology/core/states"
	. "github.com/Ontology/common"
	"errors"
)

var defLedgerPid *actor.PID

func SetLedgerActor(actr *actor.PID) {
	defLedgerPid = actr
}

func GetBlockHashFromStore(height uint32) (Uint256, error) {
	future := defLedgerPid.RequestFuture(height, 10*time.Second)
	result, err := future.Result()
	if err != nil {
		return Uint256{}, err
	}
	if hash, ok := result.(Uint256); ok {
		return hash, nil
	}
	return Uint256{}, errors.New("fail")
	//return ledger.DefaultLedger.Store.GetBlockHash(height)
}

func CurrentBlockHash() (Uint256, error) {
	future := defLedgerPid.RequestFuture(nil, 10*time.Second)
	result, err := future.Result()
	if err != nil {
		return Uint256{}, err
	}
	if hash, ok := result.(Uint256); ok {
		return hash, nil
	}
	return Uint256{}, errors.New("fail")
	//return ledger.DefaultLedger.Blockchain.CurrentBlockHash(),nil
}

func GetBlockFromStore(hash Uint256) (*types.Block, error) {
	future := defLedgerPid.RequestFuture(hash, 10*time.Second)
	result, err := future.Result()
	if err != nil {
		return nil, err
	}
	if block, ok := result.(*types.Block); ok {
		return block, nil
	}
	return nil, errors.New("fail")
	//return ledger.DefaultLedger.Store.GetBlock(hash)
}
func BlockHeight() (uint32, error) {
	future := defLedgerPid.RequestFuture(nil, 10*time.Second)
	result, err := future.Result()
	if err != nil {
		return 0, err
	}
	if height, ok := result.(uint32); ok {
		return height, nil
	}
	return 0, errors.New("fail")
	//return ledger.DefaultLedger.Blockchain.BlockHeight,nil
}

func GetTransaction(hash Uint256) (*types.Transaction, error) {
	future := defLedgerPid.RequestFuture(hash, 10*time.Second)
	result, err := future.Result()
	if err != nil {
		return nil, err
	}
	if trans, ok := result.(*types.Transaction); ok {
		return trans, nil
	}
	return nil, errors.New("fail")
	//return ledger.DefaultLedger.Store.GetTransaction(hash)
}
func GetStorageItem(codeHash Uint160, key []byte) (*states.StorageItem, error) {
	future := defLedgerPid.RequestFuture(nil, 10*time.Second)
	result, err := future.Result()
	if err != nil {
		return nil, err
	}
	if item, ok := result.(*states.StorageItem); ok {
		return item, nil
	}
	return nil, errors.New("fail")
	//return ledger.DefaultLedger.Store.GetStorageItem(&states.StorageKey{CodeHash: codeHash, Key: key})
}

func GetAccount(programHash Uint160) (*states.AccountState, error) {
	future := defLedgerPid.RequestFuture(nil, 10*time.Second)
	result, err := future.Result()
	if err != nil {
		return nil, err
	}
	if stat, ok := result.(*states.AccountState); ok {
		return stat, nil
	}
	return nil, errors.New("fail")
	//return ledger.DefaultLedger.Store.GetAccount(programHash)
}

func GetContractFromStore(hash Uint160) (*states.ContractState, error) {
	future := defLedgerPid.RequestFuture(nil, 10*time.Second)
	result, err := future.Result()
	if err != nil {
		return nil, err
	}
	if stat, ok := result.(*states.ContractState); ok {
		return stat, nil
	}
	return nil, errors.New("fail")
	//return ledger.DefaultLedger.Store.GetContract(hash)
}
func AddBlock(block *types.Block) error {
	future := defLedgerPid.RequestFuture(nil, 10*time.Second)
	result, err := future.Result()
	if err != nil {
		return err
	}
	if result == nil {
		return nil
	}
	return errors.New("fail")
	//return ledger.DefaultLedger.Blockchain.AddBlock(block)
}
