package rest

import (
	"bytes"
	. "github.com/Ontology/common"
	"github.com/Ontology/common/config"
	"github.com/Ontology/core/types"
	. "github.com/Ontology/errors"
	Err "github.com/Ontology/http/base/error"
	. "github.com/Ontology/net/protocol"
	"strconv"
	"github.com/Ontology/smartcontract/pre_exec"
	"github.com/Ontology/core/payload"
	"github.com/Ontology/common/log"
	. "github.com/Ontology/http/base/common"
	. "github.com/Ontology/http/base/actor"
)

var node Noder

const TlsPort int = 443

type ApiServer interface {
	Start() error
	Stop()
}


//Node
func GetGenerateBlockTime(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(Err.SUCCESS)
	if node != nil {
		resp["Result"] = config.DEFAULTGENBLOCKTIME
	}
	return resp
}
func GetConnectionCount(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(Err.SUCCESS)
	if node != nil {
		count,err := GetConnectionCnt()
		if err != nil{
			resp["Error"] = Err.INTERNAL_ERROR
			return resp
		}
		resp["Result"] = count
	}

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
		resp["Error"] = Err.INVALID_PARAMS
		return resp
	}
	height, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		resp["Error"] = Err.INVALID_PARAMS
		return resp
	}
	hash, err := GetBlockHashFromStore(uint32(height))
	if err != nil {
		resp["Error"] = Err.INVALID_PARAMS
		return resp
	}
	resp["Result"] = ToHexString(hash.ToArray())
	return resp
}

func GetBlockInfo(block *types.Block) BlockInfo {
	hash := block.Hash()
	blockHead := &BlockHead{
		Version:          block.Header.Version,
		PrevBlockHash:    ToHexString(block.Header.PrevBlockHash.ToArray()),
		TransactionsRoot: ToHexString(block.Header.TransactionsRoot.ToArray()),
		BlockRoot:        ToHexString(block.Header.BlockRoot.ToArray()),
		StateRoot:        ToHexString(block.Header.StateRoot.ToArray()),
		Timestamp:        block.Header.Timestamp,
		Height:           block.Header.Height,
		ConsensusData:    block.Header.ConsensusData,
		NextBookKeeper:   ToHexString(block.Header.NextBookKeeper[:]),
		Program: ProgramInfo{
			Code:      ToHexString(block.Header.Program.Code),
			Parameter: ToHexString(block.Header.Program.Parameter),
		},
		Hash: ToHexString(hash.ToArray()),
	}

	trans := make([]*Transactions, len(block.Transactions))
	for i := 0; i < len(block.Transactions); i++ {
		trans[i] = TransArryByteToHexString(block.Transactions[i])
	}

	b := BlockInfo{
		Hash:         ToHexString(hash.ToArray()),
		BlockData:    blockHead,
		Transactions: trans,
	}
	return b
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
		resp["Error"] = Err.INVALID_PARAMS
		return resp
	}
	var getTxBytes bool = false
	if raw, ok := cmd["Raw"].(string); ok && raw == "1" {
		getTxBytes = true
	}
	var hash Uint256
	hex, err := HexToBytes(param)
	if err != nil {
		resp["Error"] = Err.INVALID_PARAMS
		return resp
	}
	if err := hash.Deserialize(bytes.NewReader(hex)); err != nil {
		resp["Error"] = Err.INVALID_TRANSACTION
		return resp
	}

	resp["Result"], resp["Error"] = getBlock(hash, getTxBytes)

	return resp
}
func GetBlockTxsByHeight(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(Err.SUCCESS)

	param := cmd["Height"].(string)
	if len(param) == 0 {
		resp["Error"] = Err.INVALID_PARAMS
		return resp
	}
	height, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		resp["Error"] = Err.INVALID_PARAMS
		return resp
	}
	index := uint32(height)
	hash, err := GetBlockHashFromStore(index)
	if err != nil {
		resp["Error"] = Err.UNKNOWN_BLOCK
		return resp
	}
	block, err := GetBlockFromStore(hash)
	if err != nil {
		resp["Error"] = Err.UNKNOWN_BLOCK
		return resp
	}
	resp["Result"] = GetBlockTransactions(block)
	return resp
}
func GetBlockByHeight(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(Err.SUCCESS)

	param := cmd["Height"].(string)
	if len(param) == 0 {
		resp["Error"] = Err.INVALID_PARAMS
		return resp
	}
	var getTxBytes bool = false
	if raw, ok := cmd["Raw"].(string); ok && raw == "1" {
		getTxBytes = true
	}
	height, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		resp["Error"] = Err.INVALID_PARAMS
		return resp
	}
	index := uint32(height)
	hash, err := GetBlockHashFromStore(index)
	if err != nil {
		resp["Error"] = Err.UNKNOWN_BLOCK
		return resp
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
		resp["Error"] = Err.INVALID_PARAMS
		return resp
	}
	var hash Uint256
	err = hash.Deserialize(bytes.NewReader(bys))
	if err != nil {
		resp["Error"] = Err.INVALID_TRANSACTION
		return resp
	}
	tx, err := GetTransaction(hash)
	if err != nil {
		resp["Error"] = Err.UNKNOWN_TRANSACTION
		return resp
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
		resp["Error"] = Err.INVALID_PARAMS
		return resp
	}
	bys, err := HexToBytes(str)
	if err != nil {
		resp["Error"] = Err.INVALID_PARAMS
		return resp
	}
	var txn types.Transaction
	if err := txn.Deserialize(bytes.NewReader(bys)); err != nil {
		resp["Error"] = Err.INVALID_TRANSACTION
		return resp
	}
	if txn.TxType == types.Invoke {
		if preExec, ok := cmd["PreExec"].(string); ok && preExec == "1" {
			log.Tracef("PreExec SMARTCODE")
			if invokeCode,ok := txn.Payload.(*payload.InvokeCode);ok{
				param := invokeCode.Code.Code
				codeHash := ToCodeHash(param)
				param = append(param, 0x67)
				param = append(param, codeHash.ToArray()...)
				resp["Result"], err = pre_exec.PreExec(param, &txn)
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
		resp["Error"] = Err.INVALID_PARAMS
		return resp
	}
	height, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		resp["Error"] = Err.INVALID_PARAMS
		return resp
	}
	index := uint32(height)
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
func GetContract(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(Err.SUCCESS)
	str := cmd["Hash"].(string)
	bys, err := HexToBytes(str)
	if err != nil {
		resp["Error"] = Err.INVALID_PARAMS
		return resp
	}
	var hash Uint160
	err = hash.Deserialize(bytes.NewReader(bys))
	if err != nil {
		resp["Error"] = Err.INVALID_PARAMS
		return resp
	}
	//TODO GetContract from store
	//contract, err := GetContractFromStore(hash)
	//if err != nil {
	//	resp["Error"] = Err.INVALID_PARAMS
	//	return resp
	//}
	//resp["Result"] = string(contract)
	return resp
}
