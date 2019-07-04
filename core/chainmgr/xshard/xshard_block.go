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

package xshard

import (
	"fmt"
	"sync"

	vconfig "github.com/ontio/ontology/consensus/vbft/config"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	csm "github.com/ontio/ontology/consensus/utils"
	crossshard "github.com/ontio/ontology/core/chainmgr/message"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/payload"
	sign "github.com/ontio/ontology/core/signature"
	com "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/core/xshard_types"
)

// cross  shard pool

type CrossShardPool struct {
	lock        sync.RWMutex
	ShardID     common.ShardID
	Shards      map[common.ShardID]map[common.Uint256]*types.CrossShardMsg // key:indexed by FromShardID key:preMsgHash
	MaxBlockCap uint32
	ShardInfo   map[common.ShardID]bool
}

// BlockHeader and Cross-Shard Txs of other shards
var crossShardPool *CrossShardPool

func InitCrossShardPool(shardID common.ShardID, historyCap uint32) {
	crossShardPool = &CrossShardPool{
		ShardID:     shardID,
		Shards:      make(map[common.ShardID]map[common.Uint256]*types.CrossShardMsg),
		ShardInfo:   make(map[common.ShardID]bool),
		MaxBlockCap: historyCap,
	}
}

func InitShardInfo(lgr *ledger.Ledger) error {
	pool := crossShardPool
	pool.lock.Lock()
	defer pool.lock.Unlock()
	shardIds, err := lgr.GetAllShardIDs()
	if err != nil {
		if err != com.ErrNotFound {
			return fmt.Errorf("GetAllShardIDs failed err:%s", err)
		}
	}
	for _, shardId := range shardIds {
		pool.ShardInfo[shardId] = true
		msgHash, err := GetCrossShardHashByShardID(lgr, shardId)
		if err != nil {
			if err != com.ErrNotFound {
				return fmt.Errorf("InitShardInfo GetCrossShardHashByShardID shardID:%v,err:%s", shardId, err)
			} else {
				break
			}
		}
		for {
			msg, err := lgr.GetCrossShardMsgByHash(msgHash)
			if err != nil {
				if err != com.ErrNotFound {
					return fmt.Errorf("InitShardInfo GetCrossShardMsgByHash hash:%s,err:%s", msgHash.ToHexString(), err)
				} else {
					break
				}
			}
			if _, present := pool.Shards[shardId]; !present {
				pool.Shards[shardId] = make(map[common.Uint256]*types.CrossShardMsg)
			}
			shardMsgMap := pool.Shards[shardId]
			shardMsgMap[msg.CrossShardMsgInfo.PreCrossShardMsgHash] = msg
			if len(shardMsgMap) >= int(pool.MaxBlockCap) {
				break
			}
			hashes := msg.CrossShardMsgInfo.ShardMsgInfo.ShardMsgHashs
			hash := xshard_types.GetShardCommonMsgsHash(msg.ShardMsg)
			hashes = append(hashes, hash)
			msgHash = common.ComputeMerkleRoot(hashes)
		}
	}
	return nil
}

func addShardInfo(lgr *ledger.Ledger, shardID common.ShardID) {
	pool := crossShardPool
	if _, present := pool.ShardInfo[shardID]; present {
		return
	}
	pool.ShardInfo[shardID] = true

	shardIds := make([]common.ShardID, 0)
	for shardId := range pool.ShardInfo {
		shardIds = append(shardIds, shardId)
	}
	err := lgr.SaveAllShardIDs(shardIds)
	if err != nil {
		log.Errorf("SaveAllShardIDs shardId:%v,err:%s", shardID, err)
		return
	}
}

func GetShardInfo() map[common.ShardID]bool {
	pool := crossShardPool
	pool.lock.RLock()
	defer pool.lock.RUnlock()
	return pool.ShardInfo
}

func GetCrossShardHashByShardID(lgr *ledger.Ledger, shardID common.ShardID) (common.Uint256, error) {
	return lgr.GetCrossShardHash(shardID)
}
func SaveCrossShardHash(lgr *ledger.Ledger, shardID common.ShardID, msgHash common.Uint256) error {
	return lgr.SaveCrossShardHash(shardID, msgHash)
}

func GetCrossShardMsg(lgr *ledger.Ledger, sourceShardID common.ShardID, msgHash common.Uint256) (*types.CrossShardMsg, error) {
	pool := crossShardPool
	pool.lock.RLock()
	defer pool.lock.RUnlock()
	if csMsg, present := pool.Shards[sourceShardID]; present {
		if msg, p := csMsg[msgHash]; p {
			return msg, nil
		}
	}
	return lgr.GetCrossShardMsgByHash(msgHash)
}

func AddCrossShardInfo(lgr *ledger.Ledger, crossShardMsg *types.CrossShardMsg) error {
	pool := crossShardPool
	pool.lock.Lock()
	defer pool.lock.Unlock()
	if len(crossShardMsg.ShardMsg) == 0 {
		log.Errorf("add cross shard info shardmsg is nil")
		return nil
	}
	sourceShardID := crossShardMsg.ShardMsg[0].GetSourceShardID()
	if _, present := pool.Shards[sourceShardID]; !present {
		pool.Shards[sourceShardID] = make(map[common.Uint256]*types.CrossShardMsg)
	}
	m := pool.Shards[sourceShardID]
	if m == nil {
		return fmt.Errorf("add shard cross shard msg:%d, nil map", sourceShardID.ToUint64())
	}
	if _, present := m[crossShardMsg.CrossShardMsgInfo.PreCrossShardMsgHash]; present {
		log.Debugf("SaveCrossShardMsgByShardID premsgHash:%s had save", crossShardMsg.CrossShardMsgInfo.PreCrossShardMsgHash.ToHexString())
		return nil
	}
	if !VerifyCrossShardMsg(pool.ShardID, sourceShardID, lgr, crossShardMsg.CrossShardMsgInfo, crossShardMsg.ShardMsg) {
		return fmt.Errorf("verify cross shard msg err")
	}
	hash, err := GetCrossShardHashByShardID(lgr, sourceShardID)
	if err != nil {
		if err != com.ErrNotFound {
			return fmt.Errorf("GetCrossShardHashByShardID shardID:%v,err:%s", sourceShardID, err)
		} else {
			err = SaveCrossShardHash(lgr, sourceShardID, crossShardMsg.CrossShardMsgInfo.PreCrossShardMsgHash)
			if err != nil {
				return fmt.Errorf("SaveCrossShardHash from shardID:%v,err:%s", sourceShardID, err)
			}
		}
	} else {
		if shardmsg, present := m[hash]; present {
			if shardmsg.CrossShardMsgInfo.SignMsgHeight >= crossShardMsg.CrossShardMsgInfo.SignMsgHeight {
				return fmt.Errorf("AddCrossShardInfo cross shard msg sign msg height:%d,last msg sign msg height:%d,sourceShardID:%v,preHash:%s", shardmsg.CrossShardMsgInfo.SignMsgHeight,
					crossShardMsg.CrossShardMsgInfo.SignMsgHeight, sourceShardID, crossShardMsg.CrossShardMsgInfo.PreCrossShardMsgHash.ToHexString())
			}
		} else {
			msg, err := lgr.GetCrossShardMsgByHash(hash)
			if err != nil && err != com.ErrNotFound {
				return fmt.Errorf("AddCrossShardInfo before hash:%s not found in db sourceShardID:%v,preHash:%s", hash.ToHexString(), sourceShardID, crossShardMsg.CrossShardMsgInfo.PreCrossShardMsgHash.ToHexString())
			} else if err == nil {
				if msg.CrossShardMsgInfo.SignMsgHeight >= crossShardMsg.CrossShardMsgInfo.SignMsgHeight {
					return fmt.Errorf("AddCrossShardInfo cross shard msg sign msg height:%d,last msg sign msg height:%d,sourceShardID:%v,preHash:%s", shardmsg.CrossShardMsgInfo.SignMsgHeight,
						crossShardMsg.CrossShardMsgInfo.SignMsgHeight, sourceShardID, crossShardMsg.CrossShardMsgInfo.PreCrossShardMsgHash.ToHexString())
				}
			}
		}
	}
	m[crossShardMsg.CrossShardMsgInfo.PreCrossShardMsgHash] = crossShardMsg
	err = lgr.SaveCrossShardMsgByHash(crossShardMsg.CrossShardMsgInfo.PreCrossShardMsgHash, crossShardMsg)
	if err != nil {
		return fmt.Errorf("SaveCrossShardMsgByShardID shardID:%v,msgHash:%s,err:%s", sourceShardID, crossShardMsg.CrossShardMsgInfo.PreCrossShardMsgHash.ToHexString(), err)
	}
	addShardInfo(lgr, sourceShardID)
	log.Infof("chainmgr AddBlock from shard %d,msgHash:%v, block height %d", sourceShardID, crossShardMsg.CrossShardMsgInfo.PreCrossShardMsgHash.ToHexString(), crossShardMsg.CrossShardMsgInfo.SignMsgHeight)
	return nil
}

//
// GetShardTxsByParentHeight
// Get cross-shard Tx/Events from parent shard.
// Cross-shard Tx/events of parent shard are delivered to child shards with parent-block propagation.
// NOTE: all cross-shard tx/events should be indexed with (parentHeight, shardHeight)
//

func GetCrossShardTxs(lgr *ledger.Ledger, account *account.Account, toShardID common.ShardID, beginParentblkNum, endParentblkNum uint32) (map[uint64][]*types.CrossShardTxInfos, error) {
	pool := crossShardPool
	pool.lock.RLock()
	defer pool.lock.RUnlock()
	crossShardMapInfos := make(map[uint64][]*types.CrossShardTxInfos)
	if !toShardID.IsRootShard() && lgr.ParentLedger != nil {
		crossShardInfo := make([]*types.CrossShardTxInfos, 0)
		for blkNum := beginParentblkNum; blkNum <= endParentblkNum; blkNum++ {
			shardMsg, err := lgr.ParentLedger.GetShardMsgsInBlock(blkNum, toShardID)
			if err != nil && err != com.ErrNotFound {
				return nil, fmt.Errorf("GetShardMsgsInBlock parentblkNum:%d,shardID:%v,err:%s", blkNum, toShardID, err)
			} else if err == com.ErrNotFound {
				log.Infof("GetShardMsgsInBlock parentblkNum:%d,shardID:%v,err:%s", blkNum, toShardID, err)
				continue
			}
			tx, err := crossshard.NewCrossShardTxMsg(account, blkNum, toShardID, config.DefConfig.Common.GasPrice, config.DefConfig.Common.GasLimit, shardMsg)
			if err != nil {
				return nil, fmt.Errorf("handleCrossShardMsg NewCrossShardTxMsg height:%d,err:%s", blkNum, err)
			}
			shardTxInfo := &types.CrossShardTxInfos{
				Tx: tx,
			}
			crossShardInfo = append(crossShardInfo, shardTxInfo)
		}
		crossShardMapInfos[toShardID.ParentID().ToUint64()] = crossShardInfo
	}
	for shardID, shardMsgs := range pool.Shards {
		crossShardInfo := make([]*types.CrossShardTxInfos, 0)
		id := shardID
		msgHash, err := GetCrossShardHashByShardID(lgr, id)
		if err != nil && err != com.ErrNotFound {
			log.Errorf("GetCrossShardHashByShardID shardID:%v,err:%s", shardID, err)
			return crossShardMapInfos, nil
		} else if err == com.ErrNotFound {
			continue
		}
		crossShardMsgs := make([]*types.CrossShardMsg, 0)
		for {
			if shardMsg, persent := shardMsgs[msgHash]; !persent {
				msg, err := lgr.GetCrossShardMsgByHash(msgHash)
				if err != nil {
					if err != com.ErrNotFound {
						return nil, fmt.Errorf("GetCrossShardMsgByHash msgHash:%s,err:%v", msgHash, err)
					} else {
						break
					}
				} else {
					crossShardMsgs = append(crossShardMsgs, msg)
					msgHash = CrossShardMsgHash(msg.CrossShardMsgInfo, msg.ShardMsg)
				}
			} else {
				crossShardMsgs = append(crossShardMsgs, shardMsg)
				msgHash = CrossShardMsgHash(shardMsg.CrossShardMsgInfo, shardMsg.ShardMsg)
			}
		}
		for _, msg := range crossShardMsgs {
			tx, err := crossshard.NewCrossShardTxMsg(account, msg.CrossShardMsgInfo.SignMsgHeight, toShardID, config.DefConfig.Common.GasPrice, config.DefConfig.Common.GasLimit, msg.ShardMsg)
			if err != nil {
				log.Errorf("handleCrossShardMsg NewCrossShardTxMsg height:%d,err:%s", msg.CrossShardMsgInfo.SignMsgHeight, err)
				break
			}
			shardTxInfo := &types.CrossShardTxInfos{
				ShardMsg: msg.CrossShardMsgInfo,
				Tx:       tx,
			}
			crossShardInfo = append(crossShardInfo, shardTxInfo)
		}
		crossShardMapInfos[shardID.ToUint64()] = crossShardInfo
	}
	return crossShardMapInfos, nil
}

func DelCrossShardTxs(lgr *ledger.Ledger, crossShardTxs map[uint64][]*types.CrossShardTxInfos) error {
	pool := crossShardPool
	pool.lock.Lock()
	defer pool.lock.Unlock()
	for id, shardTxs := range crossShardTxs {
		shardID := common.NewShardIDUnchecked(id)
		for _, shardTx := range shardTxs {
			if msg, present := pool.Shards[shardID]; !present {
				log.Infof("delcrossshardtxs shardID:%v,not exist", shardID)
				return nil
			} else {
				log.Infof("delcrossshardtxs shardID:%v,msgHash:%s", shardID, shardTx.ShardMsg.PreCrossShardMsgHash.ToHexString())
				delete(msg, shardTx.ShardMsg.PreCrossShardMsgHash)
				shardCall := shardTx.Tx.Payload.(*payload.ShardCall)
				msgRoot := CrossShardMsgHash(shardTx.ShardMsg, shardCall.Msgs)
				err := SaveCrossShardHash(lgr, shardID, msgRoot)
				if err != nil {
					log.Errorf("SaveCrossShardHash shardID:%v,preMsgHash:%s,failed err:%s", shardID, msgRoot.ToHexString(), err)
				}
			}
		}
	}
	return nil
}

func CrossShardMsgHash(crossShardMsgInfo *types.CrossShardMsgInfo, msgs []xshard_types.CommonShardMsg) common.Uint256 {
	hashes := make([]common.Uint256, 0)
	for index, hash := range crossShardMsgInfo.ShardMsgInfo.ShardMsgHashs {
		if uint32(index) == crossShardMsgInfo.Index {
			hashes = append(hashes, xshard_types.GetShardCommonMsgsHash(msgs))
		}
		hashes = append(hashes, hash)
	}
	if crossShardMsgInfo.Index > uint32(len(crossShardMsgInfo.ShardMsgInfo.ShardMsgHashs)) || len(crossShardMsgInfo.ShardMsgInfo.ShardMsgHashs) == 0 {
		hashes = append(hashes, xshard_types.GetShardCommonMsgsHash(msgs))
	}
	msgRoot := common.ComputeMerkleRoot(hashes)
	return msgRoot
}

func VerifyCrossShardMsg(shardID common.ShardID, sourceShardID common.ShardID, lgr *ledger.Ledger, crossShardMsgInfo *types.CrossShardMsgInfo, shardMsg []xshard_types.CommonShardMsg) bool {
	if !shardID.IsRootShard() {
		lgr = lgr.ParentLedger
	}
	chainconfig, err := csm.GetShardConfigByShardID(lgr, sourceShardID, crossShardMsgInfo.SignMsgHeight)
	if err != nil {
		log.Errorf("GetShardConfigByShardID shardID:%v,height:%d err:%s", sourceShardID, crossShardMsgInfo.SignMsgHeight, err)
		return false
	}
	var bookkeepers []keypair.PublicKey
	m := int(chainconfig.N - (chainconfig.N-1)/3)
	for _, peer := range chainconfig.Peers {
		pubkey, err := vconfig.Pubkey(peer.ID)
		if err != nil {
			log.Errorf("pubKey peer.PeerPubkey:%s, err:%s", peer.ID, err)
			return false
		}
		bookkeepers = append(bookkeepers, pubkey)
	}
	sigData := make([][]byte, 0)
	for _, sig := range crossShardMsgInfo.ShardMsgInfo.SigData {
		sigData = append(sigData, sig)
	}
	msgRoot := CrossShardMsgHash(crossShardMsgInfo, shardMsg)
	err = sign.VerifyMultiSignature(msgRoot[:], bookkeepers, m, sigData)
	if err != nil {
		log.Errorf("verifycrossshardMsg VerifyMultiSignature:%s,Bookkeepers:%d,pubkey:%d,signnum:%d", err, len(bookkeepers), m, len(sigData))
		return false
	}
	return true
}
