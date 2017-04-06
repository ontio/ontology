package transaction


import (
	"DNA/core/asset"
	"DNA/common"
	"DNA/crypto"
	"DNA/core/transaction/payload"
	"DNA/core/contract/program"
)

//initial a new transaction with asset registration payload
func NewRegisterAssetTransaction(asset *asset.Asset,amount common.Fixed64,issuer *crypto.PubKey,conroller common.Uint160) (*Transaction, error){

	//TODO: check arguments

	assetRegPayload := &payload.RegisterAsset {
		Asset: asset,
		Amount: amount,
		//Precision: precision,
		Issuer: issuer,
		Controller: conroller,
	}

    return &Transaction{
        //nonce uint64 //TODO: genenrate nonce
        UTXOInputs: []*UTXOTxInput{},
        BalanceInputs: []*BalanceTxInput{},
        Attributes: []*TxAttribute{},
        TxType: RegisterAsset,
        Payload: assetRegPayload,
        Programs: []*program.Program{},
    }, nil
}

func NewIssueAssetTransaction() (*Transaction, error){

    assetRegPayload := &payload.IssueAsset {
    }

    return &Transaction{
        UTXOInputs: []*UTXOTxInput{},
        BalanceInputs: []*BalanceTxInput{},
        Attributes: []*TxAttribute{},
        TxType: IssueAsset,
        Payload: assetRegPayload,
        Programs: []*program.Program{},
    }, nil
}

func NewTransferAssetTransaction() (*Transaction, error){

    //TODO: check arguments

    assetRegPayload := &payload.TransferAsset {
    }

    return &Transaction{
        UTXOInputs: []*UTXOTxInput{},
        BalanceInputs: []*BalanceTxInput{},
        Attributes: []*TxAttribute{},
        TxType: TransferAsset,
        Payload: assetRegPayload,
        Programs: []*program.Program{},
    }, nil
}