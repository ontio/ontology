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
	. "github.com/Ontology/common"
	"github.com/Ontology/common/config"
	"github.com/Ontology/common/log"
	"github.com/Ontology/core/genesis"
	"github.com/Ontology/core/types"
	. "github.com/Ontology/errors"
	. "github.com/Ontology/http/base/actor"
	. "github.com/Ontology/http/base/common"
	"math/big"
)

func GetGenerateBlockTime(params []interface{}) map[string]interface{} {
	return Rpc(config.DEFAULTGENBLOCKTIME)
}

func GetBestBlockHash(params []interface{}) map[string]interface{} {
	hash, err := CurrentBlockHash()
	if err != nil {
		log.Errorf("GetBestBlockHash error:%s", err)
		return RpcFailed
	}
	return Rpc(ToHexString(hash.ToArray()))
}

// Input JSON string examples for getblock method as following:
//   {"jsonrpc": "2.0", "method": "getblock", "params": [1], "id": 0}
//   {"jsonrpc": "2.0", "method": "getblock", "params": ["aabbcc.."], "id": 0}
func GetBlock(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return RpcNil
	}
	var err error
	var hash Uint256
	switch (params[0]).(type) {
	// block height
	case float64:
		index := uint32(params[0].(float64))
		hash, err = GetBlockHashFromStore(index)
		if err != nil {
			log.Errorf("GetBlock GetBlockHashFromStore Height:%v error:%s", index, err)
			return RpcUnknownBlock
		}
		if hash.CompareTo(Uint256{}) == 0 {
			return RpcInvalidParameter
		}
		// block hash
	case string:
		str := params[0].(string)
		hex, err := hex.DecodeString(str)
		if err != nil {
			return RpcInvalidParameter
		}
		if err := hash.Deserialize(bytes.NewReader(hex)); err != nil {
			return RpcInvalidTransaction
		}
	default:
		return RpcInvalidParameter
	}
	block, err := GetBlockFromStore(hash)
	if err != nil {
		log.Errorf("GetBlock GetBlockFromStore BlockHash:%x error:%s", hash, err)
		return RpcUnknownBlock
	}
	if block == nil || block.Header == nil {
		return RpcUnknownBlock
	}
	if len(params) >= 2 {
		switch (params[1]).(type) {
		case float64:
			json := uint32(params[1].(float64))
			if json == 1{
				return Rpc(GetBlockInfo(block))
			}
		default:
			return RpcInvalidParameter
		}
	}
	w := bytes.NewBuffer(nil)
	block.Serialize(w)
	return Rpc(ToHexString(w.Bytes()))
}

func GetBlockCount(params []interface{}) map[string]interface{} {
	height, err := BlockHeight()
	if err != nil {
		log.Errorf("GetBlockCount error:%s", err)
		return RpcFailed
	}
	return Rpc(height + 1)
}

// A JSON example for getblockhash method as following:
//   {"jsonrpc": "2.0", "method": "getblockhash", "params": [1], "id": 0}
func GetBlockHash(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return RpcNil
	}
	switch params[0].(type) {
	case float64:
		height := uint32(params[0].(float64))
		hash, err := GetBlockHashFromStore(height)
		if err != nil {
			log.Errorf("GetBlockHash GetBlockHashFromStore height:%v error:%s", height, err)
			return RpcUnknownBlock
		}
		return Rpc(fmt.Sprintf("%016x", hash))
	default:
		return RpcInvalidParameter
	}
}

func GetConnectionCount(params []interface{}) map[string]interface{} {
	count, err := GetConnectionCnt()
	if err != nil {
		log.Errorf("GetConnectionCount error:%s", err)
		return RpcFailed
	}
	return Rpc(count)
}

func GetRawMemPool(params []interface{}) map[string]interface{} {
	txs := []*Transactions{}
	txpool, _ := GetTxsFromPool(false)
	for _, t := range txpool {
		txs = append(txs, TransArryByteToHexString(t))
	}
	if len(txs) == 0 {
		return RpcNil
	}
	return Rpc(txs)
}
func GetMemPoolTxState(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return RpcNil
	}
	switch params[0].(type) {
	case string:
		str := params[0].(string)
		hex, err := hex.DecodeString(str)
		if err != nil {
			return RpcInvalidParameter
		}
		var hash Uint256
		err = hash.Deserialize(bytes.NewReader(hex))
		if err != nil {
			return RpcInvalidTransaction
		}
		txEntry, err := GetTxFromPool(hash)
		if err != nil {
			return RpcUnknownTransaction
		}
		tran := TransArryByteToHexString(txEntry.Tx)
		attrs := []TXNAttrInfo{}
		for _, t := range txEntry.Attrs {
			attrs = append(attrs, TXNAttrInfo{t.Height, int(t.Type), int(t.ErrCode)})
		}
		info := TXNEntryInfo{*tran, int64(txEntry.Fee), attrs}
		return Rpc(info)
	default:
		return RpcInvalidParameter
	}
}

// A JSON example for getrawtransaction method as following:
//   {"jsonrpc": "2.0", "method": "getrawtransaction", "params": ["transactioin hash in hex"], "id": 0}
func GetRawTransaction(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return RpcNil
	}
	var tx *types.Transaction
	switch params[0].(type) {
	case string:
		str := params[0].(string)
		hex, err := hex.DecodeString(str)
		if err != nil {
			return RpcInvalidParameter
		}
		var hash Uint256
		err = hash.Deserialize(bytes.NewReader(hex))
		if err != nil {
			return RpcInvalidTransaction
		}
		t, err := GetTransaction(hash)
		if err != nil {
			log.Errorf("GetRawTransaction GetTransaction error:%s", err)
			return RpcUnknownTransaction
		}
		tx = t
		if tx == nil {
			return RpcUnknownTransaction
		}
	default:
		return RpcInvalidParameter
	}

	if len(params) >= 2 {
		switch (params[1]).(type) {
		case float64:
			json := uint32(params[1].(float64))
			if json == 1{
				return Rpc(TransArryByteToHexString(tx))
			}
		default:
			return RpcInvalidParameter
		}
	}
	w := bytes.NewBuffer(nil)
	tx.Serialize(w)
	return Rpc(ToHexString(w.Bytes()))
}

//   {"jsonrpc": "2.0", "method": "getstorage", "params": ["code hash", "key"], "id": 0}
func GetStorage(params []interface{}) map[string]interface{} {
	if len(params) < 2 {
		return RpcNil
	}

	var codeHash Address
	var key []byte
	switch params[0].(type) {
	case string:
		str := params[0].(string)
		hex, err := hex.DecodeString(str)
		if err != nil {
			return RpcInvalidParameter
		}
		if err := codeHash.Deserialize(bytes.NewReader(hex)); err != nil {
			return RpcInvalidHash
		}
	default:
		return RpcInvalidParameter
	}

	switch params[1].(type) {
	case string:
		str := params[1].(string)
		hex, err := hex.DecodeString(str)
		if err != nil {
			return RpcInvalidParameter
		}
		key = hex
	default:
		return RpcInvalidParameter
	}
	value, err := GetStorageItem(codeHash, key)
	if err != nil {
		log.Errorf("GetStorage GetStorageItem CodeHash:%x key:%s error:%s", codeHash, key, err)
		return RpcInternalError
	}
	return Rpc(ToHexString(value))
}

// A JSON example for sendrawtransaction method as following:
//   {"jsonrpc": "2.0", "method": "sendrawtransaction", "params": ["raw transactioin in hex"], "id": 0}
func SendRawTransaction(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return RpcNil
	}
	var hash Uint256
	switch params[0].(type) {
	case string:
		str := params[0].(string)
		hex, err := hex.DecodeString(str)
		if err != nil {
			return RpcInvalidParameter
		}
		var txn types.Transaction
		if err := txn.Deserialize(bytes.NewReader(hex)); err != nil {
			return RpcInvalidTransaction
		}
		hash = txn.Hash()
		if errCode := VerifyAndSendTx(&txn); errCode != ErrNoError {
			return Rpc(errCode.Error())
		}
	default:
		return RpcInvalidParameter
	}
	return Rpc(ToHexString(hash.ToArray()))
}

// A JSON example for submitblock method as following:
//   {"jsonrpc": "2.0", "method": "submitblock", "params": ["raw block in hex"], "id": 0}
func SubmitBlock(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return RpcNil
	}
	switch params[0].(type) {
	case string:
		str := params[0].(string)
		hex, _ := hex.DecodeString(str)
		var block types.Block
		if err := block.Deserialize(bytes.NewReader(hex)); err != nil {
			return RpcInvalidBlock
		}
		if err := AddBlock(&block); err != nil {
			return RpcInvalidBlock
		}
		if err := Xmit(&block); err != nil {
			return RpcInternalError
		}
	default:
		return RpcInvalidParameter
	}
	return RpcSuccess
}

func GetNodeVersion(params []interface{}) map[string]interface{} {
	return Rpc(config.Parameters.Version)
}

func GetSystemFee(params []interface{}) map[string]interface{} {
	return Rpc(config.Parameters.SystemFee)
}

func GetContractState(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return RpcNil
	}
	switch params[0].(type) {
	case string:
		str := params[0].(string)
		hex, err := hex.DecodeString(str)
		if err != nil {
			return RpcInvalidParameter
		}
		var hash Address
		if err := hash.Deserialize(bytes.NewReader(hex)); err != nil {
			return RpcInvalidParameter
		}
		contract, err := GetContractStateFromStore(hash)
		if err != nil {
			log.Errorf("GetContractState GetContractStateFromStore hash:%x error:%s", hash, err)
			return RpcInternalError
		}
		if contract == nil {
			return RpcUnKnownContact
		}
		return Rpc(TransPayloadToHex(contract))
	default:
		return RpcInvalidParameter
	}
}

func GetSmartCodeEvent(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return RpcNil
	}

	switch (params[0]).(type) {
	// block height
	case float64:
		height := uint32(params[0].(float64))
		//TODO resp
		return Rpc(map[string]interface{}{"Height": height})
	default:
		return RpcInvalidParameter
	}
	return RpcInvalidParameter
}

func GetBlockHeightByTxHash(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return RpcNil
	}

	switch (params[0]).(type) {
	// tx hash
	case string:
		str := params[0].(string)
		hex, err := hex.DecodeString(str)
		if err != nil {
			return RpcInvalidParameter
		}
		var hash Uint256
		if err := hash.Deserialize(bytes.NewReader(hex)); err != nil {
			return RpcInvalidParameter
		}
		height,err := GetBlockHeightByTxHashFromStore(hash)
		if err != nil{
			return RpcInvalidParameter
		}
		return Rpc(height)
	default:
		return RpcInvalidParameter
	}
	return RpcInvalidParameter
}

func GetBalance(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return RpcInvalidParameter
	}
	addrBase58, ok := params[0].(string)
	if !ok {
		return RpcInvalidParameter
	}
	address, err := AddressFromBase58(addrBase58)
	if err != nil {
		return RpcInvalidParameter
	}
	ont := new(big.Int)
	ong := new(big.Int)

	ontBalance, err := GetStorageItem(genesis.OntContractAddress, address.ToArray())
	if err != nil {
		log.Errorf("GetOntBalanceOf GetStorageItem ont address:%s error:%s", addrBase58, err)
		return RpcInternalError
	}
	if ontBalance != nil {
		ont.SetBytes(ontBalance)
	}
	rsp := &BalanceOfRsp{
		Ont: ont.String(),
		Ong: ong.String(),
	}
	return Rpc(rsp)
}

func RegDataFile(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return RpcNil
	}
	var hash Uint256
	switch params[0].(type) {
	case string:
		str := params[0].(string)
		hex, err := hex.DecodeString(str)
		if err != nil {
			return RpcInvalidParameter
		}
		var txn types.Transaction
		if err := txn.Deserialize(bytes.NewReader(hex)); err != nil {
			return RpcInvalidTransaction
		}

		hash = txn.Hash()
		if errCode := VerifyAndSendTx(&txn); errCode != ErrNoError {
			return RpcInternalError
		}
	default:
		return RpcInvalidParameter
	}
	return Rpc(ToHexString(hash.ToArray()))
}

func CatDataRecord(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return RpcNil
	}
	switch params[0].(type) {
	case string:
		str := params[0].(string)
		b, err := hex.DecodeString(str)
		if err != nil {
			return RpcInvalidParameter
		}
		var hash Uint256
		err = hash.Deserialize(bytes.NewReader(b))
		if err != nil {
			return RpcInvalidTransaction
		}
		tx, err := GetTransaction(hash)
		if err != nil {
			return RpcUnknownTransaction
		}
		tran := TransArryByteToHexString(tx)
		info := tran.Payload.(*DataFileInfo)
		//ref := string(record.RecordData[:])
		return Rpc(info)
	default:
		return RpcInvalidParameter
	}
}
