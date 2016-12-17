package transaction


import (
	"GoOnchain/core/asset"
	"GoOnchain/common"
	"GoOnchain/crypto"
	"GoOnchain/core/transaction/payload"
	"GoOnchain/core/contract/program"
)

//initial a new transaction with asset registration payload
func NewAssetRegistrationTransaction(asset asset.Asset,amount common.Fixed64,precision byte,issuer crypto.PubKey,conroller common.Uint160) (*Transaction, error){

	//TODO: check arguments

	assetRegPayload := &payload.AssetRegistration {
		Asset: asset,
		Amount: amount,
		Precision: precision,
		Issuer: issuer,
		Controller: conroller,
	}

	//TODO: implement NewAssetRegistrationTransaction
	return &Transaction{
		//nonce uint64
		UTXOInputs: []*UTXOTxInput{},
		BalanceInputs: []*BalanceTxInput{},
		Attributes: []*TxAttribute{},
		TxType: RegisterAsset,
		Payload: assetRegPayload,
		Programs: []*program.Program{},
	}, nil
}


