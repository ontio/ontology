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

package cross_chain_manager

import (
	"encoding/hex"
	"fmt"
	"github.com/ontio/ontology/common"

	"github.com/ontio/ontology/common/config"
	cstates "github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/merkle"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/service/native"
	ccom "github.com/ontio/ontology/smartcontract/service/native/cross_chain/common"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

func putDoneTx(native *native.NativeService, txHash []byte, chainID uint64) error {
	contract := utils.CrossChainContractAddress
	chainIDBytes, err := utils.GetUint64Bytes(chainID)
	if err != nil {
		return fmt.Errorf("putRequestID, get chainIDBytes error: %v", err)
	}
	native.CacheDB.Put(utils.ConcatKey(contract, []byte(DONE_TX), chainIDBytes, txHash), cstates.GenRawStorageItem(txHash))
	return nil
}

func checkDoneTx(native *native.NativeService, txHash []byte, chainID uint64) error {
	contract := utils.CrossChainContractAddress
	chainIDBytes, err := utils.GetUint64Bytes(chainID)
	if err != nil {
		return fmt.Errorf("checkDoneTx, get chainIDBytes error: %v", err)
	}
	value, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(DONE_TX), chainIDBytes, txHash))
	if err != nil {
		return fmt.Errorf("checkDoneTx, native.CacheDB.Get error: %v", err)
	}
	if value != nil {
		return fmt.Errorf("checkDoneTx, tx already done")
	}
	return nil
}

func putRequest(native *native.NativeService, txHash []byte, chainID uint64, request []byte) error {
	contract := utils.CrossChainContractAddress
	chainIDBytes, err := utils.GetUint64Bytes(chainID)
	if err != nil {
		return fmt.Errorf("putRequest, get chainIDBytes error: %v", err)
	}
	utils.PutBytes(native, utils.ConcatKey(contract, []byte(REQUEST), chainIDBytes, txHash), request)
	return nil
}

func MakeFromOntProof(native *native.NativeService, params *CreateCrossChainTxParam) error {
	//record cross chain tx
	txHash := native.Tx.Hash()
	merkleValue := &ccom.MakeTxParam{
		TxHash:              txHash.ToArray(),
		FromContractAddress: native.ContextRef.CallingContext().ContractAddress[:],
		ToChainID:           params.ToChainID,
		ToContractAddress:   params.ToContractAddress,
		Method:              params.Method,
		Args:                params.Args,
	}
	sink := common.NewZeroCopySink(nil)
	merkleValue.Serialization(sink)
	err := putRequest(native, merkleValue.TxHash, params.ToChainID, sink.Bytes())
	if err != nil {
		return fmt.Errorf("MakeFromOntProof, putRequest error:%s", err)
	}
	native.ContextRef.PutMerkleVal(sink.Bytes())
	chainIDBytes, err := utils.GetUint64Bytes(params.ToChainID)
	if err != nil {
		return fmt.Errorf("MakeFromOntProof, get chainIDBytes error: %v", err)
	}
	key := hex.EncodeToString(utils.ConcatKey(utils.CrossChainContractAddress, []byte(REQUEST), chainIDBytes, merkleValue.TxHash))
	args := hex.EncodeToString(params.Args)
	notifyMakeFromOntProof(native, hex.EncodeToString(merkleValue.TxHash), params.ToChainID, key,
		hex.EncodeToString(merkleValue.FromContractAddress), args)
	return nil
}

func VerifyToOntTx(native *native.NativeService, proof []byte, fromChainid uint64, header *ccom.Header) (*ccom.ToMerkleValue, error) {
	v, err := merkle.MerkleProve(proof, header.CrossStatesRoot.ToArray())
	if err != nil {
		return nil, fmt.Errorf("VerifyToOntTx, merkle.MerkleProve verify merkle proof error: %v", err)
	}

	s := common.NewZeroCopySource(v)
	merkleValue := new(ccom.ToMerkleValue)
	if err := merkleValue.Deserialization(s); err != nil {
		return nil, fmt.Errorf("VerifyToOntTx, deserialize merkleValue error:%s", err)
	}

	//record done cross chain tx
	err = checkDoneTx(native, merkleValue.TxHash, fromChainid)
	if err != nil {
		return nil, fmt.Errorf("VerifyToOntTx, checkDoneTx error:%s", err)
	}
	err = putDoneTx(native, merkleValue.TxHash, fromChainid)
	if err != nil {
		return nil, fmt.Errorf("VerifyToOntTx, putDoneTx error:%s", err)
	}

	notifyVerifyToOntProof(native, hex.EncodeToString(merkleValue.TxHash), hex.EncodeToString(merkleValue.MakeTxParam.TxHash),
		fromChainid, hex.EncodeToString(merkleValue.MakeTxParam.ToContractAddress))
	return merkleValue, nil
}

func notifyMakeFromOntProof(native *native.NativeService, txHash string, toChainID uint64, key string, contract, args string) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			ContractAddress: utils.CrossChainContractAddress,
			States:          []interface{}{MAKE_FROM_ONT_PROOF, txHash, toChainID, native.Height, key, contract, args},
		})
}

func notifyVerifyToOntProof(native *native.NativeService, txHash, rawTxHash string, fromChainID uint64, contract string) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			ContractAddress: utils.CrossChainContractAddress,
			States:          []interface{}{VERIFY_TO_ONT_PROOF, txHash, rawTxHash, fromChainID, native.Height, contract},
		})
}
