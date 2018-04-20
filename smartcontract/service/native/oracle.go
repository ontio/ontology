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

package native

import (
	"encoding/hex"
	"encoding/json"
	"math/big"

	cstates "github.com/ontio/ontology/core/states"
	scommon "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/core/genesis"
	"github.com/ontio/ontology/smartcontract/service/native/states"
)

const (
	//function name
	CREATE_ORACLE_REQUEST = "createOracleRequest"
	SET_ORACLE_OUTCOME = "setOracleOutcome"
	SET_ORACLE_CRON_OUTCOME = "setOracleCronOutcome"
	CHANGE_CRON_VIEW = "changeCronView"

	//keyPrefix
	UNDO_TXHASH         = "UndoTxHash"
	ORACLE_NUM          = "OracleNum"
	REQUEST             = "Request"
	OUTCOME_RECORD      = "OutcomeRecord"
	FINAL_OUTCOME       = "FinalOutcome"
	CRON_VIEW           = "CronView"
	CRON_OUTCOME_RECORD = "CronOutcomeRecord"
	FINAL_CRON_OUTCOME  = "FinalCronOutcome"
)

func init() {
	Contracts[genesis.OracleContractAddress] = RegisterOracleContract
}

func RegisterOracleContract(native *NativeService) {
	native.Register(CREATE_ORACLE_REQUEST, CreateOracleRequest)
	native.Register(SET_ORACLE_OUTCOME, SetOracleOutcome)
	native.Register(SET_ORACLE_CRON_OUTCOME, SetOracleCronOutcome)
	native.Register(CHANGE_CRON_VIEW, ChangeCronView)
}

func CreateOracleRequest(native *NativeService) error {
	params := new(states.CreateOracleRequestParam)
	err := json.Unmarshal(native.Input, &params)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[createOracleRequest] Contract params Unmarshal error!")
	}

	//check witness
	err = validateOwner(native, params.Address)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[createOracleRequest] validateOwner error!")
	}

	if params.OracleNum.Cmp(new(big.Int).SetInt64(0)) == 0 {
		params.OracleNum.SetInt64(1)
	}

	contract := native.ContextRef.CurrentContext().ContractAddress
	txHash := native.Tx.Hash()
	txHashBytes := txHash.ToArray()
	txHashHex := hex.EncodeToString(txHashBytes)
	undoRequests := &states.UndoRequests{
		Requests: make(map[string]interface{}),
	}

	undoRequestsBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, append(contract[:], []byte(UNDO_TXHASH)...))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[createOracleRequest] Get UndoRequests error!")
	}

	if undoRequestsBytes != nil {
		item, _ := undoRequestsBytes.(*cstates.StorageItem)
		err = json.Unmarshal(item.Value, &undoRequests)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[createOracleRequest] Unmarshal UndoRequests error")
		}
	}

	undoRequests.Requests[txHashHex] = new(interface{})

	value, err := json.Marshal(undoRequests)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[createOracleRequest] Marshal UndoRequests error")
	}

	native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(ORACLE_NUM), txHashBytes), &cstates.StorageItem{Value: params.OracleNum.Bytes()})
	native.CloneCache.Add(scommon.ST_STORAGE, append(contract[:], []byte(UNDO_TXHASH)...), &cstates.StorageItem{Value: value})
	native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(REQUEST), txHashBytes), &cstates.StorageItem{Value: native.Input})

	createOracleRequestEvent(native, contract, params.Request)

	return nil
}

func SetOracleOutcome(native *NativeService) error {
	params := new(states.SetOracleOutcomeParam)
	err := json.Unmarshal(native.Input, &params)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[setOracleOutcome] Contract params Unmarshal error!")
	}

	//check witness
	err = validateOwner(native, params.Owner)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[setOracleOutcome] validateOwner error!")
	}

	txHashHex := params.TxHash
	txHash, err := hex.DecodeString(txHashHex)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[setOracleOutcome] Decode hex txHash error!")
	}

	contract := native.ContextRef.CurrentContext().ContractAddress
	outcomeRecord := &states.OutcomeRecord{
		OutcomeRecord: make(map[string]interface{}),
	}

	outcomeRecordBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, concatKey(contract, []byte(OUTCOME_RECORD), txHash))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[setOracleOutcome] Get OutcomeRecord error!")
	}

	if outcomeRecordBytes != nil {
		item, _ := outcomeRecordBytes.(*cstates.StorageItem)
		err = json.Unmarshal(item.Value, &outcomeRecord)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[setOracleOutcome] Unmarshal OutcomeRecord error")
		}
	}

	num := new(big.Int).SetInt64(int64(len(outcomeRecord.OutcomeRecord)))
	oracleNum, err := native.CloneCache.Get(scommon.ST_STORAGE, concatKey(contract, []byte(ORACLE_NUM), txHash))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[setOracleOutcome] Get OracleNum error!")
	}
	if oracleNum == nil {
		return errors.NewErr("[setOracleOutcome] Get nil OracleNum, check input txHash!")
	}
	item, _ := oracleNum.(*cstates.StorageItem)
	quorum := new(big.Int).SetBytes(item.Value)

	//quorum achieved
	if num.Cmp(quorum) >= 0 {
		return errors.NewErr("[setOracleOutcome] Request have achieved quorum")
	}

	//save new outcomeRecord
	_, ok := outcomeRecord.OutcomeRecord[params.Owner]
	if ok {
		return errors.NewErr("[setOracleOutcome] Address has already setOutcome")
	}

	outcomeRecord.OutcomeRecord[params.Owner] = params.Outcome
	value, err := json.Marshal(outcomeRecord)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[setOracleOutcome] Marshal OutcomeRecord error")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(OUTCOME_RECORD), txHash), &cstates.StorageItem{Value: value})

	newNum := new(big.Int).SetInt64(int64(len(outcomeRecord.OutcomeRecord)))

	//quorum achieved
	if newNum.Cmp(quorum) == 0 {
		//remove txHash from undoRequests
		undoRequests := &states.UndoRequests{
			Requests: make(map[string]interface{}),
		}

		undoRequestsBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, append(contract[:], []byte(UNDO_TXHASH)...))
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[setOracleOutcome] Get UndoRequests error!")
		}

		if undoRequestsBytes != nil {
			item, _ := undoRequestsBytes.(*cstates.StorageItem)
			err = json.Unmarshal(item.Value, &undoRequests)
			if err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "[setOracleOutcome] Unmarshal UndoRequests error")
			}
		}
		delete(undoRequests.Requests, params.TxHash)
		value, err := json.Marshal(undoRequests)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[setOracleOutcome] Marshal UndoRequests error")
		}
		native.CloneCache.Add(scommon.ST_STORAGE, append(contract[:], []byte(UNDO_TXHASH)...), &cstates.StorageItem{Value: value})

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
				return errors.NewDetailErr(err, errors.ErrNoCode, "[setOracleOutcome] Marshal FinalOutcome error")
			}
			native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(FINAL_OUTCOME), txHash), &cstates.StorageItem{Value: finalValue})
		}

	}
	setOracleOutcomeEvent(native, contract, true)

	return nil
}

func SetOracleCronOutcome(native *NativeService) error {
	params := new(states.SetOracleCronOutcomeParam)
	err := json.Unmarshal(native.Input, &params)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[setOracleCronOutcome] Contract params Unmarshal error!")
	}

	//check witness
	err = validateOwner(native, params.Owner)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[setOracleCronOutcome] validateOwner error!")
	}

	txHashHex := params.TxHash
	txHash, err := hex.DecodeString(txHashHex)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[setOracleCronOutcome] Decode hex txHash error!")
	}

	contract := native.ContextRef.CurrentContext().ContractAddress
	cronOutcomeRecord := &states.CronOutcomeRecord{
		CronOutcomeRecord: make(map[string]interface{}),
	}

	cronViewBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, concatKey(contract, []byte(CRON_VIEW), txHash))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[setOracleCronOutcome] Get CronView error!")
	}
	var cronView *big.Int
	if cronViewBytes == nil {
		cronView = new(big.Int).SetInt64(1)
	} else {
		cronViewStore, _ := cronViewBytes.(*cstates.StorageItem)
		cronView = new(big.Int).SetBytes(cronViewStore.Value)
	}

	outcomeRecordBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, concatKey(contract, []byte(CRON_OUTCOME_RECORD), txHash, cronView.Bytes()))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[setOracleCronOutcome] Get CronOutcomeRecord error!")
	}

	if outcomeRecordBytes != nil {
		item, _ := outcomeRecordBytes.(*cstates.StorageItem)
		err = json.Unmarshal(item.Value, &cronOutcomeRecord)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[setOracleCronOutcome] Unmarshal CronOutcomeRecord error")
		}
	}

	num := new(big.Int).SetInt64(int64(len(cronOutcomeRecord.CronOutcomeRecord)))
	oracleNum, err := native.CloneCache.Get(scommon.ST_STORAGE, concatKey(contract, []byte(ORACLE_NUM), txHash))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[setOracleCronOutcome] Get OracleNum error!")
	}
	if oracleNum == nil {
		return errors.NewErr("[setOracleCronOutcome] Get nil OracleNum, check input txHash!")
	}
	item, _ := oracleNum.(*cstates.StorageItem)
	quorum := new(big.Int).SetBytes(item.Value)

	//quorum achieved
	if num.Cmp(quorum) >= 0 {
		return errors.NewErr("[setOracleCronOutcome] Request have achieved quorum")
	}

	//save new outcomeRecord
	_, ok := cronOutcomeRecord.CronOutcomeRecord[params.Owner]
	if ok {
		return errors.NewErr("[setOracleCronOutcome] Address has already setCronOutcome")
	}

	cronOutcomeRecord.CronOutcomeRecord[params.Owner] = params.Outcome
	value, err := json.Marshal(cronOutcomeRecord)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[setOracleCronOutcome] Marshal CronOutcomeRecord error")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(CRON_OUTCOME_RECORD), txHash, cronView.Bytes()), &cstates.StorageItem{Value: value})

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
				return errors.NewDetailErr(err, errors.ErrNoCode, "[setOracleCronOutcome] Marshal FinalCronOutcome error")
			}
			native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(FINAL_CRON_OUTCOME), txHash, cronView.Bytes()), &cstates.StorageItem{Value: finalValue})
		}

	}
	setOracleCronOutcomeEvent(native, contract, true)

	return nil
}

func ChangeCronView(native *NativeService) error {
	params := new(states.ChangeCronViewParam)
	err := json.Unmarshal(native.Input, &params)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[changeCronView] Contract params Unmarshal error!")
	}

	//check witness
	err = validateOwner(native, params.Owner)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[changeCronView] validateOwner error!")
	}

	txHashHex := params.TxHash
	txHash, err := hex.DecodeString(txHashHex)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[changeCronView] Decode hex txHash error!")
	}

	contract := native.ContextRef.CurrentContext().ContractAddress

	//check if is request owner
	request := new(states.CreateOracleRequestParam)
	requestBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, concatKey(contract, []byte(REQUEST), txHash))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[changeCronView] Get Request error!")
	}
	if requestBytes == nil {
		return errors.NewErr("[changeCronView] Request of this txHash is nil, check input txHash!")
	}
	item, _ := requestBytes.(*cstates.StorageItem)
	err = json.Unmarshal(item.Value, &request)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[setOracleCronOutcome] Unmarshal CronOutcomeRecord error")
	}
	if request.Address != params.Owner {
		return errors.NewErr("[changeCronView] Only Request Owner can ChangeCronView!")
	}

	//CronView add 1
	cronViewBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, concatKey(contract, []byte(CRON_VIEW), txHash))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[changeCronView] Get CronView error!")
	}
	var cronView *big.Int
	if cronViewBytes == nil {
		cronView = new(big.Int).SetInt64(1)
	} else {
		cronViewStore, _ := cronViewBytes.(*cstates.StorageItem)
		cronView = new(big.Int).SetBytes(cronViewStore.Value)
	}

	newCronView := new(big.Int).Add(cronView, new(big.Int).SetInt64(1))
	native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(CRON_VIEW), txHash), &cstates.StorageItem{Value: newCronView.Bytes()})

	changeCronViewEvent(native, contract, newCronView)

	return nil
}
