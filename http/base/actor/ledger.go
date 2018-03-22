/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package actor

import (
	"time"
	"github.com/ontio/ontology-eventbus/actor"
	"github.com/Ontology/core/types"
	. "github.com/Ontology/core/ledger/actor"
	. "github.com/Ontology/common"
	"errors"
	"github.com/Ontology/smartcontract/event"
	"github.com/Ontology/core/payload"
	"github.com/Ontology/common/log"
)

const (
	ReqTimeout   = 5
	ErrActorComm = "[http] Actor comm error: %v"
)

var defLedgerPid *actor.PID

func SetLedgerPid(actr *actor.PID) {
	defLedgerPid = actr
}

func GetBlockHashFromStore(height uint32) (Uint256, error) {
	future := defLedgerPid.RequestFuture(&GetBlockHashReq{height}, ReqTimeout*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ErrActorComm, err)
		return Uint256{}, err
	}
	if rsp, ok := result.(*GetBlockHashRsp); !ok {
		return Uint256{}, errors.New("fail")
	} else {
		return rsp.BlockHash, rsp.Error
	}
}

func CurrentBlockHash() (Uint256, error) {
	future := defLedgerPid.RequestFuture(&GetCurrentBlockHashReq{}, ReqTimeout*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ErrActorComm, err)
		return Uint256{}, err
	}
	if rsp, ok := result.(*GetCurrentBlockHashRsp); !ok {
		return Uint256{}, errors.New("fail")
	} else {
		return rsp.BlockHash, rsp.Error
	}
}

func GetBlockFromStore(hash Uint256) (*types.Block, error) {
	future := defLedgerPid.RequestFuture(&GetBlockByHashReq{hash}, ReqTimeout*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ErrActorComm, err)
		return nil, err
	}
	if rsp, ok := result.(*GetBlockByHashRsp); !ok {
		return nil, errors.New("fail")
	} else {
		return rsp.Block, rsp.Error
	}
}

func BlockHeight() (uint32, error) {
	future := defLedgerPid.RequestFuture(&GetCurrentBlockHeightReq{}, ReqTimeout*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ErrActorComm, err)
		return 0, err
	}
	if rsp, ok := result.(*GetCurrentBlockHeightRsp); !ok {
		return 0, errors.New("fail")
	} else {
		return rsp.Height, rsp.Error
	}
}

func GetTransaction(hash Uint256) (*types.Transaction, error) {
	future := defLedgerPid.RequestFuture(&GetTransactionReq{hash}, ReqTimeout*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ErrActorComm, err)
		return nil, err
	}
	if rsp, ok := result.(*GetTransactionRsp); !ok {
		return nil, errors.New("fail")
	} else {
		return rsp.Tx, rsp.Error
	}
}

func GetStorageItem(codeHash Address, key []byte) ([]byte, error) {
	future := defLedgerPid.RequestFuture(&GetStorageItemReq{CodeHash: &codeHash, Key: key}, ReqTimeout*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ErrActorComm, err)
		return nil, err
	}
	if rsp, ok := result.(*GetStorageItemRsp); !ok {
		return nil, errors.New("fail")
	} else {
		return rsp.Value, rsp.Error
	}
}

func GetContractStateFromStore(hash Address) (*payload.DeployCode, error) {
	future := defLedgerPid.RequestFuture(&GetContractStateReq{hash}, ReqTimeout*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ErrActorComm, err)
		return nil, err
	}
	if rsp, ok := result.(*GetContractStateRsp); !ok {
		return nil, errors.New("fail")
	} else {
		return rsp.ContractState, rsp.Error
	}
}

func GetBlockHeightByTxHashFromStore(hash Uint256) (uint32, error) {
	future := defLedgerPid.RequestFuture(&GetTransactionWithHeightReq{hash}, ReqTimeout*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ErrActorComm, err)
		return 0, err
	}
	if rsp, ok := result.(*GetTransactionWithHeightRsp); !ok {
		return 0, errors.New("fail")
	} else {
		return rsp.Height, rsp.Error
	}
}

func AddBlock(block *types.Block) error {
	future := defLedgerPid.RequestFuture(&AddBlockReq{block}, ReqTimeout*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ErrActorComm, err)
		return err
	}
	if rsp, ok := result.(*AddBlockRsp); !ok {
		return errors.New("fail")
	} else {
		return rsp.Error
	}
}

func PreExecuteContract(tx *types.Transaction) ([]interface{}, error) {
	future := defLedgerPid.RequestFuture(&PreExecuteContractReq{tx}, ReqTimeout*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ErrActorComm, err)
		return nil, err
	}
	if rsp, ok := result.(*PreExecuteContractRsp); !ok {
		return nil, errors.New("fail")
	} else {
		return rsp.Result, rsp.Error
	}
}

func GetEventNotifyByTx(txHash Uint256) ([]*event.NotifyEventInfo, error) {
	future := defLedgerPid.RequestFuture(nil, ReqTimeout*time.Second)
	_, err := future.Result()
	if err != nil {
		log.Errorf(ErrActorComm, err)
		return nil, err
	}
	//TODO
	//if rsp, ok := result.(*GetEventNotifyByTxRsp); !ok {
	//	return rsp.Result,errors.New("fail")
	//}else {
	//	return rsp.Result,rsp.Error
	//}
	return nil, err
}

func GetEventNotifyByHeight(height uint32) ([]Uint256, error) {
	future := defLedgerPid.RequestFuture(nil, ReqTimeout*time.Second)
	_, err := future.Result()
	if err != nil {
		log.Errorf(ErrActorComm, err)
		return nil, err
	}
	//TODO
	//if rsp, ok := result.(*GetEventNotifyByBlockRsp); !ok {
	//	return rsp.Result,errors.New("fail")
	//}else {
	//	return rsp.Result,rsp.Error
	//}
	return nil, err
}
