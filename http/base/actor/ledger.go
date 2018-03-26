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
	lactor "github.com/Ontology/core/ledger/actor"
	"github.com/Ontology/common"
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

func GetBlockHashFromStore(height uint32) (common.Uint256, error) {
	future := defLedgerPid.RequestFuture(&lactor.GetBlockHashReq{height}, ReqTimeout*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ErrActorComm, err)
		return common.Uint256{}, err
	}
	if rsp, ok := result.(*lactor.GetBlockHashRsp); !ok {
		return common.Uint256{}, errors.New("fail")
	} else {
		return rsp.BlockHash, rsp.Error
	}
}

func CurrentBlockHash() (common.Uint256, error) {
	future := defLedgerPid.RequestFuture(&lactor.GetCurrentBlockHashReq{}, ReqTimeout*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ErrActorComm, err)
		return common.Uint256{}, err
	}
	if rsp, ok := result.(*lactor.GetCurrentBlockHashRsp); !ok {
		return common.Uint256{}, errors.New("fail")
	} else {
		return rsp.BlockHash, rsp.Error
	}
}

func GetBlockFromStore(hash common.Uint256) (*types.Block, error) {
	future := defLedgerPid.RequestFuture(&lactor.GetBlockByHashReq{hash}, ReqTimeout*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ErrActorComm, err)
		return nil, err
	}
	if rsp, ok := result.(*lactor.GetBlockByHashRsp); !ok {
		return nil, errors.New("fail")
	} else {
		return rsp.Block, rsp.Error
	}
}

func BlockHeight() (uint32, error) {
	future := defLedgerPid.RequestFuture(&lactor.GetCurrentBlockHeightReq{}, ReqTimeout*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ErrActorComm, err)
		return 0, err
	}
	if rsp, ok := result.(*lactor.GetCurrentBlockHeightRsp); !ok {
		return 0, errors.New("fail")
	} else {
		return rsp.Height, rsp.Error
	}
}

func GetTransaction(hash common.Uint256) (*types.Transaction, error) {
	future := defLedgerPid.RequestFuture(&lactor.GetTransactionReq{hash}, ReqTimeout*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ErrActorComm, err)
		return nil, err
	}
	if rsp, ok := result.(*lactor.GetTransactionRsp); !ok {
		return nil, errors.New("fail")
	} else {
		return rsp.Tx, rsp.Error
	}
}

func GetStorageItem(codeHash common.Address, key []byte) ([]byte, error) {
	future := defLedgerPid.RequestFuture(&lactor.GetStorageItemReq{CodeHash: &codeHash, Key: key}, ReqTimeout*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ErrActorComm, err)
		return nil, err
	}
	if rsp, ok := result.(*lactor.GetStorageItemRsp); !ok {
		return nil, errors.New("fail")
	} else {
		return rsp.Value, rsp.Error
	}
}

func GetContractStateFromStore(hash common.Address) (*payload.DeployCode, error) {
	future := defLedgerPid.RequestFuture(&lactor.GetContractStateReq{hash}, ReqTimeout*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ErrActorComm, err)
		return nil, err
	}
	if rsp, ok := result.(*lactor.GetContractStateRsp); !ok {
		log.Error(ErrActorComm,"GetContractStateRsp")
		return nil, errors.New("fail")
	} else {
		return rsp.ContractState, rsp.Error
	}
}

func GetBlockHeightByTxHashFromStore(hash common.Uint256) (uint32, error) {
	future := defLedgerPid.RequestFuture(&lactor.GetTransactionWithHeightReq{hash}, ReqTimeout*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ErrActorComm, err)
		return 0, err
	}
	if rsp, ok := result.(*lactor.GetTransactionWithHeightRsp); !ok {
		return 0, errors.New("fail")
	} else {
		return rsp.Height, rsp.Error
	}
}

func AddBlock(block *types.Block) error {
	future := defLedgerPid.RequestFuture(&lactor.AddBlockReq{block}, ReqTimeout*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ErrActorComm, err)
		return err
	}
	if rsp, ok := result.(*lactor.AddBlockRsp); !ok {
		return errors.New("fail")
	} else {
		return rsp.Error
	}
}

func PreExecuteContract(tx *types.Transaction) ([]interface{}, error) {
	future := defLedgerPid.RequestFuture(&lactor.PreExecuteContractReq{tx}, ReqTimeout*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ErrActorComm, err)
		return nil, err
	}
	if rsp, ok := result.(*lactor.PreExecuteContractRsp); !ok {
		return nil, errors.New("fail")
	} else {
		return rsp.Result, rsp.Error
	}
}

func GetEventNotifyByTxHash(txHash common.Uint256) ([]*event.NotifyEventInfo, error) {
	future := defLedgerPid.RequestFuture(&lactor.GetEventNotifyByTxReq{txHash}, ReqTimeout*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ErrActorComm, err)
		return nil, err
	}
	if rsp, ok := result.(*lactor.GetEventNotifyByTxRsp); !ok {
		return nil,errors.New("fail")
	}else {
		return rsp.Notifies,rsp.Error
	}
}

func GetEventNotifyByHeight(height uint32) ([]common.Uint256, error) {
	future := defLedgerPid.RequestFuture(&lactor.GetEventNotifyByBlockReq{height}, ReqTimeout*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ErrActorComm, err)
		return nil, err
	}
	if rsp, ok := result.(*lactor.GetEventNotifyByBlockRsp); !ok {
		return nil,errors.New("fail")
	}else {
		return rsp.TxHashes,rsp.Error
	}
}
