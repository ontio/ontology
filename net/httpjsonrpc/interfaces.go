package httpjsonrpc

import (
	. "DNA/common"
	"DNA/common/log"
	"DNA/core/ledger"
	tx "DNA/core/transaction"
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
)

func getBestBlockHash(req *http.Request, cmd map[string]interface{}) map[string]interface{} {
	id := cmd["id"]
	hash := ledger.DefaultLedger.Blockchain.CurrentBlockHash()
	response := responsePacking(ToHexString(hash.ToArray()), id)
	return response
}

func getBlock(req *http.Request, cmd map[string]interface{}) map[string]interface{} {
	id := cmd["id"]
	params := cmd["params"]
	var err error
	var hash Uint256
	switch (params.([]interface{})[0]).(type) {
	// the value type is float64 after unmarshal JSON number into an interface value
	case float64:
		index := uint32(params.([]interface{})[0].(float64))
		hash, err = ledger.DefaultLedger.Store.GetBlockHash(index)
		if err != nil {
			return responsePacking([]interface{}{-100, "Unknown block hash"}, id)
		}
	case string:
		hashstr := params.([]interface{})[0].(string)
		hashslice, _ := hex.DecodeString(hashstr)
		hash.Deserialize(bytes.NewReader(hashslice[0:32]))
	}
	block, err := ledger.DefaultLedger.Store.GetBlock(hash)
	if err != nil {
		return responsePacking([]interface{}{-100, "Unknown block"}, id)
	}
	b := BlockInfo{
		Hash:      ToHexString(hash.ToArray()),
		BlockData: block.Blockdata,
	}
	return responsePacking(b, id)
}

func getBlockCount(req *http.Request, cmd map[string]interface{}) map[string]interface{} {
	id := cmd["id"]
	count := ledger.DefaultLedger.Blockchain.BlockHeight + 1
	return responsePacking(count, id)
}

func getBlockHash(req *http.Request, cmd map[string]interface{}) map[string]interface{} {
	id := cmd["id"]
	index := cmd["params"]
	var hash Uint256
	height, ok := index.(uint32)
	if ok == true {
		hash, _ = ledger.DefaultLedger.Store.GetBlockHash(height)
	}
	hashhex := fmt.Sprintf("%016x", hash)
	return responsePacking(hashhex, id)
}

func getConnectionCount(req *http.Request, cmd map[string]interface{}) map[string]interface{} {
	id := cmd["id"]
	count := node.GetConnectionCnt()
	return responsePacking(count, id)
}

func getRawMemPool(req *http.Request, cmd map[string]interface{}) map[string]interface{} {
	id := cmd["id"]
	mempoollist := node.GetTxnPool(false)
	return responsePacking(mempoollist, id)
}

func getRawTransaction(req *http.Request, cmd map[string]interface{}) map[string]interface{} {
	id := cmd["id"]
	params := cmd["params"]
	txid := params.([]interface{})[0].(string)
	txidSlice, _ := hex.DecodeString(txid)
	var txidArr Uint256
	txidArr.Deserialize(bytes.NewReader(txidSlice[0:32]))
	verbose := params.([]interface{})[1].(bool)
	tx := node.GetTransaction(txidArr)
	// FIXME Get transaction from ledger
	txBuffer := bytes.NewBuffer([]byte{})
	tx.Serialize(txBuffer)
	if verbose == true {
		t := TxInfo{
			Hash: txid,
			Hex:  hex.EncodeToString(txBuffer.Bytes()),
			Tx:   tx,
		}
		response := responsePacking(t, id)
		return response
	}

	return responsePacking(txBuffer.Bytes(), id)
}

func getTxout(req *http.Request, cmd map[string]interface{}) map[string]interface{} {
	id := cmd["id"]
	//params := cmd["params"]
	//txid := params.([]interface{})[0].(string)
	//var n int = params.([]interface{})[1].(int)
	var txout tx.TxOutput // := tx.GetTxOut() //TODO
	high := uint32(txout.Value >> 32)
	low := uint32(txout.Value)
	to := TxoutInfo{
		High:  high,
		Low:   low,
		Txout: txout,
	}
	return responsePacking(to, id)
}

func sendRawTransaction(req *http.Request, cmd map[string]interface{}) map[string]interface{} {
	id := cmd["id"]
	hexValue := cmd["params"].(string)
	hexSlice, _ := hex.DecodeString(hexValue)
	var txTransaction tx.Transaction
	txTransaction.Deserialize(bytes.NewReader(hexSlice[:]))
	err := node.Xmit(&txTransaction)
	return responsePacking(err, id)
}

func submitBlock(req *http.Request, cmd map[string]interface{}) map[string]interface{} {
	id := cmd["id"]
	hexValue := cmd["params"].(string)
	hexSlice, _ := hex.DecodeString(hexValue)
	var txTransaction tx.Transaction
	txTransaction.Deserialize(bytes.NewReader(hexSlice[:]))
	err := node.Xmit(&txTransaction)
	response := responsePacking(err, id)
	return response
}

func getNeighbor(req *http.Request, cmd map[string]interface{}) map[string]interface{} {
	id := cmd["id"]
	addr, _ := node.GetNeighborAddrs()
	return responsePacking(addr, id)
}

func getNodeState(req *http.Request, cmd map[string]interface{}) map[string]interface{} {
	id := cmd["id"]
	n := NodeInfo{
		State:    uint(node.GetState()),
		Time:     node.GetTime(),
		Port:     node.GetPort(),
		ID:       node.GetID(),
		Version:  node.Version(),
		Services: node.Services(),
		Relay:    node.GetRelay(),
		Height:   node.GetHeight(),
	}
	return responsePacking(n, id)
}

func startConsensus(req *http.Request, cmd map[string]interface{}) map[string]interface{} {
	var response map[string]interface{}
	id := cmd["id"]
	err := dBFT.Start()
	if err != nil {
		response = responsePacking("Failed to start", id)
	} else {
		response = responsePacking("Consensus Started", id)
	}
	return response
}

func stopConsensus(req *http.Request, cmd map[string]interface{}) map[string]interface{} {
	var response map[string]interface{}
	id := cmd["id"]
	err := dBFT.Halt()
	if err != nil {
		response = responsePacking("Failed to stop", id)
	} else {
		response = responsePacking("Consensus Stopped", id)
	}
	return response
}

func sendSampleTransaction(req *http.Request, cmd map[string]interface{}) map[string]interface{} {
	var response map[string]interface{}
	id := cmd["id"]
	params, err := base64.StdEncoding.DecodeString(cmd["params"].([]interface{})[0].(string))
	buf := bytes.NewBuffer(params)
	var t tx.Transaction
	if err = t.Deserialize(buf); err != nil {
		response = responsePacking("Unmarshal Sample TX error", id)
		return response
	}

	log.Debug("---------------------------")
	log.Debug("Transaction Hash:", t.Hash())
	for _, v := range t.Programs {
		log.Debug("Transaction Program Code:", v.Code)
		log.Debug("Transaction Program Parameter:", v.Parameter)
	}
	log.Debug("---------------------------")

	if !node.AppendTxnPool(&t) {
		log.Warn("Can NOT add the transaction to TxnPool")
	}
	if err = node.Xmit(&t); err != nil {
		return responsePacking("Xmit Sample TX error", id)
	}
	return responsePacking("Transaction Sended", id)
}
