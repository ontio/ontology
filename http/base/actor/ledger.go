package actor

import (
	"time"
	"github.com/Ontology/eventbus/actor"
	"github.com/Ontology/core/types"
	. "github.com/Ontology/core/ledger/actor"
	. "github.com/Ontology/common"
	"errors"
	"github.com/Ontology/smartcontract/event"
	"github.com/Ontology/core/payload"
)

var defLedgerPid *actor.PID

func SetLedgerPid(actr *actor.PID) {
	defLedgerPid = actr
}

//ledger.DefaultLedger.Store.GetBlockHash(height)
func GetBlockHashFromStore(height uint32) (Uint256, error) {
	future := defLedgerPid.RequestFuture(&GetBlockHashReq{height}, 10*time.Second)
	result, err := future.Result()
	if err != nil {
		return Uint256{}, err
	}
	if rsp, ok := result.(*GetBlockHashRsp); !ok {
		return Uint256{}, errors.New("fail")
	}else {
		return rsp.BlockHash, rsp.Error
	}
}

//ledger.DefaultLedger.Blockchain.CurrentBlockHash(),nil
func CurrentBlockHash() (Uint256, error) {
	future := defLedgerPid.RequestFuture(&GetCurrentBlockHashReq{}, 10*time.Second)
	result, err := future.Result()
	if err != nil {
		return Uint256{}, err
	}
	if rsp, ok := result.(*GetCurrentBlockHashRsp); !ok {
		return Uint256{}, errors.New("fail")
	}else {
		return rsp.BlockHash, rsp.Error
	}
}

//ledger.DefaultLedger.Store.GetBlock(hash)
func GetBlockFromStore(hash Uint256) (*types.Block, error) {
	future := defLedgerPid.RequestFuture(&GetBlockByHashReq{hash}, 10*time.Second)
	result, err := future.Result()
	if err != nil {
		return nil, err
	}
	if rsp, ok := result.(*GetBlockByHashRsp); !ok {
		return nil, errors.New("fail")
	}else {
		return rsp.Block, rsp.Error
	}
}

//ledger.DefaultLedger.Blockchain.BlockHeight,nil
func BlockHeight() (uint32, error) {
	future := defLedgerPid.RequestFuture(&GetCurrentBlockHeightReq{}, 10*time.Second)
	result, err := future.Result()
	if err != nil {
		return 0, err
	}
	if rsp, ok := result.(*GetCurrentBlockHeightRsp); !ok {
		return 0, errors.New("fail")
	}else {
		return rsp.Height, rsp.Error
	}
}

//ledger.DefaultLedger.Store.GetTransaction(hash)
func GetTransaction(hash Uint256) (*types.Transaction, error) {
	future := defLedgerPid.RequestFuture(&GetTransactionReq{hash}, 10*time.Second)
	result, err := future.Result()
	if err != nil {
		return nil, err
	}
	if rsp, ok := result.(*GetTransactionRsp); !ok {
		return nil, errors.New("fail")
	}else {
		return rsp.Tx, rsp.Error
	}
}

//ledger.DefaultLedger.Store.GetStorageItem(&states.StorageKey{CodeHash: codeHash, Key: key})
func GetStorageItem(codeHash Uint160, key []byte) ([]byte, error) {
	future := defLedgerPid.RequestFuture(&GetStorageItemReq{CodeHash:&codeHash,Key:key}, 10*time.Second)
	result, err := future.Result()
	if err != nil {
		return nil, err
	}
	if rsp, ok := result.(*GetStorageItemRsp); !ok {
		return nil, errors.New("fail")
	}else {
		return rsp.Value, rsp.Error
	}
}

//ledger.DefaultLedger.Store.GetContract(hash)
func GetContractStateFromStore(hash Uint160) (*payload.DeployCode, error) {
	future := defLedgerPid.RequestFuture(&GetContractStateReq{hash}, 10*time.Second)
	result, err := future.Result()
	if err != nil {
		return nil, err
	}
	if rsp, ok := result.(*GetContractStateRsp); !ok {
		return nil, errors.New("fail")
	}else {
		return rsp.ContractState, rsp.Error
	}
}

//ledger.DefaultLedger.Blockchain.AddBlock(block)
func AddBlock(block *types.Block) error {
	future := defLedgerPid.RequestFuture(&AddBlockReq{block}, 10*time.Second)
	result, err := future.Result()
	if err != nil {
		return err
	}
	if rsp, ok := result.(*AddBlockRsp); !ok {
		return errors.New("fail")
	}else {
		return rsp.Error
	}
}

func PreExecuteContract(tx *types.Transaction) ([]interface{}, error) {
	future := defLedgerPid.RequestFuture(&PreExecuteContractReq{tx}, 10*time.Second)
	result, err := future.Result()
	if err != nil {
		return nil,err
	}
	if rsp, ok := result.(*PreExecuteContractRsp); !ok {
		return rsp.Result,errors.New("fail")
	}else {
		return rsp.Result,rsp.Error
	}
}

func GetEventNotifyByTx(txHash Uint256) ([]*event.NotifyEventInfo, error) {
	future := defLedgerPid.RequestFuture(nil, 10*time.Second)
	_, err := future.Result()
	if err != nil {
		return nil,err
	}
	//TODO
	//if rsp, ok := result.(*GetEventNotifyByTxRsp); !ok {
	//	return rsp.Result,errors.New("fail")
	//}else {
	//	return rsp.Result,rsp.Error
	//}
	return nil,err
}

func GetEventNotifyByHeight(height uint32) ([]Uint256, error) {
	future := defLedgerPid.RequestFuture(nil, 10*time.Second)
	_, err := future.Result()
	if err != nil {
		return nil,err
	}
	//TODO
	//if rsp, ok := result.(*GetEventNotifyByBlockRsp); !ok {
	//	return rsp.Result,errors.New("fail")
	//}else {
	//	return rsp.Result,rsp.Error
	//}
	return nil,err
}