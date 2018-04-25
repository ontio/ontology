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
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	ledger "github.com/ontio/ontology/core/ledger/actor"
	"github.com/ontio/ontology/core/types"
)

const ledgerReqTimeout = 10 * time.Second

var defLedgerPid *actor.PID

func SetLedgerPid(ledgerPid *actor.PID) {
	defLedgerPid = ledgerPid
}

func AddHeader(header *types.Header) error {
	future := defLedgerPid.RequestFuture(&ledger.AddHeaderReq{Header: header}, ledgerReqTimeout)
	result, err := future.Result()
	if err != nil {
		return err
	}
	return result.(*ledger.AddHeadersRsp).Error
}

func AddHeaders(headers []*types.Header) error {
	future := defLedgerPid.RequestFuture(&ledger.AddHeadersReq{Headers: headers}, ledgerReqTimeout)
	result, err := future.Result()
	if err != nil {
		return err
	}
	return result.(*ledger.AddHeadersRsp).Error
}

func AddBlock(block *types.Block) error {
	future := defLedgerPid.RequestFuture(&ledger.AddBlockReq{Block: block}, ledgerReqTimeout)
	result, err := future.Result()
	if err != nil {
		return err
	}
	return result.(*ledger.AddBlockRsp).Error
}

func GetTxnFromLedger(hash common.Uint256) (*types.Transaction, error) {
	future := defLedgerPid.RequestFuture(&ledger.GetTransactionReq{TxHash: hash}, ledgerReqTimeout)
	result, err := future.Result()
	if err != nil {
		log.Errorf("GetTxnFromLedger Hash:%x defLedgerPid.RequestFuture error:%s", hash, err)
		return nil, err
	}
	return result.(*ledger.GetTransactionRsp).Tx, result.(*ledger.GetTransactionRsp).Error
}

func GetCurrentBlockHash() (common.Uint256, error) {
	future := defLedgerPid.RequestFuture(&ledger.GetCurrentBlockHashReq{}, ledgerReqTimeout)
	result, err := future.Result()
	if err != nil {
		log.Errorf("GetCurrentBlockHash defLedgerPid.RequestFuture error:%s", err)
		return common.Uint256{}, err
	}
	return result.(*ledger.GetCurrentBlockHashRsp).BlockHash, result.(*ledger.GetCurrentBlockHashRsp).Error
}

func GetCurrentHeaderHash() (common.Uint256, error) {
	future := defLedgerPid.RequestFuture(&ledger.GetCurrentHeaderHashReq{}, ledgerReqTimeout)
	result, err := future.Result()
	if err != nil {
		log.Errorf("GetCurrentHeaderHash defLedgerPid.RequestFuture error:%s", err)
		return common.Uint256{}, err
	}
	return result.(*ledger.GetCurrentHeaderHashRsp).BlockHash, result.(*ledger.GetCurrentHeaderHashRsp).Error
}

func GetBlockHashByHeight(height uint32) (common.Uint256, error) {
	future := defLedgerPid.RequestFuture(&ledger.GetBlockHashReq{Height: height}, ledgerReqTimeout)
	result, err := future.Result()
	if err != nil {
		log.Errorf("GetBlockHashByHeight Height:%d defLedgerPid.RequestFuture error:%s", height, err)
		return common.Uint256{}, err
	}
	return result.(*ledger.GetBlockHashRsp).BlockHash, result.(*ledger.GetBlockHashRsp).Error
}

func GetHeaderByHeight(height uint32) (*types.Header, error) {
	future := defLedgerPid.RequestFuture(&ledger.GetHeaderByHeightReq{Height: height}, ledgerReqTimeout)
	result, err := future.Result()
	if err != nil {
		log.Errorf("GetHeaderByHeight Height:%d defLedgerPid.RequestFuture error:%s", height, err)
		return nil, err
	}
	return result.(*ledger.GetHeaderByHeightRsp).Header, result.(*ledger.GetHeaderByHeightRsp).Error
}

func GetBlockByHeight(height uint32) (*types.Block, error) {
	future := defLedgerPid.RequestFuture(&ledger.GetBlockByHeightReq{Height: height}, ledgerReqTimeout)
	result, err := future.Result()
	if err != nil {
		log.Errorf("GetBlockByHeight Height:%d defLedgerPid.RequestFuture error:%s", height, err)
		return nil, err
	}
	return result.(*ledger.GetBlockByHeightRsp).Block, result.(*ledger.GetBlockByHeightRsp).Error
}

func GetHeaderByHash(hash common.Uint256) (*types.Header, error) {
	future := defLedgerPid.RequestFuture(&ledger.GetHeaderByHashReq{BlockHash: hash}, ledgerReqTimeout)
	result, err := future.Result()
	if err != nil {
		log.Errorf("GetHeaderByHash Hash:%x defLedgerPid.RequestFuture error:%s", hash, err)
		return nil, err
	}
	return result.(*ledger.GetHeaderByHashRsp).Header, result.(*ledger.GetHeaderByHashRsp).Error
}

func GetBlockByHash(hash common.Uint256) (*types.Block, error) {
	future := defLedgerPid.RequestFuture(&ledger.GetBlockByHashReq{BlockHash: hash}, ledgerReqTimeout)
	result, err := future.Result()
	if err != nil {
		log.Errorf("GetBlockByHash Hash:%x defLedgerPid.RequestFuture error:%s", hash, err)
		return nil, err
	}
	return result.(*ledger.GetBlockByHashRsp).Block, result.(*ledger.GetBlockByHashRsp).Error
}

func GetCurrentHeaderHeight() (uint32, error) {
	future := defLedgerPid.RequestFuture(&ledger.GetCurrentHeaderHeightReq{}, ledgerReqTimeout)
	result, err := future.Result()
	if err != nil {
		log.Errorf("GetCurrentHeaderHeight defLedgerPid.RequestFuture error:%s", err)
		return 0, err
	}
	return result.(*ledger.GetCurrentHeaderHeightRsp).Height, result.(*ledger.GetCurrentHeaderHeightRsp).Error
}

func GetCurrentBlockHeight() (uint32, error) {
	future := defLedgerPid.RequestFuture(&ledger.GetCurrentBlockHeightReq{}, ledgerReqTimeout)
	result, err := future.Result()
	if err != nil {
		log.Errorf("GetCurrentBlockHeight defLedgerPid.RequestFuture: %s", err)
		return 0, err
	}
	return result.(*ledger.GetCurrentBlockHeightRsp).Height, result.(*ledger.GetCurrentBlockHeightRsp).Error
}

func IsContainBlock(hash common.Uint256) (bool, error) {
	future := defLedgerPid.RequestFuture(&ledger.IsContainBlockReq{BlockHash: hash}, ledgerReqTimeout)
	result, err := future.Result()
	if err != nil {
		log.Errorf("IsContainBlock Hash:%x defLedgerPid.RequestFuture: %s", hash, err)
		return false, err
	}
	return result.(*ledger.IsContainBlockRsp).IsContain, result.(*ledger.IsContainBlockRsp).Error
}
