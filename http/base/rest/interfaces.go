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

package rest

import (
	"bytes"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/payload"
	scom "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/types"
	ontErrors "github.com/ontio/ontology/errors"
	bactor "github.com/ontio/ontology/http/base/actor"
	bcomn "github.com/ontio/ontology/http/base/common"
	berr "github.com/ontio/ontology/http/base/error"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"strconv"
)

const TLS_PORT int = 443

type ApiServer interface {
	Start() error
	Stop()
}

//Node
func GetGenerateBlockTime(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	resp["Result"] = config.DEFAULT_GEN_BLOCK_TIME
	return resp
}
func GetConnectionCount(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	count, err := bactor.GetConnectionCnt()
	if err != nil {
		return ResponsePack(berr.INTERNAL_ERROR)
	}
	resp["Result"] = count
	return resp
}

//Block
func GetBlockHeight(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	height := bactor.GetCurrentBlockHeight()
	resp["Result"] = height
	return resp
}
func GetBlockHash(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	param := cmd["Height"].(string)
	if len(param) == 0 {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	height, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	hash := bactor.GetBlockHashFromStore(uint32(height))
	resp["Result"] = hash.ToHexString()
	return resp
}

func getBlock(hash common.Uint256, getTxBytes bool) (interface{}, int64) {
	block, err := bactor.GetBlockFromStore(hash)
	if err != nil {
		return nil, berr.UNKNOWN_BLOCK
	}
	if block == nil {
		return nil, berr.UNKNOWN_BLOCK
	}
	if block.Header == nil {
		return nil, berr.UNKNOWN_BLOCK
	}
	if getTxBytes {
		w := bytes.NewBuffer(nil)
		block.Serialize(w)
		return common.ToHexString(w.Bytes()), berr.SUCCESS
	}
	return bcomn.GetBlockInfo(block), berr.SUCCESS
}
func GetBlockByHash(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	str := cmd["Hash"].(string)
	if len(str) == 0 {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	var getTxBytes = false
	if raw, ok := cmd["Raw"].(string); ok && raw == "1" {
		getTxBytes = true
	}
	var hash common.Uint256
	hash, err := common.Uint256FromHexString(str)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	resp["Result"], resp["Error"] = getBlock(hash, getTxBytes)
	return resp
}

func GetBlockHeightByTxHash(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	str := cmd["Hash"].(string)
	if len(str) == 0 {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	hash, err := common.Uint256FromHexString(str)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	height, tx, err := bactor.GetTxnWithHeightByTxHash(hash)
	if err != nil {
		return ResponsePack(berr.INTERNAL_ERROR)
	}
	if tx == nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	resp["Result"] = height
	return resp
}
func GetBlockTxsByHeight(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)

	param := cmd["Height"].(string)
	if len(param) == 0 {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	height, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	index := uint32(height)
	hash := bactor.GetBlockHashFromStore(index)
	if hash == common.UINT256_EMPTY {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	block, err := bactor.GetBlockFromStore(hash)
	if err != nil {
		return ResponsePack(berr.UNKNOWN_BLOCK)
	}
	resp["Result"] = bcomn.GetBlockTransactions(block)
	return resp
}
func GetBlockByHeight(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)

	param, ok := cmd["Height"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	if len(param) == 0 {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	var getTxBytes = false
	if raw, ok := cmd["Raw"].(string); ok && raw == "1" {
		getTxBytes = true
	}
	height, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	index := uint32(height)
	block, err := bactor.GetBlockByHeight(index)
	if err != nil || block == nil {
		return ResponsePack(berr.UNKNOWN_BLOCK)
	}
	if getTxBytes {
		w := bytes.NewBuffer(nil)
		block.Serialize(w)
		resp["Result"] = common.ToHexString(w.Bytes())
	} else {
		resp["Result"] = bcomn.GetBlockInfo(block)
	}
	return resp
}

//Transaction
func GetTransactionByHash(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)

	str, ok := cmd["Hash"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	hash, err := common.Uint256FromHexString(str)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	tx, err := bactor.GetTransaction(hash)
	if tx == nil {
		return ResponsePack(berr.UNKNOWN_TRANSACTION)
	}
	if err != nil {
		return ResponsePack(berr.UNKNOWN_TRANSACTION)
	}
	if tx == nil {
		return ResponsePack(berr.UNKNOWN_TRANSACTION)
	}
	if raw, ok := cmd["Raw"].(string); ok && raw == "1" {
		w := bytes.NewBuffer(nil)
		tx.Serialize(w)
		resp["Result"] = common.ToHexString(w.Bytes())
		return resp
	}
	tran := bcomn.TransArryByteToHexString(tx)
	resp["Result"] = tran
	return resp
}
func SendRawTransaction(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)

	str, ok := cmd["Data"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	bys, err := common.HexToBytes(str)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	var txn types.Transaction
	if err := txn.Deserialize(bytes.NewReader(bys)); err != nil {
		return ResponsePack(berr.INVALID_TRANSACTION)
	}
	if txn.TxType == types.Invoke {
		if preExec, ok := cmd["PreExec"].(string); ok && preExec == "1" {
			if _, ok := txn.Payload.(*payload.InvokeCode); ok {
				resp["Result"], err = bactor.PreExecuteContract(&txn)
				if err != nil {
					log.Error(err)
					return ResponsePack(berr.SMARTCODE_ERROR)
				}
				return resp
			}
		}
	}
	var hash common.Uint256
	hash = txn.Hash()
	if errCode := bcomn.VerifyAndSendTx(&txn); errCode != ontErrors.ErrNoError {
		resp["Error"] = int64(errCode)
		return resp
	}
	resp["Result"] = hash.ToHexString()

	return resp
}

func GetSmartCodeEventTxsByHeight(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)

	param, ok := cmd["Height"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	if len(param) == 0 {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	height, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	index := uint32(height)
	eventInfos, err := bactor.GetEventNotifyByHeight(index)
	if err != nil {
		if scom.ErrNotFound == err {
			return ResponsePack(berr.SUCCESS)
		}
		return ResponsePack(berr.INVALID_PARAMS)
	}
	eInfos := make([]*bcomn.ExecuteNotify, 0, len(eventInfos))
	for _, eventInfo := range eventInfos {
		_, notify := bcomn.GetExecuteNotify(eventInfo)
		eInfos = append(eInfos, &notify)
	}
	resp["Result"] = eInfos
	return resp
}

func GetSmartCodeEventByTxHash(cmd map[string]interface{}) map[string]interface{} {
	if !config.DefConfig.Common.EnableEventLog {
		return ResponsePack(berr.INVALID_METHOD)
	}

	resp := ResponsePack(berr.SUCCESS)

	str, ok := cmd["Hash"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	hash, err := common.Uint256FromHexString(str)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	eventInfo, err := bactor.GetEventNotifyByTxHash(hash)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	if eventInfo == nil {
		return ResponsePack(berr.INVALID_TRANSACTION)
	}
	_, notify := bcomn.GetExecuteNotify(eventInfo)
	resp["Result"] = notify
	return resp
}

func GetContractState(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	str, ok := cmd["Hash"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	var codeHash common.Address
	var err error
	if len(str) == common.ADDR_LEN*2 {
		codeHash, err = common.AddressFromHexString(str)
	} else {
		codeHash, err = common.AddressFromBase58(str)
	}
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	contract, err := bactor.GetContractStateFromStore(codeHash)
	if err != nil {
		return ResponsePack(berr.INTERNAL_ERROR)
	}
	if contract == nil {
		return ResponsePack(berr.UNKNOWN_CONTRACT)
	}
	if raw, ok := cmd["Raw"].(string); ok && raw == "1" {
		w := bytes.NewBuffer(nil)
		contract.Serialize(w)
		resp["Result"] = common.ToHexString(w.Bytes())
		return resp
	}
	resp["Result"] = bcomn.TransPayloadToHex(contract)
	return resp
}

func GetStorage(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	str, ok := cmd["Hash"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	var codeHash common.Address
	var err error
	if len(str) == common.ADDR_LEN*2 {
		codeHash, err = common.AddressFromHexString(str)
	} else {
		codeHash, err = common.AddressFromBase58(str)
	}
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	key := cmd["Key"].(string)
	item, err := common.HexToBytes(key)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	value, err := bactor.GetStorageItem(codeHash, item)
	if err != nil || value == nil {
		return ResponsePack(berr.INTERNAL_ERROR)
	}
	resp["Result"] = common.ToHexString(value)
	return resp
}

func GetBalance(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	addrBase58, ok := cmd["Addr"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	address, err := common.AddressFromBase58(addrBase58)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	balance, err := bcomn.GetBalance(address)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	resp["Result"] = balance
	return resp
}

func GetMerkleProof(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	str, ok := cmd["Hash"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	hash, err := common.Uint256FromHexString(str)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	height, tx, err := bactor.GetTxnWithHeightByTxHash(hash)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	if tx == nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	header, err := bactor.GetHeaderByHeight(height)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	curHeight := bactor.GetCurrentBlockHeight()
	curHeader, err := bactor.GetHeaderByHeight(curHeight)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	proof, err := bactor.GetMerkleProof(uint32(height), uint32(curHeight))
	if err != nil {
		return ResponsePack(berr.INTERNAL_ERROR)
	}
	var hashes []string
	for _, v := range proof {
		hashes = append(hashes, v.ToHexString())
	}
	resp["Result"] = bcomn.MerkleProof{"MerkleProof", header.TransactionsRoot.ToHexString(), height,
		curHeader.BlockRoot.ToHexString(), curHeight, hashes}
	return resp
}

func GetGasPrice(cmd map[string]interface{}) map[string]interface{} {
	result, err := bcomn.GetGasPrice()
	if err != nil {
		return ResponsePack(berr.INTERNAL_ERROR)
	}
	resp := ResponsePack(berr.SUCCESS)
	resp["Result"] = result
	return resp
}

func GetAllowance(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	asset, ok := cmd["Asset"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	fromAddrStr, ok := cmd["From"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	toAddrStr, ok := cmd["To"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	fromAddr, err := common.AddressFromBase58(fromAddrStr)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	toAddr, err := common.AddressFromBase58(toAddrStr)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	rsp, err := bcomn.GetAllowance(asset, fromAddr, toAddr)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	resp["Result"] = rsp
	return resp
}

func GetUnclaimOng(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	toAddrStr, ok := cmd["Addr"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	toAddr, err := common.AddressFromBase58(toAddrStr)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	fromAddr := utils.OntContractAddress
	rsp, err := bcomn.GetAllowance("ong", fromAddr, toAddr)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	resp["Result"] = rsp
	return resp
}

func GetMemPoolTxCount(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	count, err := bactor.GetTxnCount()
	if err != nil {
		return ResponsePack(berr.INTERNAL_ERROR)
	}
	resp["Result"] = count
	return resp
}
func GetMemPoolTxState(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	str, ok := cmd["Hash"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	hash, err := common.Uint256FromHexString(str)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	txEntry, err := bactor.GetTxFromPool(hash)
	if err != nil {
		return ResponsePack(berr.UNKNOWN_TRANSACTION)
	}
	attrs := []bcomn.TXNAttrInfo{}
	for _, t := range txEntry.Attrs {
		attrs = append(attrs, bcomn.TXNAttrInfo{t.Height, int(t.Type), int(t.ErrCode)})
	}
	resp["Result"] = bcomn.TXNEntryInfo{attrs}
	return resp
}
