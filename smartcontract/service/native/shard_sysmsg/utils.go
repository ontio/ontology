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

package shardsysmsg

import (
	"bytes"
	"fmt"
	"io"

	"github.com/ontio/ontology/common"
	cstates "github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/utils"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

func appCallTransfer(native *native.NativeService, contract common.Address, from common.Address, to common.Address, amount uint64) error {
	var sts []ont.State
	sts = append(sts, ont.State{
		From:  from,
		To:    to,
		Value: amount,
	})
	transfers := ont.Transfers{
		States: sts,
	}
	sink := common.NewZeroCopySink(nil)
	transfers.Serialization(sink)

	if _, err := native.NativeCall(contract, "transfer", sink.Bytes()); err != nil {
		return fmt.Errorf("appCallTransfer, appCall error: %v", err)
	}
	return nil
}

type ToShardsInBlock struct {
	Shards []uint64 `json:"shards"`
}

func (this *ToShardsInBlock) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, this)
}

func (this *ToShardsInBlock) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, this)
}

func addToShardsInBlock(native *native.NativeService, toShard uint64) error {
	toShards, err := getToShardsInBlock(native, uint64(native.Height))
	if err != nil {
		return err
	}
	if toShards == nil {
		toShards = []uint64{toShard}
	} else {
		for _, s := range toShards {
			if s == toShard {
				// already in
				return nil
			}
		}
	}

	contract := native.ContextRef.CurrentContext().ContractAddress
	blockNumBytes, err := shardutil.GetUint64Bytes(uint64(native.Height))
	if err != nil {
		return fmt.Errorf("serialize height: %s", err)
	}

	toShardsInBlk := &ToShardsInBlock{
		Shards: toShards,
	}
	buf := new(bytes.Buffer)
	if err := toShardsInBlk.Serialize(buf); err != nil {
		return fmt.Errorf("serialize to-shards in block: %s", err)
	}

	key := utils.ConcatKey(contract, []byte(KEY_SHARDS_IN_BLOCK), blockNumBytes)
	native.CacheDB.Put(key, cstates.GenRawStorageItem(buf.Bytes()))
	return nil
}

func getToShardsInBlock(native *native.NativeService, blockNum uint64) ([]uint64, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	blockNumBytes, err := shardutil.GetUint64Bytes(blockNum)
	if err != nil {
		return nil, fmt.Errorf("serialize height: %s", err)
	}

	key := utils.ConcatKey(contract, []byte(KEY_SHARDS_IN_BLOCK), blockNumBytes)
	toShardsBytes, err := native.CacheDB.Get(key)
	if err != nil {
		return nil, fmt.Errorf("get toShards: %s", err)
	}
	if toShardsBytes == nil {
		// not found
		return nil, nil
	}

	value, err := cstates.GetValueFromRawStorageItem(toShardsBytes)
	if err != nil {
		return nil, fmt.Errorf("get toShards from bytes: %s", err)
	}

	req := &ToShardsInBlock{}
	if err := req.Deserialize(bytes.NewBuffer(value)); err != nil {
		return nil, fmt.Errorf("deserialize toShards: %s", err)
	}

	return req.Shards, nil
}

type ReqsInBlock struct {
	Reqs [][]byte `json:"reqs"`
}

func (this *ReqsInBlock) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, this)
}

func (this *ReqsInBlock) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, this)
}

func addReqsInBlock(native *native.NativeService, req *shardstates.CommonShardReq) error {
	reqs, err := getReqsInBlock(native, uint64(native.Height), req.TargetShard)
	if err != nil {
		return err
	}
	reqBytes := new(bytes.Buffer)
	if err := req.Serialize(reqBytes); err != nil {
		return err
	}
	if reqs == nil {
		reqs = [][]byte{reqBytes.Bytes()}
	}

	contract := native.ContextRef.CurrentContext().ContractAddress
	blockNumBytes, err := shardutil.GetUint64Bytes(uint64(native.Height))
	if err != nil {
		return fmt.Errorf("serialize height: %s", err)
	}
	shardIDBytes, err := shardutil.GetUint64Bytes(req.TargetShard)
	if err != nil {
		return fmt.Errorf("serialzie toshard: %s", err)
	}

	reqInBlk := &ReqsInBlock{
		Reqs: reqs,
	}
	buf := new(bytes.Buffer)
	if err := reqInBlk.Serialize(buf); err != nil {
		return fmt.Errorf("serialize shardmgmt global state: %s", err)
	}

	key := utils.ConcatKey(contract, []byte(KEY_REQS_IN_BLOCK), blockNumBytes, shardIDBytes)
	native.CacheDB.Put(key, cstates.GenRawStorageItem(buf.Bytes()))
	return nil
}

func getReqsInBlock(native *native.NativeService, blockNum uint64, shardID uint64) ([][]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	blockNumBytes, err := shardutil.GetUint64Bytes(blockNum)
	if err != nil {
		return nil, fmt.Errorf("serialize height: %s", err)
	}
	shardIDBytes, err := shardutil.GetUint64Bytes(shardID)
	if err != nil {
		return nil, fmt.Errorf("serialize toShard: %s", err)
	}

	key := utils.ConcatKey(contract, []byte(KEY_REQS_IN_BLOCK), blockNumBytes, shardIDBytes)
	reqBytes, err := native.CacheDB.Get(key)
	if err != nil {
		return nil, fmt.Errorf("get reqs in block: %s", err)
	}
	if reqBytes == nil {
		// not found
		return nil, nil
	}

	value, err := cstates.GetValueFromRawStorageItem(reqBytes)
	if err != nil {
		return nil, fmt.Errorf("get reqs from bytes: %s", err)
	}

	req := &ReqsInBlock{}
	if err := req.Deserialize(bytes.NewBuffer(value)); err != nil {
		return nil, fmt.Errorf("deserialize reqsInBlock: %s", err)
	}

	return req.Reqs, nil
}
