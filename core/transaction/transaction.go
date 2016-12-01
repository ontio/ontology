package transaction

import (
	"GoOnchain/core/contract"
)

//for different transaction types with different payload format
//and transaction process methods
type TransactionType byte

const (
	RegisterAsset TransactionType = 0x00
	IssueAsset TransactionType = 0x01
	TransferAsset TransactionType = 0x10
	Record TransactionType =  0x11
)

//Payload define the func for loading the payload data to data
//base on payload type which have different struture
type Payload interface {
	//  Get payload data
	Data() []byte
}

//Transaction is used for carry information or action to Ledger
//validated transaction will be added to block and updates state correspondingly
type Transaction struct {
	nonce uint64
	UTXOInputs []*UTXOTxInput
	BalanceInput []*BalanceTxInput
	Attributes []*TxAttribute
	TxType byte
	Payload Payload
	Constracts []*contract.Contract
}


