package common

import (
	"encoding/hex"
	"encoding/json"
	. "github.com/Ontology/common"
	"github.com/Ontology/core/transaction/payload"
	"github.com/Ontology/core/transaction/utxo"
	Err "github.com/Ontology/net/httprestful/error"
	"time"
)

const AttributeMaxLen = 252

//record
func getRecordData(cmd map[string]interface{}) ([]byte, int64) {
	if raw, ok := cmd["Raw"].(string); ok && raw == "1" {
		str, ok := cmd["RecordData"].(string)
		if !ok {
			return nil, Err.INVALID_PARAMS
		}
		bys, err := HexToBytes(str)
		if err != nil {
			return nil, Err.INVALID_PARAMS
		}
		return bys, Err.SUCCESS
	}
	type Data struct {
		Algrithem string `json:Algrithem`
		Hash      string `json:Hash`
		Signature string `json:Signature`
		Text      string `json:Text`
	}
	type RecordData struct {
		CAkey     string  `json:CAkey`
		Data      Data    `json:Data`
		SeqNo     string  `json:SeqNo`
		Timestamp float64 `json:Timestamp`
	}

	tmp := &RecordData{}
	reqRecordData, ok := cmd["RecordData"].(map[string]interface{})
	if !ok {
		return nil, Err.INVALID_PARAMS
	}
	reqBtys, err := json.Marshal(reqRecordData)
	if err != nil {
		return nil, Err.INVALID_PARAMS
	}

	if err := json.Unmarshal(reqBtys, tmp); err != nil {
		return nil, Err.INVALID_PARAMS
	}
	tmp.CAkey, ok = cmd["CAkey"].(string)
	if !ok {
		return nil, Err.INVALID_PARAMS
	}
	repBtys, err := json.Marshal(tmp)
	if err != nil {
		return nil, Err.INVALID_PARAMS
	}
	return repBtys, Err.SUCCESS
}
func getInnerTimestamp() ([]byte, int64) {
	type InnerTimestamp struct {
		InnerTimestamp float64 `json:InnerTimestamp`
	}
	tmp := &InnerTimestamp{InnerTimestamp: float64(time.Now().Unix())}
	repBtys, err := json.Marshal(tmp)
	if err != nil {
		return nil, Err.INVALID_PARAMS
	}
	return repBtys, Err.SUCCESS
}

/*
func SendRecord(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(Err.SUCCESS)
	var recordData []byte
	var innerTime []byte
	innerTime, resp["Error"] = getInnerTimestamp()
	if innerTime == nil {
		return resp
	}
	recordData, resp["Error"] = getRecordData(cmd)
	if recordData == nil {
		return resp
	}

	var inputs []*utxo.UTXOTxInput
	var outputs []*utxo.TxOutput

	transferTx, _ := tx.NewTransferAssetTransaction(inputs, outputs)

	rcdInner := tx.NewTxAttribute(tx.Description, innerTime)
	transferTx.Attributes = append(transferTx.Attributes, &rcdInner)

	bytesBuf := bytes.NewBuffer(recordData)

	buf := make([]byte, AttributeMaxLen)
	for {
		n, err := bytesBuf.Read(buf)
		if err != nil {
			break
		}
		var data = make([]byte, n)
		copy(data, buf[0:n])
		record := tx.NewTxAttribute(tx.Description, data)
		transferTx.Attributes = append(transferTx.Attributes, &record)
	}
	if errCode := VerifyAndSendTx(transferTx); errCode != ErrNoError {
		resp["Error"] = int64(errCode)
		return resp
	}
	hash := transferTx.Hash()
	resp["Result"] = ToHexString(hash.ToArray())
	return resp
}
*/

/*
func SendRecordTransaction(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(Err.SUCCESS)
	var recordData []byte
	recordData, resp["Error"] = getRecordData(cmd)
	if recordData == nil {
		return resp
	}
	recordType := "record"
	recordTx, _ := tx.NewRecordTransaction(recordType, recordData)

	hash := recordTx.Hash()
	resp["Result"] = ToHexString(hash.ToArray())
	if errCode := VerifyAndSendTx(recordTx); errCode != ErrNoError {
		resp["Error"] = int64(errCode)
		return resp
	}
	return resp
}
*/

func getClaimData(cmd map[string]interface{}) (*payload.Claim, int64) {
	type UTXOTxInput struct {
		ReferTxID          string
		ReferTxOutputIndex uint16
	}
	type Claim struct {
		Claims []*UTXOTxInput
	}
	claim := new(Claim)
	reqClaimData, ok := cmd["data"].(map[string]interface{})
	if !ok {
		return nil, Err.INVALID_PARAMS
	}
	reqBtys, err := json.Marshal(reqClaimData)
	if err != nil {
		return nil, Err.INVALID_PARAMS
	}
	if err := json.Unmarshal(reqBtys, claim); err != nil {
		return nil, Err.INVALID_PARAMS
	}
	if len(claim.Claims) < 1 {
		return nil, Err.INVALID_PARAMS
	}
	realClaim := new(payload.Claim)
	for _, v := range claim.Claims {
		utxoTxinput := new(utxo.UTXOTxInput)
		bytex, err := hex.DecodeString(v.ReferTxID)
		if err != nil {
			return nil, Err.INVALID_PARAMS
		}
		utxoTxinput.ReferTxID, err = Uint256ParseFromBytes(bytex)
		utxoTxinput.ReferTxID, err = Uint256ParseFromBytes(utxoTxinput.ReferTxID.ToArrayReverse())
		if err != nil {
			return nil, Err.INVALID_PARAMS
		}
		utxoTxinput.ReferTxOutputIndex = v.ReferTxOutputIndex
		realClaim.Claims = append(realClaim.Claims, utxoTxinput)
	}
	return realClaim, 0
}

/*
func SendClaim(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(Err.SUCCESS)
	var claimData = new(payload.Claim)
	claimData, resp["Error"] = getClaimData(cmd)
	if claimData == nil {
		return resp
	}
	//TODO: calc txoutput
	//output:=[]*utxo.TxOutput{}
	var controller Uint160
	reference := make(map[*utxo.UTXOTxInput]*utxo.TxOutput)
	if len(claimData.Claims) <= 0 {
		resp["Error"] = int64(ErrTransactionPayload)
		return resp
	}
	// Key index，v UTXOInput
	for _, utxo := range claimData.Claims {
		transaction, err := tx.TxStore.GetTransaction(utxo.ReferTxID)
		if err != nil {
			resp["Error"] = int64(ErrTransactionPayload)
			return resp
		}
		index := utxo.ReferTxOutputIndex
		if len(transaction.Outputs) < int(index) {
			resp["Error"] = int64(ErrTransactionPayload)
			return resp
		}
		reference[utxo] = transaction.Outputs[index]
	}
	for _, output := range reference {
		controller = output.ProgramHash
		break
	}
	amount, err := ledger.CalculateBouns(claimData.Claims, false)
	if err != nil {
		resp["Error"] = int64(ErrTransactionBalance)
		return resp
	}
	output := []*utxo.TxOutput{
		{
			AssetID:     tx.ONGTokenID,
			Value:       amount,
			ProgramHash: controller,
		},
	}
	claimTx, err := tx.NewClaimTransaction(claimData.Claims, output)
	if err != nil {
		resp["Error"] = int64(ErrTransactionPayload)
		return resp
	}
	hash := claimTx.Hash()
	resp["Result"] = ToHexString(hash.ToArrayReverse())
	if errCode := VerifyAndSendTx(claimTx); errCode != ErrNoError {
		resp["Error"] = int64(errCode)
		return resp
	}
	return resp
}
*/
