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
		resp["Error"] = Err.INTERNAL_ERROR
		return resp
	}
	resp["Result"] = count

	return resp
}

//Block
func GetBlockHeight(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(Err.SUCCESS)
	height,err := BlockHeight()
	if err != nil{
		resp["Error"] = Err.INTERNAL_ERROR
		return resp
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
		return "", Err.UNKNOWN_BLOCK
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
	var getTxBytes bool = false
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

func GetTxBlockHeight(cmd map[string]interface{}) map[string]interface{} {
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
	height,err := GetTxBlockHeightFromStore(hash)
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
					resp["Error"] = Err.SMARTCODE_ERROR
					return resp
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
	//infos,_ := GetEventNotifyByTx(Uint256{})
	resp["Result"] = map[string]interface{}{"Height": index}
	//TODO resp
	return resp
}
func ResponsePack(errCode int64) map[string]interface{} {
	resp := map[string]interface{}{
		"Action":  "",
		"Result":  "",
		"Error":   errCode,
		"Desc":    "",
		"Version": "1.0.0",
	}
	return resp
}
func GetContractState(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(Err.SUCCESS)
	str := cmd["Hash"].(string)
	bys, err := HexToBytes(str)
	if err != nil {
		return ResponsePack(Err.INVALID_PARAMS)
	}
	var hash Uint160
	err = hash.Deserialize(bytes.NewReader(bys))
	if err != nil {
		return ResponsePack(Err.INVALID_PARAMS)
	}
	contract, err := GetContractStateFromStore(hash)
	if err != nil || contract == nil {
		resp["Error"] = Err.INVALID_PARAMS
		return resp
	}
	resp["Result"] = TransPayloadToHex(contract)
	return resp
}
