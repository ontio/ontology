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
	"encoding/hex"
	"encoding/json"
	"math/big"

	"fmt"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/genesis"
	cstates "github.com/ontio/ontology/core/states"
	scommon "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

const (
	//status
	RegisterOracleNodeStatus Status = iota
	OracleNodeStatus
)

const (
	//function name
	REGISTER_ORACLE_NODE    = "registerOracleNode"
	APPROVE_ORACLE_NODE     = "approveOracleNode"
	QUIT_ORACLE_NODE        = "quitOracleNode"
	CREATE_ORACLE_REQUEST   = "createOracleRequest"
	SET_ORACLE_OUTCOME      = "setOracleOutcome"
	SET_ORACLE_CRON_OUTCOME = "setOracleCronOutcome"
	CHANGE_CRON_VIEW        = "changeCronView"

	//keyPrefix
	ORACLE_NODE         = "OracleNode"
	UNDO_TXHASH         = "UndoTxHash"
	ORACLE_NUM          = "OracleNum"
	REQUEST             = "Request"
	OUTCOME_RECORD      = "OutcomeRecord"
	FINAL_OUTCOME       = "FinalOutcome"
	CRON_VIEW           = "CronView"
	CRON_OUTCOME_RECORD = "CronOutcomeRecord"
	FINAL_CRON_OUTCOME  = "FinalCronOutcome"

	//global
	MIN_GUARANTY    = 1000
	ORACLE_NODE_FEE = 500
)

func init() {
	native.Contracts[genesis.OracleContractAddress] = RegisterOracleContract
}

func RegisterOracleContract(native *native.NativeService) {
	native.Register(REGISTER_ORACLE_NODE, RegisterOracleNode)
	native.Register(APPROVE_ORACLE_NODE, ApproveOracleNode)
	native.Register(QUIT_ORACLE_NODE, QuitOracleNode)
	native.Register(CREATE_ORACLE_REQUEST, CreateOracleRequest)
	native.Register(SET_ORACLE_OUTCOME, SetOracleOutcome)
	native.Register(SET_ORACLE_CRON_OUTCOME, SetOracleCronOutcome)
	native.Register(CHANGE_CRON_VIEW, ChangeCronView)
}

func RegisterOracleNode(native *native.NativeService) ([]byte, error) {
	params := new(RegisterOracleNodeParam)
	err := json.Unmarshal(native.Input, params)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Unmarshal] Contract params Unmarshal error!")
	}

	//check witness
	err = utils.ValidateOwner(native, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[ValidateOwner] CheckWitness error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//check Guaranty
	if params.Guaranty < MIN_GUARANTY {
		return utils.BYTE_FALSE, errors.NewErr(fmt.Sprintf("[RegisterOracleNode] Guaranty must >= %v!", MIN_GUARANTY))
	}

	addressBytes, err := hex.DecodeString(params.Address)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[hex.DecodeString] Address format error!")
	}
	address, err := common.AddressParseFromBytes(addressBytes)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[common.AddressParseFromBytes] Address format error!")
	}

	oracleNodeBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(ORACLE_NODE), addressBytes))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[native.CloneCache.Get] Get oracleNodeBytes error!")
	}
	if oracleNodeBytes != nil {
		return utils.BYTE_FALSE, errors.NewErr("[RegisterOracleNode] oracleNode is already registered!")
	}

	oracleNode := &OracleNode{
		Address:  params.Address,
		Guaranty: params.Guaranty,
		Status:   RegisterOracleNodeStatus,
	}
	value, err := json.Marshal(oracleNode)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Marshal] Marshal oracleNode error")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(ORACLE_NODE), addressBytes), &cstates.StorageItem{Value: value})

	//ont transfer
	err = utils.AppCallTransferOnt(native, address, genesis.OracleContractAddress, params.Guaranty)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[AppCallTransferOnt] Ont transfer error!")
	}
	//ong transfer
	err = utils.AppCallApproveOng(native, address, genesis.FeeSplitContractAddress, ORACLE_NODE_FEE)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[appCallTransferOng] Ong transfer error!")
	}

	utils.AddCommonEvent(native, contract, REGISTER_ORACLE_NODE, params)

	return utils.BYTE_TRUE, nil
}

func ApproveOracleNode(native *native.NativeService) ([]byte, error) {
	params := new(ApproveOracleNodeParam)
	err := json.Unmarshal(native.Input, params)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Unmarshal] Contract params Unmarshal error!")
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
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[hex.DecodeString] Address format error!")
	}
	oracleNodeBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(ORACLE_NODE), addressBytes))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[native.CloneCache.Get] Get oracleNodeBytes error!")
	}
	if oracleNodeBytes == nil {
		return utils.BYTE_FALSE, errors.NewErr("[ApproveOracleNode] oracleNode is not registered!")
	}
	oracleNodeStore, _ := oracleNodeBytes.(*cstates.StorageItem)
	err = json.Unmarshal(oracleNodeStore.Value, oracleNode)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Unmarshal] Unmarshal oracleNode error!")
	}
	if oracleNode.Status != RegisterOracleNodeStatus {
		return utils.BYTE_FALSE, errors.NewErr("[ApproveOracleNode] oracleNode status is not RegisterOracleNodeStatus!")
	}
	oracleNode.Status = OracleNodeStatus

	value, err := json.Marshal(oracleNode)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Marshal] Marshal oracleNode error")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(ORACLE_NODE), addressBytes), &cstates.StorageItem{Value: value})

	utils.AddCommonEvent(native, contract, APPROVE_ORACLE_NODE, params)

	return utils.BYTE_TRUE, nil
}

func QuitOracleNode(native *native.NativeService) ([]byte, error) {
	params := new(QuitOracleNodeParam)
	err := json.Unmarshal(native.Input, params)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Unmarshal] Contract params Unmarshal error!")
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
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[hex.DecodeString] Address format error!")
	}
	address, err := common.AddressParseFromBytes(addressBytes)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[common.AddressParseFromBytes] Address format error!")
	}
	oracleNodeBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(ORACLE_NODE), addressBytes))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[native.CloneCache.Get] Get oracleNodeBytes error!")
	}
	if oracleNodeBytes == nil {
		return utils.BYTE_FALSE, errors.NewErr("[QuitOracleNode] oracleNode is not registered!")
	}
	oracleNodeStore, _ := oracleNodeBytes.(*cstates.StorageItem)
	err = json.Unmarshal(oracleNodeStore.Value, oracleNode)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Unmarshal] Unmarshal oracleNode error!")
	}

	//ont transfer
	err = utils.AppCallTransferOnt(native, address, genesis.OracleContractAddress, oracleNode.Guaranty)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[appCallTransferOnt] Ont transfer error!")
	}

	native.CloneCache.Delete(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(ORACLE_NODE), addressBytes))

	utils.AddCommonEvent(native, contract, QUIT_ORACLE_NODE, params)

	return utils.BYTE_TRUE, nil
}

func CreateOracleRequest(native *native.NativeService) ([]byte, error) {
	params := new(CreateOracleRequestParam)
	err := json.Unmarshal(native.Input, params)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Unmarshal] Contract params Unmarshal error!")
	}

	//check witness
	err = utils.ValidateOwner(native, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[ValidateOwner] validateOwner error!")
	}

	if params.OracleNum.Sign() <= 0 {
		return utils.BYTE_FALSE, errors.NewErr("[CreateOracleRequest] OracleNum must be positive!")
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
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[native.CloneCache.Get] Get UndoRequests error!")
	}

	if undoRequestsBytes != nil {
		item, _ := undoRequestsBytes.(*cstates.StorageItem)
		err = json.Unmarshal(item.Value, undoRequests)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Unmarshal] Unmarshal UndoRequests error")
		}
	}

	undoRequests.Requests[txHashHex] = struct{}{}

	value, err := json.Marshal(undoRequests)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Marshal] Marshal UndoRequests error")
	}

	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(ORACLE_NUM), txHashBytes), &cstates.StorageItem{Value: params.OracleNum.Bytes()})
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(UNDO_TXHASH)), &cstates.StorageItem{Value: value})
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(REQUEST), txHashBytes), &cstates.StorageItem{Value: native.Input})

	utils.AddCommonEvent(native, contract, CREATE_ORACLE_REQUEST, params.Request)

	return utils.BYTE_TRUE, nil
}

func SetOracleOutcome(native *native.NativeService) ([]byte, error) {
	params := new(SetOracleOutcomeParam)
	err := json.Unmarshal(native.Input, params)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Unmarshal] Contract params Unmarshal error!")
	}

	//check witness
	err = utils.ValidateOwner(native, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[ValidateOwner] validateOwner error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	oracleNode := new(OracleNode)
	addressBytes, err := hex.DecodeString(params.Address)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[hex.DecodeString] Address format error!")
	}
	oracleNodeBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(ORACLE_NODE), addressBytes))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[native.CloneCache.Get] Get oracleNodeBytes error!")
	}
	if oracleNodeBytes == nil {
		return utils.BYTE_FALSE, errors.NewErr("[SetOracleOutcome] oracleNode is not registered!")
	}
	oracleNodeStore, _ := oracleNodeBytes.(*cstates.StorageItem)
	err = json.Unmarshal(oracleNodeStore.Value, oracleNode)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Unmarshal] Unmarshal oracleNode error!")
	}
	if oracleNode.Status != OracleNodeStatus {
		return utils.BYTE_FALSE, errors.NewErr("[SetOracleOutcome] oracleNode is not approved!")
	}

	txHashHex := params.TxHash
	txHash, err := hex.DecodeString(txHashHex)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[hex.DecodeString] Decode hex txHash error!")
	}

	outcomeRecord := &OutcomeRecord{
		OutcomeRecord: make(map[string]interface{}),
	}

	outcomeRecordBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(OUTCOME_RECORD), txHash))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[native.CloneCache.Get] Get OutcomeRecord error!")
	}

	if outcomeRecordBytes != nil {
		item, _ := outcomeRecordBytes.(*cstates.StorageItem)
		err = json.Unmarshal(item.Value, outcomeRecord)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Unmarshal] Unmarshal OutcomeRecord error")
		}
	}

	num := new(big.Int).SetInt64(int64(len(outcomeRecord.OutcomeRecord)))
	oracleNum, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(ORACLE_NUM), txHash))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[native.CloneCache.Get] Get OracleNum error!")
	}
	if oracleNum == nil {
		return utils.BYTE_FALSE, errors.NewErr("[SetOracleOutcome] Get nil OracleNum, check input txHash!")
	}
	item, _ := oracleNum.(*cstates.StorageItem)
	quorum := new(big.Int).SetBytes(item.Value)

	//quorum achieved
	if num.Cmp(quorum) >= 0 {
		return utils.BYTE_FALSE, errors.NewErr("[SetOracleOutcome] Request have achieved quorum")
	}

	//save new outcomeRecord
	_, ok := outcomeRecord.OutcomeRecord[params.Address]
	if ok {
		return utils.BYTE_FALSE, errors.NewErr("[SetOracleOutcome] Address has already setOutcome")
	}

	outcomeRecord.OutcomeRecord[params.Address] = params.Outcome
	value, err := json.Marshal(outcomeRecord)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Marshal] Marshal OutcomeRecord error")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(OUTCOME_RECORD), txHash), &cstates.StorageItem{Value: value})

	newNum := new(big.Int).SetInt64(int64(len(outcomeRecord.OutcomeRecord)))

	//quorum achieved
	if newNum.Cmp(quorum) == 0 {
		//remove txHash from undoRequests
		undoRequests := &UndoRequests{
			Requests: make(map[string]struct{}),
		}

		undoRequestsBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(UNDO_TXHASH)))
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[native.CloneCache.Get] Get UndoRequests error!")
		}

		if undoRequestsBytes != nil {
			item, _ := undoRequestsBytes.(*cstates.StorageItem)
			err = json.Unmarshal(item.Value, undoRequests)
			if err != nil {
				return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Unmarshal] Unmarshal UndoRequests error")
			}
		}
		delete(undoRequests.Requests, params.TxHash)
		value, err := json.Marshal(undoRequests)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Marshal] Marshal UndoRequests error")
		}
		native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(UNDO_TXHASH)), &cstates.StorageItem{Value: value})

		//aggregate result
		consensus := true
		for _, v := range outcomeRecord.OutcomeRecord {
			if params.Outcome != v {
				consensus = false
			}
		}
		if consensus {
			finalValue, err := json.Marshal(params.Outcome)
			if err != nil {
				return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Marshal] Marshal FinalOutcome error")
			}
			native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(FINAL_OUTCOME), txHash), &cstates.StorageItem{Value: finalValue})
		}

	}
	utils.AddCommonEvent(native, contract, SET_ORACLE_OUTCOME, params)

	return utils.BYTE_TRUE, nil
}

func SetOracleCronOutcome(native *native.NativeService) ([]byte, error) {
	params := new(SetOracleCronOutcomeParam)
	err := json.Unmarshal(native.Input, params)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Unmarshal] Contract params Unmarshal error!")
	}

	//check witness
	err = utils.ValidateOwner(native, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[validateOwner] validateOwner error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	oracleNode := new(OracleNode)
	addressBytes, err := hex.DecodeString(params.Address)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[hex.DecodeString] Address format error!")
	}
	oracleNodeBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(ORACLE_NODE), addressBytes))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[native.CloneCache.Get] Get oracleNodeBytes error!")
	}
	if oracleNodeBytes == nil {
		return utils.BYTE_FALSE, errors.NewErr("[SetOracleCronOutcome] oracleNode is not registered!")
	}
	oracleNodeStore, _ := oracleNodeBytes.(*cstates.StorageItem)
	err = json.Unmarshal(oracleNodeStore.Value, oracleNode)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Unmarshal] Unmarshal oracleNode error!")
	}
	if oracleNode.Status != OracleNodeStatus {
		return utils.BYTE_FALSE, errors.NewErr("[SetOracleCronOutcome] oracleNode is not approved!")
	}

	txHashHex := params.TxHash
	txHash, err := hex.DecodeString(txHashHex)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[hex.DecodeString] Decode hex txHash error!")
	}

	cronOutcomeRecord := &CronOutcomeRecord{
		CronOutcomeRecord: make(map[string]interface{}),
	}

	cronViewBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(CRON_VIEW), txHash))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[native.CloneCache.Get] Get CronView error!")
	}
	var cronView *big.Int
	if cronViewBytes == nil {
		cronView = new(big.Int).SetInt64(1)
	} else {
		cronViewStore, _ := cronViewBytes.(*cstates.StorageItem)
		cronView = new(big.Int).SetBytes(cronViewStore.Value)
	}

	outcomeRecordBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(CRON_OUTCOME_RECORD), txHash, cronView.Bytes()))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[native.CloneCache.Get] Get CronOutcomeRecord error!")
	}

	if outcomeRecordBytes != nil {
		item, _ := outcomeRecordBytes.(*cstates.StorageItem)
		err = json.Unmarshal(item.Value, cronOutcomeRecord)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Unmarshal] Unmarshal CronOutcomeRecord error")
		}
	}

	num := new(big.Int).SetInt64(int64(len(cronOutcomeRecord.CronOutcomeRecord)))
	oracleNum, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(ORACLE_NUM), txHash))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[native.CloneCache.Get] Get OracleNum error!")
	}
	if oracleNum == nil {
		return utils.BYTE_FALSE, errors.NewErr("[SetOracleCronOutcome] Get nil OracleNum, check input txHash!")
	}
	item, _ := oracleNum.(*cstates.StorageItem)
	quorum := new(big.Int).SetBytes(item.Value)

	//quorum achieved
	if num.Cmp(quorum) >= 0 {
		return utils.BYTE_FALSE, errors.NewErr("[SetOracleCronOutcome] Request have achieved quorum")
	}

	//save new outcomeRecord
	_, ok := cronOutcomeRecord.CronOutcomeRecord[params.Address]
	if ok {
		return utils.BYTE_FALSE, errors.NewErr("[SetOracleCronOutcome] Address has already setCronOutcome")
	}

	cronOutcomeRecord.CronOutcomeRecord[params.Address] = params.Outcome
	value, err := json.Marshal(cronOutcomeRecord)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Marshal] Marshal CronOutcomeRecord error")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(CRON_OUTCOME_RECORD), txHash, cronView.Bytes()), &cstates.StorageItem{Value: value})

	newNum := new(big.Int).SetInt64(int64(len(cronOutcomeRecord.CronOutcomeRecord)))

	//quorum achieved
	if newNum.Cmp(quorum) == 0 {
		//aggregate result
		consensus := true
		for _, v := range cronOutcomeRecord.CronOutcomeRecord {
			if params.Outcome != v {
				consensus = false
			}
		}
		if consensus {
			finalValue, err := json.Marshal(params.Outcome)
			if err != nil {
				return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Marshal] Marshal FinalCronOutcome error")
			}
			native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(FINAL_CRON_OUTCOME), txHash, cronView.Bytes()), &cstates.StorageItem{Value: finalValue})
		}

	}
	utils.AddCommonEvent(native, contract, SET_ORACLE_CRON_OUTCOME, params)

	return utils.BYTE_TRUE, nil
}

func ChangeCronView(native *native.NativeService) ([]byte, error) {
	params := new(ChangeCronViewParam)
	err := json.Unmarshal(native.Input, params)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Unmarshal] Contract params Unmarshal error!")
	}

	//check witness
	err = utils.ValidateOwner(native, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[ValidateOwner] validateOwner error!")
	}

	txHashHex := params.TxHash
	txHash, err := hex.DecodeString(txHashHex)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[hex.DecodeString] Decode hex txHash error!")
	}

	contract := native.ContextRef.CurrentContext().ContractAddress

	//check if is request owner
	request := new(CreateOracleRequestParam)
	requestBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(REQUEST), txHash))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[native.CloneCache.Get] Get Request error!")
	}
	if requestBytes == nil {
		return utils.BYTE_FALSE, errors.NewErr("[ChangeCronView] Request of this txHash is nil, check input txHash!")
	}
	item, _ := requestBytes.(*cstates.StorageItem)
	err = json.Unmarshal(item.Value, request)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[setOracleCronOutcome] Unmarshal CronOutcomeRecord error")
	}
	if request.Address != params.Address {
		return utils.BYTE_FALSE, errors.NewErr("[ChangeCronView] Only Request Owner can ChangeCronView!")
	}

	//CronView add 1
	cronViewBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(CRON_VIEW), txHash))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[native.CloneCache.Get] Get CronView error!")
	}
	var cronView *big.Int
	if cronViewBytes == nil {
		cronView = new(big.Int).SetInt64(1)
	} else {
		cronViewStore, _ := cronViewBytes.(*cstates.StorageItem)
		cronView = new(big.Int).SetBytes(cronViewStore.Value)
	}

	newCronView := new(big.Int).Add(cronView, new(big.Int).SetInt64(1))
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(CRON_VIEW), txHash), &cstates.StorageItem{Value: newCronView.Bytes()})

	utils.AddCommonEvent(native, contract, CHANGE_CRON_VIEW, newCronView)

	return utils.BYTE_TRUE, nil
}
