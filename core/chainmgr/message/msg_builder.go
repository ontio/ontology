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
	"bytes"
	"fmt"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology-crypto/signature"
	"github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/chainmgr/xshard_state"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/core/utils"
	"github.com/ontio/ontology/events/message"
	"github.com/ontio/ontology/smartcontract/service/native/shard_sysmsg"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt"
	utils2 "github.com/ontio/ontology/smartcontract/service/native/utils"
)

func NewShardHelloMsg(localShard, targetShard types.ShardID, sender *actor.PID) (*CrossShardMsg, error) {
	hello := &ShardHelloMsg{
		TargetShardID: targetShard,
		SourceShardID: localShard,
	}
	payload, err := EncodeShardMsg(hello)
	if err != nil {
		return nil, fmt.Errorf("marshal hello msg: %s", err)
	}

	return &CrossShardMsg{
		Version: SHARD_PROTOCOL_VERSION,
		Type:    HELLO_MSG,
		Sender:  sender,
		Data:    payload,
	}, nil
}

func NewShardConfigMsg(accPayload []byte, sibShards map[uint64]*SibShardInfo, configPayload []byte, sender *actor.PID) (*CrossShardMsg, error) {
	ack := &ShardConfigMsg{
		Account:   accPayload,
		SibShards: sibShards,
		Config:    configPayload,
	}
	payload, err := EncodeShardMsg(ack)
	if err != nil {
		return nil, fmt.Errorf("marshal hello ack msg: %s", err)
	}

	return &CrossShardMsg{
		Version: SHARD_PROTOCOL_VERSION,
		Type:    CONFIG_MSG,
		Sender:  sender,
		Data:    payload,
	}, nil
}

func NewShardBlockRspMsg(fromShardID types.ShardID, header *types.Header, tx *ShardBlockTx, sender *actor.PID) (*CrossShardMsg, error) {
	blkRsp := &ShardBlockRspMsg{
		FromShardID: fromShardID,
		Height:      header.Height,
		BlockHeader: &ShardBlockHeader{header},
		Txs:         []*ShardBlockTx{tx},
	}

	payload, err := EncodeShardMsg(blkRsp)
	if err != nil {
		return nil, fmt.Errorf("marshal shard block rsp msg: %s", err)
	}

	return &CrossShardMsg{
		Version: SHARD_PROTOCOL_VERSION,
		Type:    BLOCK_RSP_MSG,
		Sender:  sender,
		Data:    payload,
	}, nil
}

type _CrossShardTx struct {
	Txs [][]byte `json:"txs"`
}

func (this *_CrossShardTx) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint64(uint64(len(this.Txs)))
	for _, tx := range this.Txs {
		sink.WriteVarBytes(tx)
	}
}

//
// NewCrossShardTxMsg: create cross-shard transaction, to remote ShardSysMsg contract
//  @payload: contains N sub-txns
//
//  One block can generated multiple cross-shard sub-txns, marshaled to [][]byte.
//  NewCrossShardTXMsg creates one cross-shard forwarding Tx, which contains all sub-txns.
//
func NewCrossShardTxMsg(account *account.Account, height uint32, toShardID types.ShardID, gasPrice, gasLimit uint64, payload [][]byte) (*types.Transaction, error) {
	// marshal all sub-txns to one byte-array
	tx := &_CrossShardTx{payload}
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
	paramBytes := new(bytes.Buffer)
	if err := param.Serialize(paramBytes); err != nil {
		return nil, fmt.Errorf("marshal crossShardMsg: %s", err)
	}

	// build transaction
	mutable := utils.BuildNativeTransaction(utils2.ShardSysMsgContractAddress, shardsysmsg.PROCESS_CROSS_SHARD_MSG, paramBytes.Bytes())
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

func NewShardBlockInfo(shardID types.ShardID, header *types.Header) *ShardBlockInfo {
	blockInfo := &ShardBlockInfo{
		FromShardID: shardID,
		Height:      header.Height,
		State:       ShardBlockNew,
		Header: &ShardBlockHeader{
			Header: header,
		},
	}

	// TODO: add event from block to blockInfo

	return blockInfo
}

func NewShardBlockInfoFromRemote(ShardID types.ShardID, msg *ShardBlockRspMsg) (*ShardBlockInfo, error) {
	if msg == nil {
		return nil, fmt.Errorf("newShardBlockInfo, nil msg")
	}

	blockInfo := &ShardBlockInfo{
		FromShardID: msg.FromShardID,
		Height:      msg.BlockHeader.Header.Height,
		State:       ShardBlockReceived,
		Header: &ShardBlockHeader{
			Header: msg.BlockHeader.Header,
		},
		ShardTxs: make(map[types.ShardID]*ShardBlockTx),
	}

	if len(msg.Txs) > 0 {
		blockInfo.ShardTxs[ShardID] = msg.Txs[0]
	}
	// TODO: add event from msg to blockInfo

	return blockInfo, nil
}
