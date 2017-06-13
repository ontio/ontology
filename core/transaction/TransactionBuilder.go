package transaction

import (
	"DNA/common"
	"DNA/core/asset"
	"DNA/core/contract/program"
	"DNA/core/transaction/payload"
	"DNA/crypto"
)

//initial a new transaction with asset registration payload
func NewRegisterAssetTransaction(asset *asset.Asset, amount common.Fixed64, issuer *crypto.PubKey, conroller common.Uint160) (*Transaction, error) {

	//TODO: check arguments

	assetRegPayload := &payload.RegisterAsset{
		Asset:  asset,
		Amount: amount,
		//Precision: precision,
		Issuer:     issuer,
		Controller: conroller,
	}

	return &Transaction{
		//nonce uint64 //TODO: genenrate nonce
		UTXOInputs:    []*UTXOTxInput{},
		BalanceInputs: []*BalanceTxInput{},
		Attributes:    []*TxAttribute{},
		TxType:        RegisterAsset,
		Payload:       assetRegPayload,
		Programs:      []*program.Program{},
	}, nil
}

//initial a new transaction with asset registration payload
func NewBookKeeperTransaction(pubKey *crypto.PubKey, isAdd bool, cert []byte) (*Transaction, error) {

	bookKeeperPayload := &payload.BookKeeper{
		PubKey: pubKey,
		Action: payload.BookKeeperAction_SUB,
		Cert:   cert,
	}

	if isAdd {
		bookKeeperPayload.Action = payload.BookKeeperAction_ADD
	}

	return &Transaction{
		TxType:        BookKeeper,
		Payload:       bookKeeperPayload,
		UTXOInputs:    []*UTXOTxInput{},
		BalanceInputs: []*BalanceTxInput{},
		Attributes:    []*TxAttribute{},
		Programs:      []*program.Program{},
	}, nil
}

func NewIssueAssetTransaction(outputs []*TxOutput) (*Transaction, error) {

	assetRegPayload := &payload.IssueAsset{}

	return &Transaction{
		TxType:        IssueAsset,
		Payload:       assetRegPayload,
		Attributes:    []*TxAttribute{},
		BalanceInputs: []*BalanceTxInput{},
		Outputs:       outputs,
		Programs:      []*program.Program{},
	}, nil
}

func NewTransferAssetTransaction(inputs []*UTXOTxInput, outputs []*TxOutput) (*Transaction, error) {

	//TODO: check arguments

	assetRegPayload := &payload.TransferAsset{}

	return &Transaction{
		TxType:        TransferAsset,
		Payload:       assetRegPayload,
		Attributes:    []*TxAttribute{},
		UTXOInputs:    inputs,
		BalanceInputs: []*BalanceTxInput{},
		Outputs:       outputs,
		Programs:      []*program.Program{},
	}, nil
}

//initial a new transaction with record payload
func NewRecordTransaction(recordType string, recordData []byte) (*Transaction, error) {
	//TODO: check arguments
	recordPayload := &payload.Record{
		RecordType: recordType,
		RecordData: recordData,
	}

	return &Transaction{
		TxType:        Record,
		Payload:       recordPayload,
		Attributes:    []*TxAttribute{},
		UTXOInputs:    []*UTXOTxInput{},
		BalanceInputs: []*BalanceTxInput{},
		Programs:      []*program.Program{},
	}, nil
}

func NewPrivacyPayloadTransaction(fromPrivKey []byte, fromPubkey *crypto.PubKey, toPubkey *crypto.PubKey, payloadType payload.EncryptedPayloadType, data []byte) (*Transaction, error) {
	privacyPayload := &payload.PrivacyPayload{
		PayloadType: payloadType,
		EncryptType: payload.ECDH_AES256,
		EncryptAttr: &payload.EcdhAes256{
			FromPubkey: fromPubkey,
			ToPubkey:   toPubkey,
		},
	}
	privacyPayload.Payload, _ = privacyPayload.EncryptAttr.Encrypt(data, fromPrivKey)

	return &Transaction{
		TxType:        PrivacyPayload,
		Payload:       privacyPayload,
		Attributes:    []*TxAttribute{},
		UTXOInputs:    []*UTXOTxInput{},
		BalanceInputs: []*BalanceTxInput{},
		Programs:      []*program.Program{},
	}, nil
}
