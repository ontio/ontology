package transaction

import (
	. "DNA/common"
	"DNA/common/serialization"
	"DNA/core/contract"
	"DNA/core/contract/program"
	sig "DNA/core/signature"
	"DNA/core/transaction/payload"
	. "DNA/errors"
	"crypto/sha256"
	"errors"
	"io"
	"sort"
	"fmt"
)

//for different transaction types with different payload format
//and transaction process methods
type TransactionType byte

const (
	BookKeeping   TransactionType = 0x00
	RegisterAsset TransactionType = 0x40
	IssueAsset    TransactionType = 0x01
	TransferAsset TransactionType = 0x10
	Record         TransactionType = 0x11
	DeployCode	TransactionType = 0xd0
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

var TxStore ILedgerStore

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
func (tx *Transaction) Serialize(w io.Writer) error{

	err :=tx.SerializeUnsigned(w)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Transaction txSerializeUnsigned Serialize failed.")
	}
	//Serialize  Transaction's programs
	lens := uint64(len(tx.Programs))
	err =serialization.WriteVarUint(w, lens)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Transaction WriteVarUint failed.")
	}
	if lens >0 {
		for _, p := range tx.Programs {
			err = p.Serialize(w)
			if err != nil {
				return NewDetailErr(err, ErrNoCode, "Transaction Programs Serialize failed.")
			}
		}
	}
	return nil
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
	serialization.WriteVarUint(w, tx.Nonce)
	//[]*txAttribute
	err := serialization.WriteVarUint(w, uint64(len(tx.Attributes)))
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Transaction item txAttribute length serialization failed.")
	}
	if len(tx.Attributes) > 0 {
		for _, attr := range tx.Attributes {
			attr.Serialize(w)
		}
	}
	//[]*UTXOInputs
	err = serialization.WriteVarUint(w, uint64(len(tx.UTXOInputs)))
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Transaction item UTXOInputs length serialization failed.")
	}
	if len(tx.UTXOInputs) > 0 {
		for _, utxo := range tx.UTXOInputs {
			utxo.Serialize(w)
		}
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
	if len(tx.Outputs) > 0 {
		for _, output := range tx.Outputs {
			output.Serialize(w)
		}
	}

	return nil
}

//deserialize the Transaction
func (tx *Transaction) Deserialize(r io.Reader) error {
	// tx deserialize
	tx.DeserializeUnsigned(r)

	// tx program
	lens, _ := serialization.ReadVarUint(r, 0)

	programHashes := []*program.Program{}
	if lens>0 {
		for i := 0; i < int(lens); i++ {
			outputHashes := new(program.Program)
			outputHashes.Deserialize(r)
			programHashes = append(programHashes, outputHashes)
		}
		tx.Programs = programHashes
	}
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
		tx.Payload = new(payload.RegisterAsset)
	} else if tx.TxType == IssueAsset {
		// Issue Asset
		tx.Payload = new(payload.IssueAsset)
	} else if tx.TxType == TransferAsset {
		// Transfer Asset
		tx.Payload = new(payload.TransferAsset)
	}else if tx.TxType == BookKeeping{
		tx.Payload = new(payload.BookKeeping)
	}
	tx.Payload.Deserialize(r)
	//attributes

	nonce, err := serialization.ReadVarUint(r, 0)
	if err != nil {
		return errors.New("Parse nonce error")
	}
	tx.Nonce = nonce

	Len, err := serialization.ReadVarUint(r, 0)
	if err != nil {
		return err
	}
	if Len > uint64(0) {
		for i := uint64(0); i < Len; i++ {
			attr := new(TxAttribute)
			err = attr.Deserialize(r)
			if err != nil {
				return err
			}
			tx.Attributes = append(tx.Attributes, attr)
		}
	}
	//UTXOInputs
	Len, err = serialization.ReadVarUint(r, 0)
	if err != nil {
		return err
	}
	if Len > uint64(0) {
		for i := uint64(0); i < Len; i++ {
			utxo := new(UTXOTxInput)
			err = utxo.Deserialize(r)
			if err != nil {
				return err
			}
			tx.UTXOInputs = append(tx.UTXOInputs, utxo)
		}
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
	if Len > uint64(0) {
		for i := uint64(0); i < Len; i++ {
			output := new(TxOutput)
			output.Deserialize(r)

			tx.Outputs = append(tx.Outputs, output)
		}
	}
	return nil
}

func (tx *Transaction) GetProgramHashes() ([]Uint160, error) {
	if tx == nil {
		return []Uint160{}, errors.New("[Transaction],GetProgramHashes transaction is nil.")
	}
	hashs := []Uint160{}
	// add inputUTXO's transaction
	referenceWithUTXO_Output, err := tx.GetReference()
	if err != nil {
		return nil, NewDetailErr(err, ErrNoCode, "[Transaction], GetProgramHashes failed.")
	}
	for _, output := range referenceWithUTXO_Output {
		programHash := output.ProgramHash
		hashs = append(hashs, programHash)
	}
	for _, attribute := range tx.Attributes {
		if attribute.Usage == Script {
			dataHash, err := Uint160ParseFromBytes(attribute.Date)
			if err != nil {
				return nil, NewDetailErr(errors.New("[Transaction], GetProgramHashes err."), ErrNoCode, "")
			}
			hashs = append(hashs, Uint160(dataHash))
		}
	}
	switch tx.TxType {
	case RegisterAsset:
		issuer := tx.Payload.(*payload.RegisterAsset).Issuer
		signatureRedeemScript, err := contract.CreateSignatureRedeemScript(issuer)
		if err != nil {
			return nil, NewDetailErr(err, ErrNoCode, "[Transaction], GetProgramHashes CreateSignatureRedeemScript failed.")
		}

		astHash, err := ToCodeHash(signatureRedeemScript)
		if err != nil {
			return nil, NewDetailErr(err, ErrNoCode, "[Transaction], GetProgramHashes ToCodeHash failed.")
		}
		hashs = append(hashs, astHash)
	case IssueAsset:
		result, err := tx.GetTransactionResults()
		if err != nil {
			return nil, NewDetailErr(err, ErrNoCode, "[Transaction], GetTransactionResults failed.")
		}
		for _, v := range result {
			tx,err := TxStore.GetTransaction(v.AssetId)
			if err != nil {
				return nil, NewDetailErr(err, ErrNoCode, fmt.Sprintf("[Transaction], GetTransaction failed With AssetID:=%x",v.AssetId))
			}
			if tx.TxType != RegisterAsset{
				return nil, NewDetailErr(err, ErrNoCode, fmt.Sprintf("[Transaction], Transaction Type ileage With AssetID:=%x",v.AssetId))
			}

			switch v1 := tx.Payload.(type){
				case *payload.RegisterAsset:
					hashs = append(hashs,v1.Controller)
				default:
					return nil, NewDetailErr(err, ErrNoCode, fmt.Sprintf("[Transaction], payload is illegal",v.AssetId))
			}
		}

	case TransferAsset:
	default:
	}
	sort.Sort(byProgramHashes(hashs))
	return hashs, nil
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
	return sig.GetHashForSigning(tx)
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

func (tx *Transaction) Type() InventoryType {
	return TRANSACTION
}
func (tx *Transaction) Verify() error {
	//TODO: Verify()
	return nil
}

func (tx *Transaction) GetReference() (map[*UTXOTxInput]*TxOutput, error) {
	if tx.TxType == RegisterAsset {
		return nil, nil
	}
	//UTXO input /  Outputs
	reference := make(map[*UTXOTxInput]*TxOutput)
	// Key index，v UTXOInput
	for _, utxo := range tx.UTXOInputs {
		transaction, err := TxStore.GetTransaction(utxo.ReferTxID)
		if err != nil {
			return nil, NewDetailErr(err, ErrNoCode, "[Transaction], GetReference failed.")
		}
		index := utxo.ReferTxOutputIndex
		reference[utxo] = transaction.Outputs[index]
	}
	return reference, nil
}
func (tx *Transaction) GetTransactionResults() ([]*TransactionResult, error) {
	reference, err := tx.GetReference()
	if err != nil {
		return nil, err
	}
	result := []*TransactionResult{}
	var finded bool
	for _, o := range tx.Outputs {
		finded = false
		res := new(TransactionResult)
		for _, r := range reference {
			if r.AssetID == o.AssetID {
				finded = true
				res.AssetId = r.AssetID
				res.Amount = r.Value - o.Value
			}
		}
		if finded == false {
			res.AssetId = o.AssetID
			res.Amount = o.Value * Fixed64(-1)
		}
		result = append(result, res)
	}
	return result, nil
}

type byProgramHashes []Uint160

func (a byProgramHashes) Len() int      { return len(a) }
func (a byProgramHashes) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byProgramHashes) Less(i, j int) bool {
	if a[i].CompareTo(a[j]) > 0 {
		return false
	} else {
		return true
	}
}
