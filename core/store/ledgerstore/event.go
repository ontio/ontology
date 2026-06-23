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
package ledgerstore

import (
	"github.com/ethereum/go-ethereum/core"
	types2 "github.com/ethereum/go-ethereum/core/types"
	common2 "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/events"
	"github.com/ontio/ontology/events/message"
	"github.com/ontio/ontology/http/ethrpc/utils"
	"github.com/ontio/ontology/smartcontract/event"
)

func PushSmartCodeEvent(txHash common2.Uint256, errcode int64, action string, result interface{}) {
	if events.DefActorPublisher == nil {
		return
	}
	smartCodeEvt := &types.SmartCodeEvent{
		TxHash: txHash,
		Action: action,
		Result: result,
		Error:  errcode,
	}
	events.DefActorPublisher.Publish(message.TOPIC_SMART_CODE_EVENT, &message.SmartCodeEventMsg{Event: smartCodeEvt})
}

func PushEthSmartCodeEvent(rawNotify *event.ExecuteNotify, blk *types.Block) {
	if events.DefActorPublisher == nil {
		return
	}
	msg := extractSingleEthLog(rawNotify, blk)
	events.DefActorPublisher.Publish(message.TOPIC_ETH_SC_EVENT, &message.EthSmartCodeEventMsg{Event: msg})
}

func PushChainEvent(rawNotify []*event.ExecuteNotify, blk *types.Block, bloom types2.Bloom) {
	if events.DefActorPublisher == nil {
		return
	}
	events.DefActorPublisher.Publish(message.TOPIC_CHAIN_EVENT, &message.ChainEventMsg{
		ChainEvent: &core.ChainEvent{
			Header: utils.RawEthBlockFromOntology(blk, bloom).Header(),
		},
	})
}

func extractSingleEthLog(rawNotify *event.ExecuteNotify, blk *types.Block) []*types2.Log {
	var res []*types2.Log
	if isEIP155Tx(blk, rawNotify.TxHash) {
		res = genEthLog(rawNotify, blk)
	}
	return res
}

func extractEthLog(rawNotify []*event.ExecuteNotify, blk *types.Block) []*types2.Log {
	var res []*types2.Log
	for _, rn := range rawNotify {
		res = append(res, extractSingleEthLog(rn, blk)...)
	}
	return res
}

func genEthLog(rawNotify *event.ExecuteNotify, blk *types.Block) []*types2.Log {
	var res []*types2.Log
	txHash := rawNotify.TxHash
	ethHash := utils.OntToEthHash(txHash)
	for idx, n := range rawNotify.Notify {
		storageLog, err := event.NotifyEventInfoToEvmLog(n)
		if err != nil {
			return nil
		}
		res = append(res,
			&types2.Log{
				Address:     storageLog.Address,
				Topics:      storageLog.Topics,
				Data:        storageLog.Data,
				BlockNumber: uint64(blk.Header.Height),
				TxHash:      utils.OntToEthHash(txHash),
				TxIndex:     uint(rawNotify.TxIndex),
				BlockHash:   ethHash,
				Index:       uint(idx),
				Removed:     false,
			})
	}
	return res
}

func isEIP155Tx(block *types.Block, txHash common2.Uint256) bool {
	for _, tx := range block.Transactions {
		if tx.Hash() == txHash {
			return tx.IsEipTx()
		}
	}
	return false
}
