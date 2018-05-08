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

package rpc

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/types"
	ontErrors "github.com/ontio/ontology/errors"
	bactor "github.com/ontio/ontology/http/base/actor"
	bcomn "github.com/ontio/ontology/http/base/common"
	berr "github.com/ontio/ontology/http/base/error"
)

func GetGenerateBlockTime(params []interface{}) map[string]interface{} {
	return responseSuccess(config.DEFAULT_GEN_BLOCK_TIME)
}

func GetBestBlockHash(params []interface{}) map[string]interface{} {
	hash := bactor.CurrentBlockHash()
	return responseSuccess(common.ToHexString(hash.ToArray()))
}

// Input JSON string examples for getblock method as following:
//   {"jsonrpc": "2.0", "method": "getblock", "params": [1], "id": 0}
//   {"jsonrpc": "2.0", "method": "getblock", "params": ["aabbcc.."], "id": 0}
func GetBlock(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return responsePack(berr.INVALID_PARAMS, nil)
	}
	var err error
	var hash common.Uint256
	switch (params[0]).(type) {
	// block height
	case float64:
		index := uint32(params[0].(float64))
		hash = bactor.GetBlockHashFromStore(index)
		if err != nil {
			log.Errorf("GetBlock GetBlockHashFromStore Height:%v error:%s", index, err)
			return responsePack(berr.UNKNOWN_BLOCK, "unknown block")
		}
		if hash == common.UINT256_EMPTY {
			return responsePack(berr.INVALID_PARAMS, "")
		}
		// block hash
	case string:
		str := params[0].(string)
		hex, err := hex.DecodeString(str)
		if err != nil {
			return responsePack(berr.INVALID_PARAMS, "")
		}
		if err := hash.Deserialize(bytes.NewReader(hex)); err != nil {
			return responsePack(berr.INVALID_PARAMS, "")
		}
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}
	block, err := bactor.GetBlockFromStore(hash)
	if err != nil {
		log.Errorf("GetBlock GetBlockFromStore BlockHash:%x error:%s", hash, err)
		return responsePack(berr.UNKNOWN_BLOCK, "unknown block")
	}
	if block == nil || block.Header == nil {
		return responsePack(berr.UNKNOWN_BLOCK, "unknown block")
	}
	if len(params) >= 2 {
		switch (params[1]).(type) {
		case float64:
			json := uint32(params[1].(float64))
			if json == 1 {
				return responseSuccess(bcomn.GetBlockInfo(block))
			}
		default:
			return responsePack(berr.INVALID_PARAMS, "")
		}
	}
	w := bytes.NewBuffer(nil)
	block.Serialize(w)
	return responseSuccess(common.ToHexString(w.Bytes()))
}

func GetBlockCount(params []interface{}) map[string]interface{} {
	height := bactor.GetCurrentBlockHeight()
	return responseSuccess(height + 1)
}

// A JSON example for getblockhash method as following:
//   {"jsonrpc": "2.0", "method": "getblockhash", "params": [1], "id": 0}
func GetBlockHash(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return responsePack(berr.INVALID_PARAMS, nil)
	}
	switch params[0].(type) {
	case float64:
		height := uint32(params[0].(float64))
		hash := bactor.GetBlockHashFromStore(height)
		return responseSuccess(fmt.Sprintf("%016x", hash))
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}
}

func GetConnectionCount(params []interface{}) map[string]interface{} {
	count, err := bactor.GetConnectionCnt()
	if err != nil {
		log.Errorf("GetConnectionCount error:%s", err)
		return responsePack(berr.INTERNAL_ERROR, false)
	}
	return responseSuccess(count)
}

func GetRawMemPool(params []interface{}) map[string]interface{} {
	txs := []*bcomn.Transactions{}
	txpool := bactor.GetTxsFromPool(false)
	for _, t := range txpool {
		txs = append(txs, bcomn.TransArryByteToHexString(t))
	}
	if len(txs) == 0 {
		return responsePack(berr.INVALID_PARAMS, nil)
	}
	return responseSuccess(txs)
}
func GetMemPoolTxState(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return responsePack(berr.INVALID_PARAMS, nil)
	}
	switch params[0].(type) {
	case string:
		str := params[0].(string)
		hex, err := hex.DecodeString(str)
		if err != nil {
			return responsePack(berr.INVALID_PARAMS, "")
		}
		var hash common.Uint256
		err = hash.Deserialize(bytes.NewReader(hex))
		if err != nil {
			return responsePack(berr.INVALID_TRANSACTION, "")
		}
		txEntry, err := bactor.GetTxFromPool(hash)
		if err != nil {
			return responsePack(berr.UNKNOWN_TRANSACTION, "unknown transaction")
		}
		tran := bcomn.TransArryByteToHexString(txEntry.Tx)
		attrs := []bcomn.TXNAttrInfo{}
		for _, t := range txEntry.Attrs {
			attrs = append(attrs, bcomn.TXNAttrInfo{t.Height, int(t.Type), int(t.ErrCode)})
		}
		info := bcomn.TXNEntryInfo{*tran, int64(txEntry.Tx.GasPrice), attrs}
		return responseSuccess(info)
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}
}

// A JSON example for getrawtransaction method as following:
//   {"jsonrpc": "2.0", "method": "getrawtransaction", "params": ["transactioin hash in hex"], "id": 0}
func GetRawTransaction(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return responsePack(berr.INVALID_PARAMS, nil)
	}
	var tx *types.Transaction
	switch params[0].(type) {
	case string:
		str := params[0].(string)
		hex, err := hex.DecodeString(str)
		if err != nil {
			return responsePack(berr.INVALID_PARAMS, "")
		}
		var hash common.Uint256
		err = hash.Deserialize(bytes.NewReader(hex))
		if err != nil {
			return responsePack(berr.INVALID_TRANSACTION, "")
		}
		t, err := bactor.GetTransaction(hash)
		if t == nil {
			return responsePack(berr.UNKNOWN_TRANSACTION, "unknown transaction")
		}
		if err != nil {
			log.Errorf("GetRawTransaction GetTransaction error:%s", err)
			return responsePack(berr.UNKNOWN_TRANSACTION, "unknown transaction")
		}
		tx = t
		if tx == nil {
			return responsePack(berr.UNKNOWN_TRANSACTION, "unknown transaction")
		}
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}

	if len(params) >= 2 {
		switch (params[1]).(type) {
		case float64:
			json := uint32(params[1].(float64))
			if json == 1 {
				return responseSuccess(bcomn.TransArryByteToHexString(tx))
			}
		default:
			return responsePack(berr.INVALID_PARAMS, "")
		}
	}
	w := bytes.NewBuffer(nil)
	tx.Serialize(w)
	return responseSuccess(common.ToHexString(w.Bytes()))
}

//   {"jsonrpc": "2.0", "method": "getstorage", "params": ["code hash", "key"], "id": 0}
func GetStorage(params []interface{}) map[string]interface{} {
	if len(params) < 2 {
		return responsePack(berr.INVALID_PARAMS, nil)
	}

	var codeHash common.Address
	var key []byte
	switch params[0].(type) {
	case string:
		str := params[0].(string)
		hex, err := hex.DecodeString(str)
		if err != nil {
			return responsePack(berr.INVALID_PARAMS, "")
		}
		if err := codeHash.Deserialize(bytes.NewReader(hex)); err != nil {
			return responsePack(berr.INVALID_PARAMS, "")
		}
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}

	switch params[1].(type) {
	case string:
		str := params[1].(string)
		hex, err := hex.DecodeString(str)
		if err != nil {
			return responsePack(berr.INVALID_PARAMS, "")
		}
		key = hex
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}
	value, err := bactor.GetStorageItem(codeHash, key)
	if err != nil {
		log.Errorf("GetStorage GetStorageItem CodeHash:%x key:%s error:%s", codeHash, key, err)
		return responsePack(berr.INTERNAL_ERROR, "internal error")
	}
	return responseSuccess(common.ToHexString(value))
}

// A JSON example for sendrawtransaction method as following:
//   {"jsonrpc": "2.0", "method": "sendrawtransaction", "params": ["raw transactioin in hex"], "id": 0}
func SendRawTransaction(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return responsePack(berr.INVALID_PARAMS, nil)
	}
	var hash common.Uint256
	switch params[0].(type) {
	case string:
		str := params[0].(string)
		hex, err := hex.DecodeString(str)
		if err != nil {
			return responsePack(berr.INVALID_PARAMS, "")
		}
		var txn types.Transaction
		if err := txn.Deserialize(bytes.NewReader(hex)); err != nil {
			return responsePack(berr.INVALID_TRANSACTION, "")
		}
		if txn.TxType == types.Invoke && len(params) > 1 {
			preExec, ok := params[1].(float64)
			if ok && preExec == 1 {
				if _, ok := txn.Payload.(*payload.InvokeCode); ok {
					result, err := bactor.PreExecuteContract(&txn)
					if err != nil {
						log.Error(err)
						return responsePack(berr.SMARTCODE_ERROR, "")
					}
					return responseSuccess(result)
				}
			}
		}
		hash = txn.Hash()
		if errCode := bcomn.VerifyAndSendTx(&txn); errCode != ontErrors.ErrNoError {
			return responseSuccess(errCode.Error())
		}
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}
	return responseSuccess(common.ToHexString(hash.ToArray()))
}

func GetNodeVersion(params []interface{}) map[string]interface{} {
	return responseSuccess(config.Version)
}

func GetSystemFee(params []interface{}) map[string]interface{} {
	return responseSuccess(config.DefConfig.Common.SystemFee)
}

func GetContractState(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return responsePack(berr.INVALID_PARAMS, nil)
	}
	var contract *payload.DeployCode
	switch params[0].(type) {
	case string:
		str := params[0].(string)
		hex, err := hex.DecodeString(str)
		if err != nil {
			return responsePack(berr.INVALID_PARAMS, "")
		}
		var hash common.Address
		if err := hash.Deserialize(bytes.NewReader(hex)); err != nil {
			return responsePack(berr.INVALID_PARAMS, "")
		}
		c, err := bactor.GetContractStateFromStore(hash)
		if err != nil {
			log.Errorf("GetContractState GetContractStateFromStore hash:%x error:%s", hash, err)
			return responsePack(berr.INTERNAL_ERROR, "internal error")
		}
		if c == nil {
			return responsePack(berr.UNKNWN_CONTRACT, "unknow contract")
		}
		contract = c
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}
	if len(params) >= 2 {
		switch (params[1]).(type) {
		case float64:
			json := uint32(params[1].(float64))
			if json == 1 {
				return responseSuccess(bcomn.TransPayloadToHex(contract))
			}
		default:
			return responsePack(berr.INVALID_PARAMS, "")
		}
	}
	w := bytes.NewBuffer(nil)
	contract.Serialize(w)
	return responseSuccess(common.ToHexString(w.Bytes()))
}

func GetSmartCodeEvent(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return responsePack(berr.INVALID_PARAMS, nil)
	}

	switch (params[0]).(type) {
	// block height
	case float64:
		height := uint32(params[0].(float64))
		txs, err := bactor.GetEventNotifyByHeight(height)
		if err != nil {
			return responsePack(berr.INVALID_PARAMS, "")
		}
		var txhexs []string
		for _, v := range txs {
			txhexs = append(txhexs, common.ToHexString(v.ToArray()))
		}
		return responseSuccess(txhexs)
		//txhash
	case string:
		str := params[0].(string)
		hex, err := hex.DecodeString(str)
		if err != nil {
			return responsePack(berr.INVALID_PARAMS, "")
		}
		var hash common.Uint256
		if err := hash.Deserialize(bytes.NewReader(hex)); err != nil {
			return responsePack(berr.INVALID_PARAMS, "")
		}
		eventInfo, err := bactor.GetEventNotifyByTxHash(hash)
		if err != nil {
			return responsePack(berr.INVALID_PARAMS, "")
		}
		_, notify := bcomn.GetExecuteNotify(eventInfo)
		return responseSuccess(notify)
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}
	return responsePack(berr.INVALID_PARAMS, "")
}

func GetBlockHeightByTxHash(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return responsePack(berr.INVALID_PARAMS, nil)
	}

	switch (params[0]).(type) {
	// tx hash
	case string:
		str := params[0].(string)
		hex, err := hex.DecodeString(str)
		if err != nil {
			return responsePack(berr.INVALID_PARAMS, "")
		}
		var hash common.Uint256
		if err := hash.Deserialize(bytes.NewReader(hex)); err != nil {
			return responsePack(berr.INVALID_PARAMS, "")
		}
		height, tx, err := bactor.GetTxnWithHeightByTxHash(hash)
		if err != nil {
			return responsePack(berr.INVALID_PARAMS, "")
		}
		if tx == nil {
			return responsePack(berr.INVALID_PARAMS, "")
		}
		return responseSuccess(height)
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}
	return responsePack(berr.INVALID_PARAMS, "")
}

func getStoreUint64Value(contractAddress, accountAddress common.Address) (uint64, error) {
	data, err := bactor.GetStorageItem(contractAddress, accountAddress[:])
	if err != nil {
		return 0, fmt.Errorf("GetOntBalanceOf GetStorageItem ont address:%s error:%s", accountAddress.ToBase58(), err)
	}
	if len(data) == 0 {
		return 0, nil
	}
	value, err := serialization.ReadUint64(bytes.NewBuffer(data))
	if err != nil {
		return 0, fmt.Errorf("serialization.ReadUint64:%x error:%s", data, err)
	}
	return value, err
}

func GetBalance(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return responsePack(berr.INVALID_PARAMS, "")
	}
	addrBase58, ok := params[0].(string)
	if !ok {
		return responsePack(berr.INVALID_PARAMS, "")
	}
	address, err := common.AddressFromBase58(addrBase58)
	if err != nil {
		return responsePack(berr.INVALID_PARAMS, "")
	}
	rsp, err := bcomn.GetBalance(address)
	if err != nil {
		log.Errorf("GetBalance address:%s error:%s", addrBase58, err)
		return responsePack(berr.INTERNAL_ERROR, "")
	}
	return responseSuccess(rsp)
}

func GetMerkleProof(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return responsePack(berr.INVALID_PARAMS, "")
	}
	str := params[0].(string)
	hex, err := hex.DecodeString(str)
	if err != nil {
		return responsePack(berr.INVALID_PARAMS, "")
	}
	var hash common.Uint256
	if err := hash.Deserialize(bytes.NewReader(hex)); err != nil {
		return responsePack(berr.INVALID_PARAMS, "")
	}
	height, tx, err := bactor.GetTxnWithHeightByTxHash(hash)
	if err != nil {
		return responsePack(berr.INVALID_PARAMS, "")
	}
	if tx == nil {
		return responsePack(berr.INVALID_PARAMS, "")
	}
	header, err := bactor.GetHeaderByHeight(height)
	if err != nil {
		return responsePack(berr.INVALID_PARAMS, "")
	}

	curHeight := bactor.GetCurrentBlockHeight()
	curHeader, err := bactor.GetHeaderByHeight(curHeight)
	if err != nil {
		return responsePack(berr.INVALID_PARAMS, "")
	}
	proof, err := bactor.GetMerkleProof(uint32(height), uint32(curHeight))
	if err != nil {
		return responsePack(berr.INTERNAL_ERROR, "")
	}
	var hashes []string
	for _, v := range proof {
		hashes = append(hashes, common.ToHexString(v[:]))
	}
	return responseSuccess(bcomn.MerkleProof{"MerkleProof", common.ToHexString(header.TransactionsRoot[:]), height,
		common.ToHexString(curHeader.BlockRoot[:]), curHeight, hashes})
}
