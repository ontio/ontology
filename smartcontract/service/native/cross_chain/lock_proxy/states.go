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

package lock_proxy

import (
	"fmt"
	"io"
	"math/big"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

// Args for lock and unlock
type Args struct {
	TargetAssetHash []byte // to contract asset hash
	ToAddress       []byte
	Value           uint64
}

func (this *Args) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes(this.TargetAssetHash)
	sink.WriteVarBytes(this.ToAddress)
	sink.WriteVarUint(this.Value)
}

func (this *Args) Deserialization(source *common.ZeroCopySource) error {
	assetHash, _, irregular, eof := source.NextVarBytes()
	if irregular {
		return fmt.Errorf("Args.Deserialization NextVarBytes AssetHash error")
	}
	if eof {
		return fmt.Errorf("Args.Deserialization NextVarBytes AssetHash error:%s", io.ErrUnexpectedEOF)
	}

	toAddress, _, irregular, eof := source.NextVarBytes()
	if irregular {
		return fmt.Errorf("Args.Deserialization NextVarBytes ToAddress error")
	}
	if eof {
		return fmt.Errorf("Args.Deserialization NextVarBytes ToAddress error:%s", io.ErrUnexpectedEOF)
	}

	value, _, irregular, eof := source.NextVarUint()
	if irregular {
		return fmt.Errorf("Args.Deserialization NextVarUint Value error")
	}
	if eof {
		return fmt.Errorf("Args.Deserialization NextVarUint Value error:%s", io.ErrUnexpectedEOF)
	}
	this.TargetAssetHash = assetHash
	this.ToAddress = toAddress
	this.Value = value
	return nil
}

type LockParam struct {
	SourceAssetHash common.Address
	FromAddress     common.Address
	ToChainID       uint64
	ToAddress       []byte
	Value           uint64
}

func (this *LockParam) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeAddress(sink, this.SourceAssetHash)
	utils.EncodeAddress(sink, this.FromAddress)
	utils.EncodeVarUint(sink, this.ToChainID)
	utils.EncodeVarBytes(sink, this.ToAddress)
	utils.EncodeVarUint(sink, this.Value)
}

func (this *LockParam) Deserialization(source *common.ZeroCopySource) error {
	var err error
	if this.SourceAssetHash, err = utils.DecodeAddress(source); err != nil {
		return fmt.Errorf("LockParam.Deserialization DecodeAddress SourceAssetHash error:%s", err)
	}
	if this.FromAddress, err = utils.DecodeAddress(source); err != nil {
		return fmt.Errorf("LockParam.Deserialization DecodeAddress FromAddress error:%s", err)
	}
	if this.ToChainID, err = utils.DecodeVarUint(source); err != nil {
		return fmt.Errorf("LockParam.Deserialization DecodeVarUint ToChainID error:%s", err)
	}
	if this.ToAddress, err = utils.DecodeVarBytes(source); err != nil {
		return fmt.Errorf("LockParam.Deserialization DecodeVarBytes ToAddress error:%s", err)
	}
	if this.Value, err = utils.DecodeVarUint(source); err != nil {
		return fmt.Errorf("LockParam.Deserialization DecodeVarUint Value error:%s", err)
	}
	return nil
}

type UnlockParam struct {
	ArgsBs             []byte
	FromContractHashBs []byte
	FromChainId        uint64
}

func (this *UnlockParam) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarBytes(sink, this.ArgsBs)
	utils.EncodeVarBytes(sink, this.FromContractHashBs)
	utils.EncodeVarUint(sink, this.FromChainId)
}

func (this *UnlockParam) Deserialization(source *common.ZeroCopySource) error {
	var err error
	if this.ArgsBs, err = utils.DecodeVarBytes(source); err != nil {
		return fmt.Errorf("UnlockParam.Deserialization DecodeVarBytes ArgsBs error:%s", err)
	}
	if this.FromContractHashBs, err = utils.DecodeVarBytes(source); err != nil {
		return fmt.Errorf("UnlockParam.Deserialization DecodeVarBytes FromContractHashBs error:%s", err)
	}
	if this.FromChainId, err = utils.DecodeVarUint(source); err != nil {
		return fmt.Errorf("UnlockParam.Deserialization DecodeVarUint FromChainId error:%s", err)
	}
	return nil
}

type BindProxyParam struct {
	TargetChainId uint64
	TargetHash    []byte
}

func (this *BindProxyParam) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarUint(sink, this.TargetChainId)
	utils.EncodeVarBytes(sink, this.TargetHash)
}

func (this *BindProxyParam) Deserialization(source *common.ZeroCopySource) error {
	var err error
	if this.TargetChainId, err = utils.DecodeVarUint(source); err != nil {
		return fmt.Errorf("BindProxyParam.Deserialization DecodeVarUint TargetChainId error:%s", err)
	}
	if this.TargetHash, err = utils.DecodeVarBytes(source); err != nil {
		return fmt.Errorf("BindProxyParam.Deserialization DecodeVarBytes TargetAssetHash error:%s", err)
	}
	return nil
}

type BindAssetParam struct {
	SourceAssetHash    common.Address
	TargetChainId      uint64
	TargetAssetHash    []byte
	Limit              *big.Int
	IsTargetChainAsset bool
}

func (this *BindAssetParam) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeAddress(sink, this.SourceAssetHash)
	utils.EncodeVarUint(sink, this.TargetChainId)
	utils.EncodeVarBytes(sink, this.TargetAssetHash)
	utils.EncodeVarBytes(sink, common.BigIntToNeoBytes(this.Limit))
	sink.WriteBool(this.IsTargetChainAsset)
}

func (this *BindAssetParam) Deserialization(source *common.ZeroCopySource) error {
	var err error
	if this.SourceAssetHash, err = utils.DecodeAddress(source); err != nil {
		return fmt.Errorf("BindAssetParam.Deserialization DecodeAddress SourceAssetAddress error:%s", err)
	}
	if this.TargetChainId, err = utils.DecodeVarUint(source); err != nil {
		return fmt.Errorf("BindAssetParam.Deserialization DecodeVarUint TargetChainId error:%s", err)
	}
	if this.TargetAssetHash, err = utils.DecodeVarBytes(source); err != nil {
		return fmt.Errorf("BindAssetParam.Deserialization DecodeVarBytes TargetAssetHash error:%s", err)
	}
	limitNeoBytes, err := utils.DecodeVarBytes(source)
	if err != nil {
		return fmt.Errorf("BindAssetParam.Deserialization DecodeVarBytes Limit error:%s", err)
	}
	this.Limit = common.BigIntFromNeoBytes(limitNeoBytes)
	if this.IsTargetChainAsset, err = utils.DecodeBool(source); err != nil {
		return fmt.Errorf("BindAssetParam.Deserialization DecodeBool IsTargetChainAsset error:%s", err)
	}
	return nil
}
