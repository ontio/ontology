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

package chainmgr

import (
	"encoding/hex"
	"fmt"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/core/chainmgr/message"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
	"bytes"
	"encoding/json"
	"time"
	"net/http"
	"io/ioutil"
)

func (this *ChainManager) addShardBlockInfo(blkInfo *message.ShardBlockInfo) error {
	this.lock.Lock()
	defer this.lock.Unlock()

	if err := this.blockPool.AddBlock(blkInfo); err != nil {
		return err
	}

	return nil
}

func (this *ChainManager) getShardBlockInfo(shardID uint64, height uint64) *message.ShardBlockInfo {
	this.lock.RLock()
	defer this.lock.RUnlock()

	return this.blockPool.GetBlock(shardID, height)
}

func (this *ChainManager) addShardEvent(evt *shardstates.ShardEventState) error {
	this.lock.Lock()
	defer this.lock.Unlock()

	if err := this.blockPool.AddEvent(this.shardID, evt); err != nil {
		return err
	}
	return nil
}

func (this *ChainManager) updateShardBlockInfo(shardID uint64, height uint64, blk *types.Block, shardTxs map[uint64]*message.ShardBlockTx) {
	this.lock.Lock()
	defer this.lock.Unlock()

	blkInfo := this.blockPool.GetBlock(shardID, height)
	if blkInfo == nil {
		return
	}

	blkInfo.Header = &message.ShardBlockHeader{Header: blk.Header}
	blkInfo.ShardTxs = shardTxs
}

func (this *ChainManager) getChildShards() map[uint64]*ShardInfo {

	shards := make(map[uint64]*ShardInfo)

	for _, shardInfo := range this.shards {
		if shardInfo.ConnType == CONN_TYPE_CHILD {
			shards[shardInfo.ShardID] = shardInfo
		}
	}

	return shards
}

func (self *ChainManager) initShardInfo(shardID uint64, shard *shardstates.ShardState) (*ShardInfo, error) {
	if shardID != shard.ShardID {
		return nil, fmt.Errorf("unmatched shard ID with shardstate")
	}

	peerPK := hex.EncodeToString(keypair.SerializePublicKey(self.account.PublicKey))
	info := &ShardInfo{}
	if i, present := self.shards[shard.ShardID]; present {
		info = i
	}
	info.ShardID = shard.ShardID
	info.ParentShardID = shard.ParentShardID

	if _, present := shard.Peers[peerPK]; present {
		// peer is in the shard
		// build shard config
		if self.shardID == shard.ShardID {
			// self shards
			info.ConnType = CONN_TYPE_SELF
		} else if self.parentShardID == shard.ShardID {
			// parent shard
			info.ConnType = CONN_TYPE_PARENT
		} else if self.shardID == shard.ParentShardID {
			// child shard
			info.ConnType = CONN_TYPE_CHILD
		}
	} else {
		if self.shardID == shard.ParentShardID {
			// child shards
			info.ConnType = CONN_TYPE_CHILD
		} else if self.parentShardID == shard.ParentShardID {
			// sib shards
			info.ConnType = CONN_TYPE_SIB
		}
	}

	if info.ConnType != CONN_TYPE_UNKNOWN {
		self.shards[shard.ShardID] = info
	}
	return info, nil
}

type RestfulReq struct {
	Action  string
	Version string
	Type    int
	Data    string
}

func sendRawTx(tx *types.Transaction) error {
	var buffer bytes.Buffer
	err := tx.Serialize(&buffer)
	if err != nil {
		return fmt.Errorf("Serialize error:%s", err)
	}
	reqUrl := "http://127.0.0.1:20334/api/v1/transaction"
	restReq := &RestfulReq{
		Action:  "sendrawtransaction",
		Version: "1.0.0",
		Data:    hex.EncodeToString(buffer.Bytes()),
	}
	reqData, err := json.Marshal(restReq)
	if err != nil {
		return fmt.Errorf("json.Marshal error:%s", err)
	}
	httpClient := &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   5,
			IdleConnTimeout:       time.Second * 300,
			ResponseHeaderTimeout: time.Second * 300,
		},
		Timeout: time.Second * 300, //timeout for http response
	}

	resp, err := httpClient.Post(reqUrl, "application/json", bytes.NewReader(reqData))
	if err != nil {
		return fmt.Errorf("send http post request error:%s", err)
	}
	defer resp.Body.Close()
	ioutil.ReadAll(resp.Body)
	return nil
}