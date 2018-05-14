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

package oracle

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/genesis"
	cstates "github.com/ontio/ontology/core/states"
	scommon "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/governance"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

const (
	//status
	RegisterOracleNodeStatus Status = iota
	OracleNodeStatus
)

const (
	//function name
	REGISTER_ORACLE_NODE  = "registerOracleNode"
	APPROVE_ORACLE_NODE   = "approveOracleNode"
	QUIT_ORACLE_NODE      = "quitOracleNode"
	CREATE_ORACLE_REQUEST = "createOracleRequest"
	SET_ORACLE_OUTCOME    = "setOracleOutcome"

	//keyPrefix
	ORACLE_NODE    = "OracleNode"
	UNDO_TXHASH    = "UndoTxHash"
	REQUEST        = "Request"
	OUTCOME_RECORD = "OutcomeRecord"

	//global
	MIN_GUARANTY    = 1000
	ORACLE_NODE_FEE = 500
)

func InitOracle() {
	native.Contracts[genesis.OracleContractAddress] = RegisterOracleContract
}

func RegisterOracleContract(native *native.NativeService) {
	native.Register(REGISTER_ORACLE_NODE, RegisterOracleNode)
	native.Register(APPROVE_ORACLE_NODE, ApproveOracleNode)
	native.Register(QUIT_ORACLE_NODE, QuitOracleNode)
	native.Register(CREATE_ORACLE_REQUEST, CreateOracleRequest)
	native.Register(SET_ORACLE_OUTCOME, SetOracleOutcome)
}

func RegisterOracleNode(native *native.NativeService) ([]byte, error) {
	params := new(RegisterOracleNodeParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, contract params deserialize error!")
	}

	//check witness
	err := utils.ValidateOwner(native, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "validateOwner, checkWitness error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//check Guaranty
	if params.Guaranty < MIN_GUARANTY {
		return utils.BYTE_FALSE, errors.NewErr(fmt.Sprintf("registerOracleNode, guaranty must >= %v!", MIN_GUARANTY))
	}

	addressBytes, err := hex.DecodeString(params.Address)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, address format error!")
	}
	address, err := common.AddressParseFromBytes(addressBytes)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "common.AddressParseFromBytes, address format error!")
	}

	oracleNodeBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(ORACLE_NODE), addressBytes))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Get, get oracleNodeBytes error!")
	}
	if oracleNodeBytes != nil {
		return utils.BYTE_FALSE, errors.NewErr("registerOracleNode, oracleNode is already registered!")
	}

	oracleNode := &OracleNode{
		Address:  params.Address,
		Guaranty: params.Guaranty,
		Status:   RegisterOracleNodeStatus,
	}
	bf := new(bytes.Buffer)
	if err := oracleNode.Serialize(bf); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize oracleNode error!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(ORACLE_NODE), addressBytes), &cstates.StorageItem{Value: bf.Bytes()})

	//ont transfer
	err = governance.AppCallTransferOnt(native, address, genesis.OracleContractAddress, params.Guaranty)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferOnt, ont transfer error!")
	}
	//ong transfer
	err = governance.AppCallTransferOng(native, address, genesis.FeeSplitContractAddress, ORACLE_NODE_FEE)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferOng, ong transfer error!")
	}

	utils.AddCommonEvent(native, contract, REGISTER_ORACLE_NODE, params)

	return utils.BYTE_TRUE, nil
}

func ApproveOracleNode(native *native.NativeService) ([]byte, error) {
	params := new(ApproveOracleNodeParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, contract params deserialize error!")
	}

	//TODO: check witness
	//err = validateOwner(native, params.Address)
	//if err != nil {
	//	return errors.NewDetailErr(err, errors.ErrNoCode, "[registerOracleNode] CheckWitness error!")
	//}
	contract := native.ContextRef.CurrentContext().ContractAddress

	oracleNode := new(OracleNode)
	addressBytes, err := hex.DecodeString(params.Address)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, address format error!")
	}
	oracleNodeBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(ORACLE_NODE), addressBytes))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Get, get oracleNodeBytes error!")
	}
	if oracleNodeBytes == nil {
		return utils.BYTE_FALSE, errors.NewErr("approveOracleNode, oracleNode is not registered!")
	}
	oracleNodeStore, _ := oracleNodeBytes.(*cstates.StorageItem)
	if err := oracleNode.Deserialize(bytes.NewBuffer(oracleNodeStore.Value)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize oracleNode error!")
	}
	if oracleNode.Status != RegisterOracleNodeStatus {
		return utils.BYTE_FALSE, errors.NewErr("approveOracleNode, oracleNode status is not RegisterOracleNodeStatus!")
	}
	oracleNode.Status = OracleNodeStatus

	bf := new(bytes.Buffer)
	if err := oracleNode.Serialize(bf); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize oracleNode error!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(ORACLE_NODE), addressBytes), &cstates.StorageItem{Value: bf.Bytes()})

	utils.AddCommonEvent(native, contract, APPROVE_ORACLE_NODE, params)

	return utils.BYTE_TRUE, nil
}

func QuitOracleNode(native *native.NativeService) ([]byte, error) {
	params := new(QuitOracleNodeParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, contract params deserialize error!")
	}

	//TODO: check witness
	//err = validateOwner(native, params.Address)
	//if err != nil {
	//	return errors.NewDetailErr(err, errors.ErrNoCode, "[registerOracleNode] CheckWitness error!")
	//}
	contract := native.ContextRef.CurrentContext().ContractAddress

	oracleNode := new(OracleNode)
	addressBytes, err := hex.DecodeString(params.Address)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, address format error!")
	}
	address, err := common.AddressParseFromBytes(addressBytes)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "common.AddressParseFromBytes, address format error!")
	}
	oracleNodeBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(ORACLE_NODE), addressBytes))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Get, get oracleNodeBytes error!")
	}
	if oracleNodeBytes == nil {
		return utils.BYTE_FALSE, errors.NewErr("quitOracleNode, oracleNode is not registered!")
	}
	oracleNodeStore, _ := oracleNodeBytes.(*cstates.StorageItem)
	if err := oracleNode.Deserialize(bytes.NewBuffer(oracleNodeStore.Value)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize oracleNode error!")
	}

	//ont transfer
	err = governance.AppCallTransferOnt(native, address, genesis.OracleContractAddress, oracleNode.Guaranty)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferOnt, ont transfer error!")
	}

	native.CloneCache.Delete(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(ORACLE_NODE), addressBytes))

	utils.AddCommonEvent(native, contract, QUIT_ORACLE_NODE, params)

	return utils.BYTE_TRUE, nil
}

func CreateOracleRequest(native *native.NativeService) ([]byte, error) {
	params := new(CreateOracleRequestParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, contract params deserialize error!")
	}

	//check witness
	err := utils.ValidateOwner(native, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "validateOwner, validateOwner error!")
	}

	contract := native.ContextRef.CurrentContext().ContractAddress
	txHash := native.Tx.Hash()
	txHashBytes := txHash.ToArray()
	txHashHex := hex.EncodeToString(txHashBytes)
	undoRequests := &UndoRequests{
		Requests: make(map[string]struct{}),
	}

	undoRequestsBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(UNDO_TXHASH)))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Get, get UndoRequests error!")
	}

	if undoRequestsBytes != nil {
		item, _ := undoRequestsBytes.(*cstates.StorageItem)
		err = json.Unmarshal(item.Value, undoRequests)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "json.Unmarshal, unmarshal UndoRequests error")
		}
	}

	undoRequests.Requests[txHashHex] = struct{}{}

	value, err := json.Marshal(undoRequests)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "json.Marshal, marshal UndoRequests error")
	}

	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(UNDO_TXHASH)), &cstates.StorageItem{Value: value})
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(REQUEST), txHashBytes), &cstates.StorageItem{Value: native.Input})

	addressBytes, err := hex.DecodeString(params.Address)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, address format error!")
	}
	address, err := common.AddressParseFromBytes(addressBytes)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "common.AddressParseFromBytes, address format error!")
	}
	//ong transfer
	err = governance.AppCallTransferOng(native, address, genesis.OracleContractAddress, params.Fee)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferOng, ong transfer error!")
	}

	utils.AddCommonEvent(native, contract, CREATE_ORACLE_REQUEST, params)

	return utils.BYTE_TRUE, nil
}

func SetOracleOutcome(native *native.NativeService) ([]byte, error) {
	params := new(SetOracleOutcomeParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, contract params deserialize error!")
	}

	//check witness
	err := utils.ValidateOwner(native, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "validateOwner, validateOwner error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	oracleNode := new(OracleNode)
	addressBytes, err := hex.DecodeString(params.Address)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, address format error!")
	}
	oracleNodeBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(ORACLE_NODE), addressBytes))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Get, get oracleNodeBytes error!")
	}
	if oracleNodeBytes == nil {
		return utils.BYTE_FALSE, errors.NewErr("setOracleOutcome, oracleNode is not registered!")
	}
	oracleNodeStore, _ := oracleNodeBytes.(*cstates.StorageItem)
	if err := oracleNode.Deserialize(bytes.NewBuffer(oracleNodeStore.Value)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize oracleNode error!")
	}
	if oracleNode.Status != OracleNodeStatus {
		return utils.BYTE_FALSE, errors.NewErr("setOracleOutcome, oracleNode is not approved!")
	}

	txHashHex := params.TxHash
	txHash, err := hex.DecodeString(txHashHex)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, decode hex txHash error!")
	}
	//get request
	request := new(CreateOracleRequestParam)
	requestBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(REQUEST), txHash))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Get, get requestBytes error!")
	}
	if requestBytes == nil {
		return utils.BYTE_FALSE, errors.NewErr("setOracleOutcome, request is not exist!")
	}
	requestStore, _ := requestBytes.(*cstates.StorageItem)
	if err := request.Deserialize(bytes.NewBuffer(requestStore.Value)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize request error!")
	}
	if request.OracleNode != params.Address {
		return utils.BYTE_FALSE, errors.NewErr("setOracleOutcome, request is not assigned to this address!")
	}

	outcomeRecordBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(OUTCOME_RECORD), txHash))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Get, get OutcomeRecord error!")
	}
	if outcomeRecordBytes != nil {
		return utils.BYTE_FALSE, errors.NewErr("setOracleOutcome, OutcomeRecord is already set!")
	}

	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(OUTCOME_RECORD), txHash), &cstates.StorageItem{Value: []byte(params.Outcome)})

	//remove txHash from undoRequests
	undoRequests := &UndoRequests{
		Requests: make(map[string]struct{}),
	}

	undoRequestsBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(UNDO_TXHASH)))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Get, get UndoRequests error!")
	}

	if undoRequestsBytes != nil {
		item, _ := undoRequestsBytes.(*cstates.StorageItem)
		err = json.Unmarshal(item.Value, undoRequests)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "json.Unmarshal, unmarshal UndoRequests error")
		}
	}
	delete(undoRequests.Requests, params.TxHash)
	value, err := json.Marshal(undoRequests)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "json.Marshal, marshal UndoRequests error")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(UNDO_TXHASH)), &cstates.StorageItem{Value: value})

	address, err := common.AddressParseFromBytes(addressBytes)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "common.AddressParseFromBytes, address format error!")
	}
	//ong transfer
	err = governance.AppCallTransferOng(native, genesis.OracleContractAddress, address, request.Fee)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferOng, ong transfer error!")
	}

	utils.AddCommonEvent(native, contract, SET_ORACLE_OUTCOME, params)

	return utils.BYTE_TRUE, nil
}
