package rpc

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	. "github.com/Ontology/common"
	"github.com/Ontology/common/config"
	"github.com/Ontology/core/types"
	. "github.com/Ontology/errors"
	. "github.com/Ontology/http/base/actor"
	. "github.com/Ontology/http/base/common"
	"math/rand"
	"os"
)

func GetBestBlockHash(params []interface{}) map[string]interface{} {
	hash, err := CurrentBlockHash()
	if err != nil {
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
			return RpcUnknownBlock
		}
		if hash.CompareTo(Uint256{}) == 0{
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
		return RpcUnknownBlock
	}
	if block.Header == nil{
		return RpcUnknownBlock
	}
	return Rpc(GetBlockInfo(block))
}

func GetBlockCount(params []interface{}) map[string]interface{} {
	height, err := BlockHeight()
	if err != nil {
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
		tx, err := GetTransaction(hash) //ledger.DefaultLedger.Store.GetTransaction(hash)
		if err != nil {
			return RpcUnknownTransaction
		}
		if tx == nil {
			return RpcUnknownTransaction
		}
		tran := TransArryByteToHexString(tx)
		return Rpc(tran)
	default:
		return RpcInvalidParameter
	}
}

//   {"jsonrpc": "2.0", "method": "getstorage", "params": ["code hash", "key"], "id": 0}
func GetStorage(params []interface{}) map[string]interface{} {
	if len(params) < 2 {
		return RpcNil
	}

	var codeHash Uint160
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
		var hash Uint160
		if err := hash.Deserialize(bytes.NewReader(hex)); err != nil {
			return RpcInvalidParameter
		}
		contract, err := GetContractStateFromStore(hash)
		if err != nil || contract == nil{
			return RpcInternalError
		}
		return Rpc(TransPayloadToHex(contract))
	default:
		return RpcInvalidParameter
	}
}

func UploadDataFile(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return RpcNil
	}

	rbuf := make([]byte, 4)
	rand.Read(rbuf)
	tmpname := hex.EncodeToString(rbuf)

	str := params[0].(string)

	data, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return RpcInvalidParameter
	}
	f, err := os.OpenFile(tmpname, os.O_WRONLY|os.O_CREATE, 0664)
	if err != nil {
		return RpcIOError
	}
	defer f.Close()
	f.Write(data)

	refpath, err := AddFileIPFS(tmpname, true)
	if err != nil {
		return RpcAPIError
	}

	return Rpc(refpath)

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

func GetTxBlockHeight(params []interface{}) map[string]interface{} {
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
		//TODO resp
		return Rpc(map[string]interface{}{"Height": 0})
	default:
		return RpcInvalidParameter
	}
	return RpcInvalidParameter
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

func GetDataFile(params []interface{}) map[string]interface{} {
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
		tx, err := GetTransaction(hash)
		if err != nil {
			return RpcUnknownTransaction
		}

		tran := TransArryByteToHexString(tx)
		info := tran.Payload.(*DataFileInfo)

		err = GetFileIPFS(info.IPFSPath, info.Filename)
		if err != nil {
			return RpcAPIError
		}
		//TODO: shoud return download address
		return RpcSuccess
	default:
		return RpcInvalidParameter
	}
}
