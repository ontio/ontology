package transaction

import (
	. "GoOnchain/common"
	"GoOnchain/common/serialization"
	"GoOnchain/core/contract/program"
	sig "GoOnchain/core/signature"
	msg "GoOnchain/node/message"
	"GoOnchain/core/transaction/payload"
	. "GoOnchain/errors"
	"crypto/sha256"
	"errors"
	"io"
)

//for different transaction types with different payload format
//and transaction process methods
type TransactionType byte

const (
	BookKeeping   TransactionType = 0x00
	RegisterAsset TransactionType = 0x40
	IssueAsset    TransactionType = 0x01
	TransferAsset TransactionType = 0x10
	Record        TransactionType = 0x11
)

//Payload define the func for loading the payload data
//base on payload type which have different struture
type Payload interface {
	//  Get payload data
	Data() []byte

	//Serialize payload data
	Serialize(w io.Writer)

	Deserialize(r io.Reader) error
}

//Transaction is used for carry information or action to Ledger
//validated transaction will be added to block and updates state correspondingly
type Transaction struct {
	TxType         TransactionType
	PayloadVersion byte
	Payload        Payload
	Nonce          uint64
	Attributes     []*TxAttribute
	UTXOInputs     []*UTXOTxInput
	BalanceInputs  []*BalanceTxInput
	Outputs        []*TxOutput
	Programs       []*program.Program

	//Inputs/Outputs map base on Asset (needn't serialize)
	AssetOutputs      map[Uint256][]*TxOutput
	AssetInputAmount  map[Uint256]Fixed64
	AssetOutputAmount map[Uint256]Fixed64

	hash *Uint256
}

//Serialize the Transaction
func (tx *Transaction) Serialize(w io.Writer) {

	tx.SerializeUnsigned(w)

	//Serialize  Transaction's programs
	len := uint64(len(tx.Programs))
	serialization.WriteVarUint(w, len)
	for _, p := range tx.Programs {
		p.Serialize(w)
	}
}

//Serialize the Transaction data without contracts
func (tx *Transaction) SerializeUnsigned(w io.Writer) error {
	//txType
	w.Write([]byte{byte(tx.TxType)})
	//PayloadVersion
	w.Write([]byte{tx.PayloadVersion})
	//Payload
	if tx.Payload == nil {
		return errors.New("Transaction Payload is nil.")
	}
	tx.Payload.Serialize(w)
	//nonce
	//serialization.WriteVarUint(w, tx.Nonce)
	//[]*txAttribute
	err := serialization.WriteVarUint(w, uint64(len(tx.Attributes)))
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Transaction item txAttribute length serialization failed.")
	}
	for _, attr := range tx.Attributes {
		attr.Serialize(w)
	}
	//[]*UTXOInputs
	err = serialization.WriteVarUint(w, uint64(len(tx.UTXOInputs)))
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Transaction item UTXOInputs length serialization failed.")
	}
	for _, utxo := range tx.UTXOInputs {
		utxo.Serialize(w)
	}
	/*
		//[]*BalanceInputs
		err = serialization.WriteVarUint(w, uint64(len(tx.BalanceInputs)))
		if err != nil {
			return NewDetailErr(err, ErrNoCode, "Transaction item BalanceInputs length serialization failed.")
		}
		for _, balance := range tx.BalanceInputs {
			balance.Serialize(w)
		}
	*/
	//[]*Outputs
	err = serialization.WriteVarUint(w, uint64(len(tx.Outputs)))
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Transaction item Outputs length serialization failed.")
	}
	for _, output := range tx.Outputs {
		output.Serialize(w)
	}

	return nil
}

//deserialize the Transaction
func (tx *Transaction) Deserialize(r io.Reader) error {
	// tx deserialize
	tx.DeserializeUnsigned(r)

	// tx program
	len, _ := serialization.ReadVarUint(r, 0)

	programHashes := []*program.Program{}

	for i := 0; i < int(len); i++ {
		outputHashes := new(program.Program)
		outputHashes.Deserialize(r)
		programHashes = append(programHashes, outputHashes)
	}
	tx.Programs = programHashes

	return nil
}

func (tx *Transaction) DeserializeUnsigned(r io.Reader) error {
	var txType [1]byte
	_, err := io.ReadFull(r, txType[:])
	tx.TxType = TransactionType(txType[0])
	if err != nil {
		return err
	}
	/*
		if txType[0] != byte(tx.TxType) {
			return errors.New("Transaction Type is different.")
		}
	*/
	return tx.DeserializeUnsignedWithoutType(r)
}

func (tx *Transaction) DeserializeUnsignedWithoutType(r io.Reader) error {
	var payloadVersion [1]byte
	_, err := io.ReadFull(r, payloadVersion[:])
	tx.PayloadVersion = payloadVersion[0]
	if err != nil {
		return err
	}

	//payload
	//tx.Payload.Deserialize(r)
	if tx.TxType == RegisterAsset {
		// Asset Registration
		tx.Payload = new(payload.AssetRegistration)
		tx.Payload.Deserialize(r)
	} else if tx.TxType == IssueAsset {
		// Asset Issue
	}
	//	else if tx.TxType == 0x00 {
	//		// Miner
	//		tx.Payload = new(payload.Miner)
	//		tx.Payload.Deserialize(r)
	//	}
	//attributes
	Len, err := serialization.ReadVarUint(r, 0)
	if err != nil {
		return err
	}
	for i := uint64(0); i < Len; i++ {
		attr := new(TxAttribute)
		err = attr.Deserialize(r)
		if err != nil {
			return err
		}
		tx.Attributes = append(tx.Attributes, attr)
	}

	//UTXOInputs
	Len, err = serialization.ReadVarUint(r, 0)
	if err != nil {
		return err
	}
	for i := uint64(0); i < Len; i++ {
		utxo := new(UTXOTxInput)
		err = utxo.Deserialize(r)
		if err != nil {
			return err
		}
		tx.UTXOInputs = append(tx.UTXOInputs, utxo)
	}
	/*
		//balanceInputs
		Len, err = serialization.ReadVarUint(r, 0)
		if err != nil {
			return err
		}
		for i := uint64(0); i < Len; i++ {
			balanceInput := new(BalanceTxInput)
			err = balanceInput.Deserialize(r)
			if err != nil {
				return err
			}
			tx.BalanceInputs = append(tx.BalanceInputs, balanceInput)
		}
	*/
	//Outputs
	Len, err = serialization.ReadVarUint(r, 0)
	if err != nil {
		return err
	}
	for i := uint64(0); i < Len; i++ {
		output := new(TxOutput)
		output.Deserialize(r)

		tx.Outputs = append(tx.Outputs, output)
	}
	return nil
}

func (tx *Transaction) GetProgramHashes() ([]Uint160, error) {

	//Set Utxo Inputs' hashes
	programHashes := []Uint160{}
	outputHashes, _ := tx.GetOutputHashes() //check error
	programHashes = append(programHashes, outputHashes[:]...)

	return programHashes, nil
}

func (tx *Transaction) SetPrograms(programs []*program.Program) {
	tx.Programs = programs
}

func (tx *Transaction) GetPrograms() []*program.Program {
	return tx.Programs
}

func (tx *Transaction) GetOutputHashes() ([]Uint160, error) {
	//TODO: implement Transaction.GetOutputHashes()

	return []Uint160{}, nil
}

func (tx *Transaction) GenerateAssetMaps() {
	//TODO: implement Transaction.GenerateAssetMaps()
}

func (tx *Transaction) GetMessage() []byte {
	return sig.GetHashData(tx)
}

func (tx *Transaction) Hash() Uint256 {
	if tx.hash == nil {
		d := sig.GetHashData(tx)
		temp := sha256.Sum256([]byte(d))
		f := Uint256(sha256.Sum256(temp[:]))
		tx.hash = &f
	}
	return *tx.hash

}

func (tx *Transaction) SetHash(hash Uint256) {
	tx.hash = &hash
}

func (tx *Transaction) Type() InventoryType{
	return TRANSACTION
}
func (tx *Transaction) Verify() error {
	//TODO: Verify()
	return nil
}
