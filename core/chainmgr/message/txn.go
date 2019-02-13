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

package message

import (
	"encoding/json"
	"fmt"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/errors"
)

type ShardTxRequest interface {
	Type() int
	ShardID() uint64
	Done()
}

type ShardTxResponse interface {
	Type() int
}

type TxResult struct {
	Err  errors.ErrCode `json:"err"`
	Hash common.Uint256 `json:"hash"`
	Desc string         `json:"desc"`
}

func (this *TxResult) Type() int {
	return TXN_RSP_MSG
}

type TxRequest struct {
	Tx         *types.Transaction
	TxResultCh chan *TxResult
}

func (this *TxRequest) ShardID() uint64 {
	return this.Tx.ShardID
}

func (this *TxRequest) Type() int {
	return TXN_REQ_MSG
}

func (this *TxRequest) Done() {
	close(this.TxResultCh)
}

type TxRequestHelper struct {
	Payload []byte `json:"payload"`
}

func (this *TxRequest) MarshalJSON() ([]byte, error) {
	sink := common.NewZeroCopySink(nil)
	if this.Tx != nil {
		if err := this.Tx.Serialization(sink); err != nil {
			return nil, fmt.Errorf("TxRequest marshal: %s", err)
		}
	}
	result, err := json.Marshal(&TxRequestHelper{sink.Bytes()})
	return result, err
}

func (this *TxRequest) UnmarshalJSON(data []byte) error {
	helper := &TxRequestHelper{}
	if err := json.Unmarshal(data, helper); err != nil {
		return fmt.Errorf("TxRequest bytes unmarshal: %s", err)
	}
	tx := &types.Transaction{}
	if err := tx.Deserialization(common.NewZeroCopySource(helper.Payload)); err != nil {
		return fmt.Errorf("TxRequest unmarshal: %s", err)
	}

	this.Tx = tx
	return nil
}

type StorageResult struct {
	ShardID uint64         `json:"shard_id"`
	Address common.Address `json:"address"`
	Key     []byte         `json:"key"`
	Data    []byte         `json:"data"`
	Err     string         `json:"err"`
}

func (this *StorageResult) Type() int {
	return STORAGE_RSP_MSG
}

type StorageRequest struct {
	ShardId  uint64              `json:"shard_id"`
	Address  common.Address      `json:"address"`
	Key      []byte              `json:"key"`
	ResultCh chan *StorageResult `json:"-"`
}

func (this *StorageRequest) Type() int {
	return STORAGE_REQ_MSG
}

func (this *StorageRequest) ShardID() uint64 {
	return this.ShardId
}

func (this *StorageRequest) Done() {
	close(this.ResultCh)
}
