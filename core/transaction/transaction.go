package transaction

import (
	"io"
	"GoOnchain/common/serialization"
	"GoOnchain/core/contract/program"
	"GoOnchain/common"
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

//Payload define the func for loading the payload data
//base on payload type which have different struture
type Payload interface {
	//  Get payload data
	Data() []byte

	//Serialize payload data
	Serialize(w io.Writer)
}

//Transaction is used for carry information or action to Ledger
//validated transaction will be added to block and updates state correspondingly
type Transaction struct {
	TxType TransactionType
	PayloadVersion byte
	Payload Payload
	nonce uint64
	Attributes []*TxAttribute
	UTXOInputs []*UTXOTxInput
	BalanceInputs []*BalanceTxInput
	Outputs []*TxOutput
	Programs []*program.Program


	//Inputs/Outputs map base on Asset (needn't serialize)
	AssetUTXOInputs map[common.Uint256]*UTXOTxInput
	AssetOutputs map[common.Uint256]*TxOutput

	AssetInputAmount map[common.Uint256]*common.Fixed64
	AssetOutputAmount map[common.Uint256]*common.Fixed64

	AssetInputOutputs map[*UTXOTxInput]*TxOutput


}

//Serialize the Transaction
func (tx *Transaction) Serialize(w io.Writer)  {

	tx.SerializeUnsigned(w)

	//Serialize  Transaction's programs
	len := uint64(len(tx.Programs))
	serialization.WriteVarInt(w,len)
	for _, p := range tx.Programs {
		p.Serialize(w)
	}
}

//Serialize the Transaction data without contracts
func (tx *Transaction) SerializeUnsigned(w io.Writer) error  {
	w.Write([]byte{byte(tx.TxType)})
	w.Write([]byte{tx.PayloadVersion})

	tx.Payload.Serialize(w)

	serialization.WriteVarInt(w,tx.nonce)

	serialization.WriteVarInt(w,uint64(len(tx.Attributes))) //TODO: check error
	for _, attr := range tx.Attributes {
		attr.Serialize(w)
	}

	serialization.WriteVarInt(w,uint64(len(tx.UTXOInputs))) //TODO: check error
	for _, utxo := range tx.UTXOInputs {
		utxo.Serialize(w)
	}

	serialization.WriteVarInt(w,uint64(len(tx.BalanceInputs))) //TODO: check error
	for _, balance := range tx.BalanceInputs {
		balance.Serialize(w)
	}

	serialization.WriteVarInt(w,uint64(len(tx.Outputs))) //TODO: check error
	for _, output := range tx.Outputs {
		output.Serialize(w)
	}

	return nil
}


func (tx *Transaction) GetProgramHashes() ([]common.Uint160, error){

	//Set Utxo Inputs' hashes
	programHashes := []common.Uint160{}
	outputHashes,_ := tx.GetOutputHashes() //check error
	programHashes = append(programHashes,outputHashes[:]...)

	return programHashes, nil
}


func (tx *Transaction) SetPrograms(programs []*program.Program){
	tx.Programs = programs
}

func  (tx *Transaction) GetOutputHashes()  ([]common.Uint160, error){
	//TODO: implement Transaction.GetOutputHashes()


	return []common.Uint160{}, nil
}

func  (tx *Transaction) GenerateAssetMaps() {
	//TODO: implement Transaction.GenerateAssetMaps()
}

