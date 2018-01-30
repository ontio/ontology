package transaction

import (
	"github.com/Ontology/common"
	"github.com/Ontology/core/asset"
	"github.com/Ontology/core/contract/program"
	"github.com/Ontology/core/transaction/payload"
	"github.com/Ontology/crypto"
	. "github.com/Ontology/core/transaction/utxo"
	"github.com/Ontology/core/code"
	"github.com/Ontology/smartcontract/types"
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
func NewBookKeeperTransaction(pubKey *crypto.PubKey, isAdd bool, cert []byte, issuer *crypto.PubKey) (*Transaction, error) {

	bookKeeperPayload := &payload.BookKeeper{
		PubKey: pubKey,
		Action: payload.BookKeeperAction_SUB,
		Cert:   cert,
		Issuer: issuer,
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
func NewDataFileTransaction(path string, fileName string, note string, issuer *crypto.PubKey) (*Transaction, error) {
	//TODO: check arguments
	DataFilePayload := &payload.DataFile{
		IPFSPath: path,
		Filename: fileName,
		Note:     note,
		Issuer:   issuer,
	}

	return &Transaction{
		TxType:        DataFile,
		Payload:       DataFilePayload,
		Attributes:    []*TxAttribute{},
		UTXOInputs:    []*UTXOTxInput{},
		BalanceInputs: []*BalanceTxInput{},
		Programs:      []*program.Program{},
	}, nil
}

func NewClaimTransaction(claims []*UTXOTxInput, output []*TxOutput) (*Transaction, error) {
	ClaimPayload := &payload.Claim{
		Claims: claims,
	}
	return &Transaction{
		TxType:        Claim,
		Payload:       ClaimPayload,
		Attributes:    []*TxAttribute{},
		UTXOInputs:    []*UTXOTxInput{},
		BalanceInputs: []*BalanceTxInput{},
		Programs:      []*program.Program{},
		Outputs:       output,
	}, nil
}

func NewDeployTransaction(fc *code.FunctionCode, programHash common.Uint160, name, codeversion, author, email, desp string, vmType types.VmType, needStorage bool) (*Transaction, error) {
	//TODO: check arguments
	DeployCodePayload := &payload.DeployCode{
		Code:        fc,
		NeedStorage: needStorage,
		Name:        name,
		CodeVersion: codeversion,
		Author:      author,
		Email:       email,
		Description: desp,
	}

	return &Transaction{
		TxType:        Deploy,
		Payload:       DeployCodePayload,
		Attributes:    []*TxAttribute{},
		UTXOInputs:    []*UTXOTxInput{},
		BalanceInputs: []*BalanceTxInput{},
		Programs:      []*program.Program{},
	}, nil
}

func NewInvokeTransaction(fc []byte, codeHash common.Uint160) (*Transaction, error) {
	//TODO: check arguments
	InvokeCodePayload := &payload.InvokeCode{
		Code:     fc,
		CodeHash: codeHash,
	}

	return &Transaction{
		TxType:        Invoke,
		Payload:       InvokeCodePayload,
		Attributes:    []*TxAttribute{},
		UTXOInputs:    []*UTXOTxInput{},
		BalanceInputs: []*BalanceTxInput{},
		Programs:      []*program.Program{},
	}, nil
}

