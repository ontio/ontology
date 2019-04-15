/*
 * Copyright (C) 2019 The ontology Authors
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

package message

import (
	"fmt"
	"math"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology-crypto/signature"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/chainmgr/xshard_state"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/events/message"
	bcommon "github.com/ontio/ontology/http/base/common"
	"github.com/ontio/ontology/smartcontract/service/native/shard_sysmsg"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

//
// NewCrossShardTxMsg: create cross-shard transaction, to remote ShardSysMsg contract
//  @payload: contains N sub-txns
//
//  One block can generated multiple cross-shard sub-txns, marshaled to [][]byte.
//  NewCrossShardTXMsg creates one cross-shard forwarding Tx, which contains all sub-txns.
//
func NewCrossShardTxMsg(account *account.Account, height uint32, toShardID types.ShardID, gasPrice, gasLimit uint64, payload [][]byte) (*types.Transaction, error) {
	// marshal all sub-txns to one byte-array
	tx := &xshard_state.CrossShardTx{payload}
	sink := common.NewZeroCopySink(0)
	tx.Serialization(sink)
	// cross-shard forwarding Tx payload
	evt := &message.ShardEventState{
		Version:    shardmgmt.VERSION_CONTRACT_SHARD_MGMT,
		EventType:  xshard_state.EVENT_SHARD_MSG_COMMON,
		ToShard:    toShardID,
		FromHeight: height,
		Payload:    sink.Bytes(),
	}

	// marshal to CrossShardMsgParam
	param := &shardsysmsg.CrossShardMsgParam{
		Events: []*message.ShardEventState{evt},
	}
	// build transaction
	mutable, err := bcommon.NewNativeInvokeTransaction(0, math.MaxUint32, utils.ShardSysMsgContractAddress,
		byte(0), shardsysmsg.PROCESS_CROSS_SHARD_MSG, []interface{}{param})
	if err != nil {
		return nil, fmt.Errorf("NewCrossShardTxMsg: build tx failed, err: %s", err)
	}
	mutable.ShardID = toShardID.ToUint64()
	mutable.GasPrice = gasPrice
	mutable.GasLimit = gasLimit
	mutable.Payer = account.Address

	// add signatures
	txHash := mutable.Hash()
	sig, err := signature.Sign(account.SigScheme, account.PrivateKey, txHash.ToArray(), nil)
	if err != nil {
		return nil, fmt.Errorf("sign tx: %s", err)
	}
	sigData, err := signature.Serialize(sig)
	if err != nil {
		return nil, fmt.Errorf("serialize sig: %s", err)
	}
	mutable.Sigs = []types.Sig{
		{
			PubKeys: []keypair.PublicKey{account.PubKey()},
			M:       1,
			SigData: [][]byte{sigData},
		},
	}
	return mutable.IntoImmutable()
}

func NewShardBlockInfo(shardID types.ShardID, block *types.Block) *ShardBlockInfo {
	blockInfo := &ShardBlockInfo{
		FromShardID: shardID,
		Height:      block.Header.Height,
		State:       ShardBlockNew,
		Block:       block,
	}

	// TODO: add event from block to blockInfo

	return blockInfo
}
