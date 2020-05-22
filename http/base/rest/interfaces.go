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
	"strconv"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	scom "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/types"
	ontErrors "github.com/ontio/ontology/errors"
	bactor "github.com/ontio/ontology/http/base/actor"
	bcomn "github.com/ontio/ontology/http/base/common"
	berr "github.com/ontio/ontology/http/base/error"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

const TLS_PORT int = 443

type ApiServer interface {
	Start() error
	Stop()
}

// get node verison
func GetNodeVersion(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	resp["Result"] = config.Version
	return resp
}

// get networkid
func GetNetworkId(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	resp["Result"] = config.DefConfig.P2PNode.NetworkId
	return resp
}

//get connection node count
func GetConnectionCount(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	count := bactor.GetConnectionCnt()
	resp["Result"] = count
	return resp
}

func GetNodeSyncStatus(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	status, err := bcomn.GetSyncStatus()
	if err != nil {
		return ResponsePack(berr.INTERNAL_ERROR)
	}
	resp["Result"] = status
	return resp
}

//get block height
func GetBlockHeight(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	height := bactor.GetCurrentBlockHeight()
	resp["Result"] = height
	return resp
}

//get block hash by height
func GetBlockHash(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	param, ok := cmd["Height"].(string)
	if !ok || len(param) == 0 {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	height, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	hash := bactor.GetBlockHashFromStore(uint32(height))
	if hash == common.UINT256_EMPTY {
		return ResponsePack(berr.UNKNOWN_BLOCK)
	}
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
		return common.ToHexString(block.ToArray()), berr.SUCCESS
	}
	return bcomn.GetBlockInfo(block), berr.SUCCESS
}

//get block by hash
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

//get block height by transaction hash
func GetBlockHeightByTxHash(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	str, ok := cmd["Hash"].(string)
	if !ok || len(str) == 0 {
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

//get block transaction hashes by height
func GetBlockTxsByHeight(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)

	param, ok := cmd["Height"].(string)
	if !ok || len(param) == 0 {
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

//get block by height
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
		resp["Result"] = common.ToHexString(block.ToArray())
	} else {
		resp["Result"] = bcomn.GetBlockInfo(block)
	}
	return resp
}

//get transaction by hash
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
	height, tx, err := bactor.GetTxnWithHeightByTxHash(hash)
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
		resp["Result"] = common.ToHexString(common.SerializeToBytes(tx))
		return resp
	}
	tran := bcomn.TransArryByteToHexString(tx)
	tran.Height = height
	resp["Result"] = tran
	return resp
}

//send raw transaction
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

	txn, err := types.TransactionFromRawBytes(bys)
	if err != nil {
		return ResponsePack(berr.INVALID_TRANSACTION)
	}
	var hash common.Uint256
	hash = txn.Hash()
	log.Debugf("SendRawTransaction recv %s", hash.ToHexString())
	if txn.TxType == types.InvokeNeo || txn.TxType == types.InvokeWasm || txn.TxType == types.Deploy {
		if preExec, ok := cmd["PreExec"].(string); ok && preExec == "1" {
			rst, err := bactor.PreExecuteContract(txn)
			if err != nil {
				log.Infof("PreExec: ", err)
				resp = ResponsePack(berr.SMARTCODE_ERROR)
				resp["Result"] = err.Error()
				return resp
			}
			resp["Result"] = bcomn.ConvertPreExecuteResult(rst)
			return resp
		}
	}
	log.Debugf("SendRawTransaction send to txpool %s", hash.ToHexString())
	if errCode, desc := bcomn.SendTxToPool(txn); errCode != ontErrors.ErrNoError {
		resp["Error"] = int64(errCode)
		resp["Result"] = desc
		log.Warnf("SendRawTransaction verified %s error: %s", hash.ToHexString(), desc)
		return resp
	}
	log.Debugf("SendRawTransaction verified %s", hash.ToHexString())
	resp["Result"] = hash.ToHexString()
	return resp
}

//get smartcontract event by height
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
		return ResponsePack(berr.INTERNAL_ERROR)
	}
	eInfos := make([]*bcomn.ExecuteNotify, 0, len(eventInfos))
	for _, eventInfo := range eventInfos {
		_, notify := bcomn.GetExecuteNotify(eventInfo)
		eInfos = append(eInfos, &notify)
	}
	resp["Result"] = eInfos
	return resp
}

//get smartcontract event by transaction hash
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
		if scom.ErrNotFound == err {
			return ResponsePack(berr.SUCCESS)
		}
		return ResponsePack(berr.INTERNAL_ERROR)
	}
	if eventInfo == nil {
		return ResponsePack(berr.INVALID_TRANSACTION)
	}
	_, notify := bcomn.GetExecuteNotify(eventInfo)
	resp["Result"] = notify
	return resp
}

//get contract state
func GetContractState(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	str, ok := cmd["Hash"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	address, err := bcomn.GetAddress(str)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	contract, err := bactor.GetContractStateFromStore(address)
	if err != nil {
		return ResponsePack(berr.INTERNAL_ERROR)
	}
	if contract == nil {
		return ResponsePack(berr.UNKNOWN_CONTRACT)
	}
	if raw, ok := cmd["Raw"].(string); ok && raw == "1" {
		sink := common.NewZeroCopySink(nil)
		contract.Serialization(sink)
		resp["Result"] = common.ToHexString(sink.Bytes())
		return resp
	}
	resp["Result"] = bcomn.TransPayloadToHex(contract)
	return resp
}

//get storage from contract
func GetStorage(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	str, ok := cmd["Hash"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	address, err := bcomn.GetAddress(str)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	key := cmd["Key"].(string)
	item, err := common.HexToBytes(key)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	value, err := bactor.GetStorageItem(address, item)
	if err != nil {
		if err == scom.ErrNotFound {
			return ResponsePack(berr.SUCCESS)
		}
		return ResponsePack(berr.INTERNAL_ERROR)
	}
	resp["Result"] = common.ToHexString(value)
	return resp
}

//get balance of address
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

//get merkle proof by transaction hash
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

//get avg gas price in block
func GetGasPrice(cmd map[string]interface{}) map[string]interface{} {
	result, err := bcomn.GetGasPrice()
	if err != nil {
		return ResponsePack(berr.INTERNAL_ERROR)
	}
	resp := ResponsePack(berr.SUCCESS)
	resp["Result"] = result
	return resp
}

//get allowance
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
	fromAddr, err := bcomn.GetAddress(fromAddrStr)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	toAddr, err := bcomn.GetAddress(toAddrStr)
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

//get unbound ong
func GetUnboundOng(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	toAddrStr, ok := cmd["Addr"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	toAddr, err := bcomn.GetAddress(toAddrStr)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	fromAddr := utils.OntContractAddress
	rsp, err := bcomn.GetAllowance("ong", fromAddr, toAddr)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	bcomn.GetGrantOng(toAddr)
	resp["Result"] = rsp
	return resp
}

//get grant ong
func GetGrantOng(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	toAddrStr, ok := cmd["Addr"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	toAddr, err := bcomn.GetAddress(toAddrStr)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	rsp, err := bcomn.GetGrantOng(toAddr)
	if err != nil {
		return ResponsePack(berr.INTERNAL_ERROR)
	}
	resp["Result"] = rsp
	return resp
}

//get memory pool transaction count
func GetMemPoolTxCount(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	count, err := bactor.GetTxnCount()
	if err != nil {
		return ResponsePack(berr.INTERNAL_ERROR)
	}
	resp["Result"] = count
	return resp
}

//get memory pool transaction hash list
func GetMemPoolTxHashList(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	txHashList, err := bactor.GetTxnHashList()
	if err != nil {
		return ResponsePack(berr.INTERNAL_ERROR)
	}
	resp["Result"] = txHashList
	return resp
}

//get memory poll transaction state
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
