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
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ontio/ontology/common"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/chainmgr/message"
	"github.com/ontio/ontology/core/types"
	shardstates "github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
)

const JSON_RPC_VERSION = "2.0"

func (this *ChainManager) addShardBlockInfo(blkInfo *message.ShardBlockInfo) error {
	this.lock.Lock()
	defer this.lock.Unlock()

	return this.blockPool.AddBlockInfo(blkInfo)
}

func (this *ChainManager) getShardBlockInfo(shardID common.ShardID, height uint32) *message.ShardBlockInfo {
	this.lock.RLock()
	defer this.lock.RUnlock()

	return this.blockPool.GetBlockInfo(shardID, height)
}

func (this *ChainManager) updateShardBlockInfo(shardID common.ShardID, header *types.Header, shardTxs map[common.ShardID]*message.ShardBlockTx) {
	this.lock.Lock()
	defer this.lock.Unlock()

	blkInfo := this.blockPool.GetBlockInfo(shardID, block.Header.Height)
	if blkInfo == nil {
		return
	}

	blkInfo.Block = block
	blkInfo.ShardTxs = shardTxs
}

func (self *ChainManager) initShardInfo(shardID common.ShardID, shard *shardstates.ShardState) (*ShardInfo, error) {
	if shardID != shard.ShardID {
		return nil, fmt.Errorf("unmatched shard ID with shardstate")
	}

	info := &ShardInfo{}
	if i, present := self.shards[shard.ShardID]; present {
		info = i
	}
	info.ShardID = shard.ShardID

	seedList := make([]string, 0)
	for _, p := range shard.Peers {
		seedList = append(seedList, p.IpAddress)
	}
	info.SeedList = seedList
		self.shards[shard.ShardID] = info
	return info, nil
}

type JsonRpcRequest struct {
	Version string        `json:"jsonrpc"`
	Id      string        `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

func sendRawTx(tx *types.Transaction, shardPeerIp string, shardPort uint) error {
	var buffer bytes.Buffer
	err := tx.Serialize(&buffer)
	if err != nil {
		return fmt.Errorf("serialize error:%s", err)
	}
	reqAddr := fmt.Sprintf("http://%s:%d", shardPeerIp, shardPort)

	rpcReq := &JsonRpcRequest{
		Version: JSON_RPC_VERSION,
		Id:      "rpc",
		Method:  "sendrawtransaction",
		Params:  []interface{}{hex.EncodeToString(buffer.Bytes())},
	}
	reqData, err := json.Marshal(rpcReq)
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

	log.Debugf("chainmgr forward tx to %s", reqAddr)
	resp, err := httpClient.Post(reqAddr, "application/json", bytes.NewReader(reqData))
	if err != nil {
		return fmt.Errorf("send http post request error:%s", err)
	}
	defer resp.Body.Close()
	ioutil.ReadAll(resp.Body)
	return nil
}
