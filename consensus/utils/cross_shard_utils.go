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

package utils

import (
	"fmt"
	"sort"

	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	vconfig "github.com/ontio/ontology/consensus/vbft/config"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/signature"
	common2 "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/core/xshard_types"
	shardstates "github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
	state "github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
)

func BuildCrossShardMsgs(signer *account.Account, lgr *ledger.Ledger, blkNum uint32, shardMsgs []xshard_types.CommonShardMsg, crossShardMsgHash *types.CrossShardMsgHash) (map[common.ShardID]*types.CrossShardMsg, common.Uint256, error) {
	builtMsgs := make(map[common.ShardID]*types.CrossShardMsg)
	hashRoot := common.UINT256_EMPTY
	if len(shardMsgs) == 0 {
		return builtMsgs, hashRoot, nil
	}

	// store all ShardMsgs by their target shard
	shardList := make([]common.ShardID, 0)
	shardMsgMap := make(map[common.ShardID][]xshard_types.CommonShardMsg)
	for _, msg := range shardMsgs {
		targetShardID := msg.GetTargetShardID()
		if _, present := shardMsgMap[targetShardID]; !present {
			shardList = append(shardList, targetShardID)
		}
		shardMsgMap[targetShardID] = append(shardMsgMap[targetShardID], msg)
	}

	// sort shards by shard-id
	sort.Slice(shardList, func(i, j int) bool { return shardList[i].ToUint64() < shardList[j].ToUint64() })

	// hash of serialized shard msgs
	shardMsgHashMap := make(map[common.ShardID]common.Uint256)
	for shardID, msgs := range shardMsgMap {
		msgHash := xshard_types.GetShardCommonMsgsHash(msgs)
		shardMsgHashMap[shardID] = msgHash
	}

	// compute hash Root in order of shard-id
	hashes := make([]common.Uint256, 0)
	for _, shardID := range shardList {
		hashes = append(hashes, shardMsgHashMap[shardID])
	}
	hashRoot = common.ComputeMerkleRoot(hashes)
	sigData := make(map[uint32][]byte)
	if crossShardMsgHash == nil {
		// sign on the hash root
		sig, err := signature.Sign(signer, hashRoot[:])
		if err != nil {
			return builtMsgs, hashRoot, fmt.Errorf("sign cross shard msg root failed,msg hash:%s,err:%s", hashRoot.ToHexString(), err)
		}
		//sigData := make(map[uint32][]byte)
		sigData[0] = sig
	} else {
		sigData = crossShardMsgHash.SigData
	}
	// broadcasting shard msgs to target shards
	for index, targetShardID := range shardList {
		// get msg-hash of other-shards
		otherShardMsgHashes := hashes[:index]
		if index+1 < len(shardList) {
			otherShardMsgHashes = append(otherShardMsgHashes, hashes[index+1:]...)
		}

		// get last shard-msg-root of the target shard
		prevMsgHash, err := lgr.GetShardMsgHash(targetShardID)
		if err != nil {
			if err != common2.ErrNotFound {
				return builtMsgs, hashRoot, fmt.Errorf("SendCrossShardMsgToAll getshardmsghash err:%s", err)
			}
		}

		// build cross-shard msg
		crossShardMsgInfo := &types.CrossShardMsgHash{
			ShardMsgHashs: otherShardMsgHashes,
			SigData:       sigData,
		}
		crossShardMsg := &types.CrossShardMsg{
			CrossShardMsgInfo: &types.CrossShardMsgInfo{
				SignMsgHeight:        blkNum,
				PreCrossShardMsgHash: prevMsgHash,
				Index:                uint32(index),
				ShardMsgInfo:         crossShardMsgInfo,
			},
			ShardMsg: shardMsgMap[targetShardID],
		}

		builtMsgs[targetShardID] = crossShardMsg
	}

	return builtMsgs, hashRoot, nil
}

func BuildCrossShardMsgHash(shardMsgs []xshard_types.CommonShardMsg) ([]common.Uint256, common.Uint256) {
	shardList := make([]common.ShardID, 0)
	shardMsgMap := make(map[common.ShardID][]xshard_types.CommonShardMsg)
	for _, msg := range shardMsgs {
		targetShardID := msg.GetTargetShardID()
		if _, present := shardMsgMap[targetShardID]; !present {
			shardMsgMap[targetShardID] = make([]xshard_types.CommonShardMsg, 0)
			shardList = append(shardList, targetShardID)
		}
		shardMsgMap[targetShardID] = append(shardMsgMap[targetShardID], msg)
	}

	// sort shards by shard-id
	sort.Slice(shardList, func(i, j int) bool { return shardList[i].ToUint64() < shardList[j].ToUint64() })

	// hash of serialized shard msgs
	shardMsgHashMap := make(map[common.ShardID]common.Uint256)
	for shardID, msgs := range shardMsgMap {
		msgHash := xshard_types.GetShardCommonMsgsHash(msgs)
		shardMsgHashMap[shardID] = msgHash
	}
	// compute hash Root in order of shard-id
	hashes := make([]common.Uint256, 0)
	for _, shardID := range shardList {
		hashes = append(hashes, shardMsgHashMap[shardID])
	}
	return hashes, common.ComputeMerkleRoot(hashes)
}

func GetShardConfigByShardID(lgr *ledger.Ledger, shardID common.ShardID, blkNum uint32) (*vconfig.ChainConfig, error) {
	data, err := lgr.GetShardConsensusConfig(shardID, blkNum)
	if err != nil {
		log.Errorf("getshardconsensusconfig shardID:%v,err:%s", shardID, err)
		return nil, err
	}
	source := common.NewZeroCopySource(data)
	shardEvent := &shardstates.ConfigShardEvent{}
	err = shardEvent.Deserialization(source)
	if err != nil {
		log.Errorf("getshardconfigbyshardID deserialization,shardID:%v err:%s", shardID, err)
		return nil, err
	}
	var peersInfo []*vconfig.PeerConfig
	for _, id := range shardEvent.Peers {
		if id.NodeType == state.CONSENSUS_NODE {
			peerInfo := &vconfig.PeerConfig{
				Index: id.Index,
				ID:    id.PeerPubKey,
			}
			peersInfo = append(peersInfo, peerInfo)
		}
	}
	cfg := &vconfig.ChainConfig{
		N:     shardEvent.Config.VbftCfg.N,
		C:     shardEvent.Config.VbftCfg.C,
		Peers: peersInfo,
	}
	return cfg, err
}
