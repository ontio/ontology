package httpjsonrpc

import (
	"DNA/client"
	. "DNA/common"
	"DNA/common/log"
	"DNA/core/ledger"
	tx "DNA/core/transaction"
	"bytes"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"
)

const (
	RANDBYTELEN = 4
)

func TransArryByteToHexString(ptx *tx.Transaction) *Transactions {

	trans := new(Transactions)
	trans.TxType = ptx.TxType
	trans.PayloadVersion = ptx.PayloadVersion
	trans.Payload = TransPayloadToHex(ptx.Payload)
	trans.Nonce = ptx.Nonce

	n := 0
	trans.Attributes = make([]TxAttributeInfo, len(ptx.Attributes))
	for _, v := range ptx.Attributes {
		trans.Attributes[n].Usage = v.Usage
		trans.Attributes[n].Date = ToHexString(v.Date)
		trans.Attributes[n].Size = v.Size
		n++
	}

	n = 0
	trans.UTXOInputs = make([]UTXOTxInputInfo, len(ptx.UTXOInputs))
	for _, v := range ptx.UTXOInputs {
		trans.UTXOInputs[n].ReferTxID = ToHexString(v.ReferTxID.ToArray())
		trans.UTXOInputs[n].ReferTxOutputIndex = v.ReferTxOutputIndex
		n++
	}

	n = 0
	trans.BalanceInputs = make([]BalanceTxInputInfo, len(ptx.BalanceInputs))
	for _, v := range ptx.BalanceInputs {
		trans.BalanceInputs[n].AssetID = ToHexString(v.AssetID.ToArray())
		trans.BalanceInputs[n].Value = v.Value
		trans.BalanceInputs[n].ProgramHash = ToHexString(v.ProgramHash.ToArray())
		n++
	}

	n = 0
	trans.Outputs = make([]TxoutputInfo, len(ptx.Outputs))
	for _, v := range ptx.Outputs {
		trans.Outputs[n].AssetID = ToHexString(v.AssetID.ToArray())
		trans.Outputs[n].Value = v.Value
		trans.Outputs[n].ProgramHash = ToHexString(v.ProgramHash.ToArray())
		n++
	}

	n = 0
	trans.Programs = make([]ProgramInfo, len(ptx.Programs))
	for _, v := range ptx.Programs {
		trans.Programs[n].Code = ToHexString(v.Code)
		trans.Programs[n].Parameter = ToHexString(v.Parameter)
		n++
	}

	n = 0
	trans.AssetOutputs = make([]TxoutputMap, len(ptx.AssetOutputs))
	for k, v := range ptx.AssetOutputs {
		trans.AssetOutputs[n].Key = k
		trans.AssetOutputs[n].Txout = make([]TxoutputInfo, len(v))
		for m := 0; m < len(v); m++ {
			trans.AssetOutputs[n].Txout[m].AssetID = ToHexString(v[m].AssetID.ToArray())
			trans.AssetOutputs[n].Txout[m].Value = v[m].Value
			trans.AssetOutputs[n].Txout[m].ProgramHash = ToHexString(v[m].ProgramHash.ToArray())
		}
		n += 1
	}

	n = 0
	trans.AssetInputAmount = make([]AmountMap, len(ptx.AssetInputAmount))
	for k, v := range ptx.AssetInputAmount {
		trans.AssetInputAmount[n].Key = k
		trans.AssetInputAmount[n].Value = v
		n += 1
	}

	n = 0
	trans.AssetOutputAmount = make([]AmountMap, len(ptx.AssetOutputAmount))
	for k, v := range ptx.AssetOutputAmount {
		trans.AssetInputAmount[n].Key = k
		trans.AssetInputAmount[n].Value = v
		n += 1
	}

	mhash := ptx.Hash()
	trans.Hash = ToHexString(mhash.ToArray())

	return trans
}

func getBestBlockHash(cmd map[string]interface{}) map[string]interface{} {
	id := cmd["id"]
	hash := ledger.DefaultLedger.Blockchain.CurrentBlockHash()
	response := responsePacking(ToHexString(hash.ToArray()), id)
	return response
}

func getBlock(cmd map[string]interface{}) map[string]interface{} {
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

	blockHead := &BlockHead{
		Version:          block.Blockdata.Version,
		PrevBlockHash:    ToHexString(block.Blockdata.PrevBlockHash.ToArray()),
		TransactionsRoot: ToHexString(block.Blockdata.TransactionsRoot.ToArray()),
		Timestamp:        block.Blockdata.Timestamp,
		Height:           block.Blockdata.Height,
		ConsensusData:    block.Blockdata.ConsensusData,
		NextMiner:        ToHexString(block.Blockdata.NextMiner.ToArray()),
		Program: ProgramInfo{
			Code:      ToHexString(block.Blockdata.Program.Code),
			Parameter: ToHexString(block.Blockdata.Program.Parameter),
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
	return responsePacking(b, id)
}

func getBlockCount(cmd map[string]interface{}) map[string]interface{} {
	id := cmd["id"]
	count := ledger.DefaultLedger.Blockchain.BlockHeight + 1
	return responsePacking(count, id)
}

func getBlockHash(cmd map[string]interface{}) map[string]interface{} {
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

func getTxn(cmd map[string]interface{}) map[string]interface{} {
	id := cmd["id"]
	params := cmd["params"]
	var hash Uint256

	txid := params.([]interface{})[0].(string)
	hashslice, _ := hex.DecodeString(txid)
	hash.Deserialize(bytes.NewReader(hashslice[0:32]))

	tx, err := ledger.DefaultLedger.Store.GetTransaction(hash)
	if err != nil {
		return responsePacking([]interface{}{-100, "Unknown Transaction Hash"}, id)
	}

	tran := TransArryByteToHexString(tx)
	return responsePacking(tran, id)
}

func getAddrTxn(cmd map[string]interface{}) map[string]interface{} {
	return nil
}

func getConnectionCount(cmd map[string]interface{}) map[string]interface{} {
	id := cmd["id"]
	count := node.GetConnectionCnt()
	return responsePacking(count, id)
}

func getRawMemPool(cmd map[string]interface{}) map[string]interface{} {
	id := cmd["id"]
	mempoollist := node.GetTxnPool(false)
	return responsePacking(mempoollist, id)
}

func getRawTransaction(cmd map[string]interface{}) map[string]interface{} {
	id := cmd["id"]
	params := cmd["params"]
	var hash Uint256

	txid := params.([]interface{})[0].(string)
	verbose := params.([]interface{})[1].(bool)
	hashslice, _ := hex.DecodeString(txid)
	hash.Deserialize(bytes.NewReader(hashslice[0:32]))

	tx, err := ledger.DefaultLedger.Store.GetTransaction(hash)
	if err != nil {
		return responsePacking([]interface{}{-100, "Unknown Transaction Hash"}, id)
	}

	tran := TransArryByteToHexString(tx)
	txBuffer := bytes.NewBuffer([]byte{})
	tx.Serialize(txBuffer)

	if verbose == true {
		t := TxInfo{
			Hash: txid,
			Hex:  ToHexString(hash.ToArray()),
			Tx:   tran,
		}
		response := responsePacking(t, id)
		return response
	}
	return responsePacking(ToHexString(txBuffer.Bytes()), id)
}

func getTxout(cmd map[string]interface{}) map[string]interface{} {
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

func sendRawTransaction(cmd map[string]interface{}) map[string]interface{} {
	retFlag := false
	id := cmd["id"]
	params := cmd["params"]
	//hexValue := cmd["params"].(string)
	hexValue := params.([]interface{})[0].(string)

	hexSlice, _ := hex.DecodeString(hexValue)
	var txTransaction tx.Transaction
	txTransaction.Deserialize(bytes.NewReader(hexSlice[:]))
	err := node.Xmit(&txTransaction)
	if err == nil {
		retFlag = true
	}
	return responsePacking(retFlag, id)
}

func submitBlock(cmd map[string]interface{}) map[string]interface{} {
	id := cmd["id"]
	hexValue := cmd["params"].(string)
	hexSlice, _ := hex.DecodeString(hexValue)
	var txTransaction tx.Transaction
	txTransaction.Deserialize(bytes.NewReader(hexSlice[:]))
	err := node.Xmit(&txTransaction)
	response := responsePacking(err, id)
	return response
}

func getNeighbor(cmd map[string]interface{}) map[string]interface{} {
	id := cmd["id"]
	addr, _ := node.GetNeighborAddrs()
	return responsePacking(addr, id)
}

func getNodeState(cmd map[string]interface{}) map[string]interface{} {
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
		TxnCnt:   node.GetTxnCnt(),
		RxTxnCnt: node.GetRxTxnCnt(),
	}
	return responsePacking(n, id)
}

func startConsensus(cmd map[string]interface{}) map[string]interface{} {
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

func stopConsensus(cmd map[string]interface{}) map[string]interface{} {
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

func sendSampleTransaction(cmd map[string]interface{}) map[string]interface{} {
	id := cmd["id"]
	txType := cmd["params"].([]interface{})[0].(string)
	issuer, err := client.NewAccount()
	if err != nil {
		return responsePacking("Failed to create account", id)
	}
	admin := issuer

	var regHash, issueHash, transferHash, recordHash Uint256
	rbuf := make([]byte, RANDBYTELEN)
	rand.Read(rbuf)
	switch string(txType) {
	case "perf":
		txNum := cmd["params"].([]interface{})[1].(float64)
		nosign := cmd["params"].([]interface{})[2].(bool)
		num := int(txNum)
		for i := 0; i < num; i++ {
			regTx := NewRegTx(ToHexString(rbuf), i, admin, issuer)
			regHash = regTx.Hash()
			if !nosign {
				SignTx(admin, regTx)
			}
			SendTx(regTx)
		}
		return responsePacking(fmt.Sprintf("%d transactions was sended", num), id)
	case "full":
		regTx := NewRegTx(ToHexString(rbuf), 0, admin, issuer)
		regHash = regTx.Hash()
		SignTx(admin, regTx)
		SendTx(regTx)

		// wait for the block
		time.Sleep(5 * time.Second)
		issueTx := NewIssueTx(admin, regHash)
		issueHash = issueTx.Hash()
		SignTx(admin, issueTx)
		SendTx(issueTx)

		// wait for the block
		time.Sleep(5 * time.Second)
		transferTx := NewTransferTx(regHash, issueHash, issuer)
		transferHash = transferTx.Hash()
		SignTx(admin, transferTx)
		SendTx(transferTx)

		// wait for the block
		time.Sleep(5 * time.Second)
		NewRecordTx := NewRecordTx(ToHexString(rbuf))
		recordHash = NewRecordTx.Hash()
		SignTx(admin, NewRecordTx)
		SendTx(NewRecordTx)

		return responsePacking(fmt.Sprintf("regist: %x, issue: %x, transfer: %x, record: %x", regHash, issueHash, transferHash, recordHash), id)
	default:
		return responsePacking("Invalid transacion type", id)
	}
}

func setDebugInfo(cmd map[string]interface{}) map[string]interface{} {
	id := cmd["id"]
	param := cmd["params"].([]interface{})[0].(float64)
	level := int(param)
	err := log.Log.SetDebugLevel(level)
	if err != nil {
		return responsePacking("Invaild Debug Level", id)
	}
	return responsePacking("debug level is set successfully", id)
}
