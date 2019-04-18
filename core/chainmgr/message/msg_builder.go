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
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology-crypto/signature"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	payload2 "github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/types"

	"github.com/ontio/ontology/core/xshard_types"
)

//
// NewCrossShardTxMsg: create cross-shard transaction, to remote ShardSysMsg contract
//  @payload: contains N sub-txns
//
//  One block can generated multiple cross-shard sub-txns, marshaled to [][]byte.
//  NewCrossShardTXMsg creates one cross-shard forwarding Tx, which contains all sub-txns.
//
func NewCrossShardTxMsg(account *account.Account, height uint32, toShardID common.ShardID, gasPrice, gasLimit uint64,
	msgs []xshard_types.CommonShardMsg) (*types.Transaction, error) {
	// build transaction
	shardCall := &payload2.ShardCall{
		Msgs: msgs,
	}
	mutable := &types.MutableTransaction{
		ShardID:  toShardID.ToUint64(),
		GasPrice: gasPrice,
		GasLimit: gasLimit,
		Payer:    account.Address,
		TxType:   types.ShardCall,
		Nonce:    height, // use height as nonce
		Payload:  shardCall,
		Sigs:     nil,
	}

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

func NewShardBlockInfo(shardID common.ShardID, block *types.Block) *ShardBlockInfo {
	blockInfo := &ShardBlockInfo{
		FromShardID: shardID,
		Height:      block.Header.Height,
		State:       ShardBlockNew,
		Block:       block,
	}

	// TODO: add event from block to blockInfo

	return blockInfo
}
