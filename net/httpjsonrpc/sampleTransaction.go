package httpjsonrpc

import (
	"DNA/client"
	. "DNA/common"
	"DNA/common/log"
	. "DNA/core/asset"
	"DNA/core/contract"
	"DNA/core/signature"
	"DNA/core/transaction"
	"strconv"
)

const (
	ASSETPREFIX = "DNA"
)

func NewRegTx(rand string, index int, admin, issuer *client.Account) *transaction.Transaction {
	name := ASSETPREFIX + "-" + strconv.Itoa(index) + "-" + rand
	asset := &Asset{name, byte(0x00), AssetType(Share), UTXO}
	amount := Fixed64(1000)
	controller, _ := contract.CreateSignatureContract(admin.PubKey())
	tx, _ := transaction.NewRegisterAssetTransaction(asset, amount, issuer.PubKey(), controller.ProgramHash)
	return tx
}

func NewIssueTx(admin *client.Account, hash Uint256) *transaction.Transaction {
	tx, _ := transaction.NewIssueAssetTransaction()
	temp, err := admin.PublicKey.EncodePoint(true)
	if err != nil {
		log.Error("EncodePoint error.")
	}
	hashx, err := ToCodeHash(temp)
	if err != nil {
		log.Error("TocodeHash hash error.")
	}
	issueTxOutput := &transaction.TxOutput{
		AssetID:     hash,
		Value:       Fixed64(100),
		ProgramHash: hashx,
	}
	tx.Outputs = append(tx.Outputs, issueTxOutput)
	return tx
}

func NewTransferTx(regHash, issueHash Uint256, toUser *client.Account) *transaction.Transaction {
	tx, _ := transaction.NewTransferAssetTransaction()
	// TODO: fill the UTXO inputs after TX inputs verification works
	/*
		transferUTXOInput := &transaction.UTXOTxInput{
			ReferTxID:          issueHash,
			ReferTxOutputIndex: uint16(0),
		}
		tx.UTXOInputs = append(tx.UTXOInputs, transferUTXOInput)
	*/
	temp, err := toUser.PublicKey.EncodePoint(true)
	if err != nil {
		log.Error("EncodePoint error.")
	}
	hashx, err := ToCodeHash(temp)
	if err != nil {
		log.Error("TocodeHash hash error.")
	}
	transferTxOutput := &transaction.TxOutput{
		AssetID:     regHash,
		Value:       Fixed64(100),
		ProgramHash: hashx,
	}
	tx.Outputs = append(tx.Outputs, transferTxOutput)
	return tx
}

func NewRecordTx(rand string) *transaction.Transaction {
	recordType := string("txt")
	recordData := []byte("hello world " + rand)
	tx, _ := transaction.NewRecordTransaction(recordType, recordData)
	return tx
}

func SignTx(admin *client.Account, tx *transaction.Transaction) {
	signdate, err := signature.SignBySigner(tx, admin)
	if err != nil {
		log.Error(err, "signdate SignBySigner failed")
	}
	transactionContract, _ := contract.CreateSignatureContract(admin.PublicKey)
	transactionContractContext := contract.NewContractContext(tx)
	transactionContractContext.AddContract(transactionContract, admin.PublicKey, signdate)
	tx.SetPrograms(transactionContractContext.GetPrograms())
}

func SendTx(tx *transaction.Transaction) {
	if !node.AppendTxnPool(tx) {
		log.Warn("Can NOT add the transaction to TxnPool")
	}
	if err := node.Xmit(tx); err != nil {
		log.Error("Xmit Tx Error")
	}
}
