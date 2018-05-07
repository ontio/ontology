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
	"math/big"
	"strconv"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/genesis"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/types"
	ontErrors "github.com/ontio/ontology/errors"
	bactor "github.com/ontio/ontology/http/base/actor"
	bcomn "github.com/ontio/ontology/http/base/common"
	berr "github.com/ontio/ontology/http/base/error"
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
	height:= bactor.GetCurrentBlockHeight()
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
	hash:= bactor.GetBlockHashFromStore(uint32(height))
	resp["Result"] = common.ToHexString(hash.ToArray())
	return resp
}

func GetBlockTransactions(block *types.Block) interface{} {
	trans := make([]string, len(block.Transactions))
	for i := 0; i < len(block.Transactions); i++ {
		h := block.Transactions[i].Hash()
		trans[i] = common.ToHexString(h.ToArray())
	}
	hash := block.Hash()
	type BlockTransactions struct {
		Hash         string
		Height       uint32
		Transactions []string
	}
	b := BlockTransactions{
		Hash:         common.ToHexString(hash.ToArray()),
		Height:       block.Header.Height,
		Transactions: trans,
	}
	return b
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
	param := cmd["Hash"].(string)
	if len(param) == 0 {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	var getTxBytes = false
	if raw, ok := cmd["Raw"].(string); ok && raw == "1" {
		getTxBytes = true
	}
	var hash common.Uint256
	hex, err := common.HexToBytes(param)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	if err := hash.Deserialize(bytes.NewReader(hex)); err != nil {
		return ResponsePack(berr.INVALID_TRANSACTION)
	}
	resp["Result"], resp["Error"] = getBlock(hash, getTxBytes)
	return resp
}

func GetBlockHeightByTxHash(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	param := cmd["Hash"].(string)
	if len(param) == 0 {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	var hash common.Uint256
	hex, err := common.HexToBytes(param)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	if err := hash.Deserialize(bytes.NewReader(hex)); err != nil {
		return ResponsePack(berr.INVALID_TRANSACTION)
	}
	height,tx, err := bactor.GetTxnWithHeightByTxHash(hash)
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
	hash:= bactor.GetBlockHashFromStore(index)
	if err != nil {
		return ResponsePack(berr.UNKNOWN_BLOCK)
	}
	if hash == common.UINT256_EMPTY {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	block, err := bactor.GetBlockFromStore(hash)
	if err != nil {
		return ResponsePack(berr.UNKNOWN_BLOCK)
	}
	resp["Result"] = GetBlockTransactions(block)
	return resp
}
func GetBlockByHeight(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)

	param := cmd["Height"].(string)
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
		resp["Result"] =  common.ToHexString(w.Bytes())
	} else {
		resp["Result"] = bcomn.GetBlockInfo(block)
	}
	return resp
}

//Transaction
func GetTransactionByHash(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)

	str := cmd["Hash"].(string)
	bys, err := common.HexToBytes(str)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	var hash common.Uint256
	err = hash.Deserialize(bytes.NewReader(bys))
	if err != nil {
		return ResponsePack(berr.INVALID_TRANSACTION)
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
	resp["Result"] = common.ToHexString(hash.ToArray())

	return resp
}

func GetSmartCodeEventTxsByHeight(cmd map[string]interface{}) map[string]interface{} {
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
	txs, err := bactor.GetEventNotifyByHeight(index)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	var txhexs []string
	for _, v := range txs {
		txhexs = append(txhexs, common.ToHexString(v.ToArray()))
	}
	resp["Result"] = txhexs
	return resp
}

func GetSmartCodeEventByTxHash(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)

	str := cmd["Hash"].(string)
	bys, err := common.HexToBytes(str)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	var hash common.Uint256
	err = hash.Deserialize(bytes.NewReader(bys))
	if err != nil {
		return ResponsePack(berr.INVALID_TRANSACTION)
	}
	eventInfos, err := bactor.GetEventNotifyByTxHash(hash)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	var evts []bcomn.NotifyEventInfo
	for _, v := range eventInfos {
		evts = append(evts, bcomn.NotifyEventInfo{common.ToHexString(v.TxHash[:]),v.ContractAddress.ToHexString(), v.States})
	}
	resp["Result"] = evts
	return resp
}

func GetContractState(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	str := cmd["Hash"].(string)
	bys, err := common.HexToBytes(str)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	var hash common.Address
	err = hash.Deserialize(bytes.NewReader(bys))
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	contract, err := bactor.GetContractStateFromStore(hash)
	if err != nil {
		return ResponsePack(berr.INTERNAL_ERROR)
	}
	if contract == nil {
		return ResponsePack(berr.UNKNWN_CONTRACT)
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
	str := cmd["Hash"].(string)
	bys, err := common.HexToBytes(str)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	var hash common.Address
	err = hash.Deserialize(bytes.NewReader(bys))
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	key := cmd["Key"].(string)
	item, err := common.HexToBytes(key)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	value, err := bactor.GetStorageItem(hash, item)
	if err != nil || value == nil {
		return ResponsePack(berr.INTERNAL_ERROR)
	}
	resp["Result"] = common.ToHexString(value)
	return resp
}

func GetBalance(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	addrBase58 := cmd["Addr"].(string)
	address, err := common.AddressFromBase58(addrBase58)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	ont := big.NewInt(0)
	ong := big.NewInt(0)
	appove := big.NewInt(0)

	ontBalance, err := bactor.GetStorageItem(genesis.OntContractAddress, address[:])
	if err != nil {
		log.Errorf("GetOntBalanceOf GetStorageItem ont address:%s error:%s", address, err)
		return ResponsePack(berr.INTERNAL_ERROR)
	}
	if ontBalance != nil {
		ont.SetBytes(ontBalance)
	}

	ongBalance, err := bactor.GetStorageItem(genesis.OngContractAddress, address[:])
	if err != nil {
		log.Errorf("GetOngBalanceOf GetStorageItem ong address:%s error:%s", address, err)
		return ResponsePack(berr.INTERNAL_ERROR)
	}
	if ongBalance != nil {
		ong.SetBytes(ongBalance)
	}

	appoveKey := append(genesis.OntContractAddress[:], address[:]...)
	ongappove, err := bactor.GetStorageItem(genesis.OngContractAddress, appoveKey[:])

	if ongappove != nil {
		appove.SetBytes(ongappove)
	}
	rsp := &bcomn.BalanceOfRsp{
		Ont:       ont.String(),
		Ong:       ong.String(),
		OngAppove: appove.String(),
	}
	resp["Result"] = rsp
	return resp
}

func GetMerkleProof(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	str := cmd["Hash"].(string)
	bys, err := common.HexToBytes(str)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	var hash common.Uint256
	err = hash.Deserialize(bytes.NewReader(bys))
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

	curHeight:= bactor.GetCurrentBlockHeight()
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
		hashes = append(hashes, common.ToHexString(v[:]))
	}
	resp["Result"] = bcomn.MerkleProof{"MerkleProof", common.ToHexString(header.TransactionsRoot[:]), height,
		common.ToHexString(curHeader.BlockRoot[:]), curHeight, hashes}
	return resp
}