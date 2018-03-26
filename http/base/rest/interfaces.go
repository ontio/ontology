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
	. "github.com/Ontology/common"
	"github.com/Ontology/common/config"
	"github.com/Ontology/core/types"
	. "github.com/Ontology/errors"
	Err "github.com/Ontology/http/base/error"
	"strconv"
	"github.com/Ontology/core/payload"
	"github.com/Ontology/common/log"
	. "github.com/Ontology/http/base/common"
	. "github.com/Ontology/http/base/actor"
	"math/big"
	"github.com/Ontology/core/genesis"
)

const TlsPort int = 443

type ApiServer interface {
	Start() error
	Stop()
}

//Node
func GetGenerateBlockTime(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(Err.SUCCESS)
	resp["Result"] = config.DEFAULTGENBLOCKTIME
	return resp
}
func GetConnectionCount(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(Err.SUCCESS)
	count, err := GetConnectionCnt()
	if err != nil {
		return ResponsePack(Err.INTERNAL_ERROR)
	}
	resp["Result"] = count
	return resp
}

//Block
func GetBlockHeight(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(Err.SUCCESS)
	height,err := BlockHeight()
	if err != nil{
		return ResponsePack(Err.INTERNAL_ERROR)
	}
	resp["Result"] = height
	return resp
}
func GetBlockHash(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(Err.SUCCESS)
	param := cmd["Height"].(string)
	if len(param) == 0 {
		return ResponsePack(Err.INVALID_PARAMS)
	}
	height, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		return ResponsePack(Err.INVALID_PARAMS)
	}
	hash, err := GetBlockHashFromStore(uint32(height))
	if err != nil {
		return ResponsePack(Err.INVALID_PARAMS)
	}
	resp["Result"] = ToHexString(hash.ToArray())
	return resp
}

func GetBlockTransactions(block *types.Block) interface{} {
	trans := make([]string, len(block.Transactions))
	for i := 0; i < len(block.Transactions); i++ {
		h := block.Transactions[i].Hash()
		trans[i] = ToHexString(h.ToArray())
	}
	hash := block.Hash()
	type BlockTransactions struct {
		Hash         string
		Height       uint32
		Transactions []string
	}
	b := BlockTransactions{
		Hash:         ToHexString(hash.ToArray()),
		Height:       block.Header.Height,
		Transactions: trans,
	}
	return b
}
func getBlock(hash Uint256, getTxBytes bool) (interface{}, int64) {
	block, err := GetBlockFromStore(hash)
	if err != nil {
		return nil, Err.UNKNOWN_BLOCK
	}
	if block.Header == nil {
		return nil, Err.UNKNOWN_BLOCK
	}
	if getTxBytes {
		w := bytes.NewBuffer(nil)
		block.Serialize(w)
		return ToHexString(w.Bytes()), Err.SUCCESS
	}
	return GetBlockInfo(block), Err.SUCCESS
}
func GetBlockByHash(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(Err.SUCCESS)
	param := cmd["Hash"].(string)
	if len(param) == 0 {
		return ResponsePack(Err.INVALID_PARAMS)
	}
	var getTxBytes = false
	if raw, ok := cmd["Raw"].(string); ok && raw == "1" {
		getTxBytes = true
	}
	var hash Uint256
	hex, err := HexToBytes(param)
	if err != nil {
		return ResponsePack(Err.INVALID_PARAMS)
	}
	if err := hash.Deserialize(bytes.NewReader(hex)); err != nil {
		return ResponsePack(Err.INVALID_TRANSACTION)
	}
	resp["Result"], resp["Error"] = getBlock(hash, getTxBytes)
	return resp
}

func GetBlockHeightByTxHash(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(Err.SUCCESS)
	param := cmd["Hash"].(string)
	if len(param) == 0 {
		return ResponsePack(Err.INVALID_PARAMS)
	}
	var hash Uint256
	hex, err := HexToBytes(param)
	if err != nil {
		return ResponsePack(Err.INVALID_PARAMS)
	}
	if err := hash.Deserialize(bytes.NewReader(hex)); err != nil {
		return ResponsePack(Err.INVALID_TRANSACTION)
	}
	height,err := GetBlockHeightByTxHashFromStore(hash)
	if err != nil {
		return ResponsePack(Err.INTERNAL_ERROR)
	}
	resp["Result"] = height
	return resp
}
func GetBlockTxsByHeight(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(Err.SUCCESS)

	param := cmd["Height"].(string)
	if len(param) == 0 {
		return ResponsePack(Err.INVALID_PARAMS)
	}
	height, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		return ResponsePack(Err.INVALID_PARAMS)
	}
	index := uint32(height)
	hash, err := GetBlockHashFromStore(index)
	if err != nil {
		return ResponsePack(Err.UNKNOWN_BLOCK)
	}
	if hash.CompareTo(Uint256{}) == 0{
		return ResponsePack(Err.INVALID_PARAMS)
	}
	block, err := GetBlockFromStore(hash)
	if err != nil {
		return ResponsePack(Err.UNKNOWN_BLOCK)
	}
	resp["Result"] = GetBlockTransactions(block)
	return resp
}
func GetBlockByHeight(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(Err.SUCCESS)

	param := cmd["Height"].(string)
	if len(param) == 0 {
		return ResponsePack(Err.INVALID_PARAMS)
	}
	var getTxBytes bool = false
	if raw, ok := cmd["Raw"].(string); ok && raw == "1" {
		getTxBytes = true
	}
	height, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		return ResponsePack(Err.INVALID_PARAMS)
	}
	index := uint32(height)
	hash, err := GetBlockHashFromStore(index)
	if err != nil {
		return ResponsePack(Err.UNKNOWN_BLOCK)
	}
	resp["Result"], resp["Error"] = getBlock(hash, getTxBytes)
	return resp
}


//Transaction
func GetTransactionByHash(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(Err.SUCCESS)

	str := cmd["Hash"].(string)
	bys, err := HexToBytes(str)
	if err != nil {
		return ResponsePack(Err.INVALID_PARAMS)
	}
	var hash Uint256
	err = hash.Deserialize(bytes.NewReader(bys))
	if err != nil {
		return ResponsePack(Err.INVALID_TRANSACTION)
	}
	tx, err := GetTransaction(hash)
	if err != nil {
		return ResponsePack(Err.UNKNOWN_TRANSACTION)
	}
	if tx == nil {
		return ResponsePack(Err.UNKNOWN_TRANSACTION)
	}
	if raw, ok := cmd["Raw"].(string); ok && raw == "1" {
		w := bytes.NewBuffer(nil)
		tx.Serialize(w)
		resp["Result"] = ToHexString(w.Bytes())
		return resp
	}
	tran := TransArryByteToHexString(tx)
	resp["Result"] = tran
	return resp
}
func SendRawTransaction(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(Err.SUCCESS)

	str, ok := cmd["Data"].(string)
	if !ok {
		return ResponsePack(Err.INVALID_PARAMS)
	}
	bys, err := HexToBytes(str)
	if err != nil {
		return ResponsePack(Err.INVALID_PARAMS)
	}
	var txn types.Transaction
	if err := txn.Deserialize(bytes.NewReader(bys)); err != nil {
		return ResponsePack(Err.INVALID_TRANSACTION)
	}
	if txn.TxType == types.Invoke {
		if preExec, ok := cmd["PreExec"].(string); ok && preExec == "1" {
			log.Tracef("PreExec SMARTCODE")
			if _, ok := txn.Payload.(*payload.InvokeCode); ok {
				resp["Result"], err = PreExecuteContract(&txn)
				if err != nil {
					log.Error(err)
					return ResponsePack(Err.SMARTCODE_ERROR)
				}
				return resp
			}
		}
	}
	var hash Uint256
	hash = txn.Hash()
	if errCode := VerifyAndSendTx(&txn); errCode != ErrNoError {
		resp["Error"] = int64(errCode)
		return resp
	}
	resp["Result"] = ToHexString(hash.ToArray())

	if txn.TxType == types.Invoke {
		if userid, ok := cmd["Userid"].(string); ok && len(userid) > 0 {
			resp["Userid"] = userid
		}
	}
	return resp
}

func GetSmartCodeEventByHeight(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(Err.SUCCESS)

	param := cmd["Height"].(string)
	if len(param) == 0 {
		return ResponsePack(Err.INVALID_PARAMS)
	}
	height, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		return ResponsePack(Err.INVALID_PARAMS)
	}
	index := uint32(height)
	txs, err := GetEventNotifyByHeight(index)
	if err != nil {
		return ResponsePack(Err.INVALID_PARAMS)
	}
	var txhexs []string
	for _, v := range txs {
		txhexs = append(txhexs, ToHexString(v.ToArray()))
	}
	resp["Result"] = txhexs
	return resp
}

func GetSmartCodeEventByTxHash(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(Err.SUCCESS)

	str := cmd["Hash"].(string)
	bys, err := HexToBytes(str)
	if err != nil {
		return ResponsePack(Err.INVALID_PARAMS)
	}
	var hash Uint256
	err = hash.Deserialize(bytes.NewReader(bys))
	if err != nil {
		return ResponsePack(Err.INVALID_TRANSACTION)
	}
	eventInfos, err := GetEventNotifyByTxHash(hash)
	if err != nil {
		return ResponsePack(Err.INVALID_PARAMS)
	}
	var evs []map[string]interface{}
	for _, v := range eventInfos {
		evs = append(evs, map[string]interface{}{"CodeHash": v.CodeHash,
			"States": v.States,
			"Container": v.Container})
	}
	resp["Result"] = evs
	return resp
}

func GetContractState(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(Err.SUCCESS)
	str := cmd["Hash"].(string)
	bys, err := HexToBytes(str)
	if err != nil {
		return ResponsePack(Err.INVALID_PARAMS)
	}
	var hash Address
	err = hash.Deserialize(bytes.NewReader(bys))
	if err != nil {
		return ResponsePack(Err.INVALID_PARAMS)
	}
	contract, err := GetContractStateFromStore(hash)
	if err != nil || contract == nil {
		return ResponsePack(Err.INTERNAL_ERROR)
	}
	resp["Result"] = TransPayloadToHex(contract)
	return resp
}

func GetStorage(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(Err.SUCCESS)
	str := cmd["Hash"].(string)
	bys, err := HexToBytes(str)
	if err != nil {
		return ResponsePack(Err.INVALID_PARAMS)
	}
	var hash Address
	err = hash.Deserialize(bytes.NewReader(bys))
	if err != nil {
		return ResponsePack(Err.INVALID_PARAMS)
	}
	key := cmd["Key"].(string)
	item, err := HexToBytes(key)
	if err != nil {
		return ResponsePack(Err.INVALID_PARAMS)
	}
	log.Info("[GetStorage] ",str,key)
	value, err := GetStorageItem(hash,item)
	if err != nil || value == nil {
		return ResponsePack(Err.INTERNAL_ERROR)
	}
	resp["Result"] = ToHexString(value)
	return resp
}
func GetBalance(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(Err.SUCCESS)
	addrBase58 := cmd["Addr"].(string)
	address, err := AddressFromBase58(addrBase58)
	if err != nil {
		return ResponsePack(Err.INVALID_PARAMS)
	}
	ont := new(big.Int)
	ong := new(big.Int)

	ontBalance, err := GetStorageItem(genesis.OntContractAddress, address[:])
	if err != nil {
		log.Errorf("GetOntBalanceOf GetStorageItem ont address:%s error:%s", address, err)
		return ResponsePack(Err.INTERNAL_ERROR)
	}
	if ontBalance != nil {
		ont.SetBytes(ontBalance)
	}
	rsp := &BalanceOfRsp{
		Ont: ont.String(),
		Ong: ong.String(),
	}
	resp["Result"] = rsp
	return resp
}