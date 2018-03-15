package types

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"io"

	"fmt"
	. "github.com/Ontology/common"
	. "github.com/Ontology/common/config"
	"github.com/Ontology/common/serialization"
	"github.com/Ontology/core/payload"
	"github.com/Ontology/crypto"
	. "github.com/Ontology/errors"
)

type Transaction struct {
	Version    byte
	TxType     TransactionType
	Nonce      uint32
	Payload    Payload
	Attributes []*TxAttribute
	Fee        []*Fee
	NetWorkFee Fixed64
	Sigs       []*Sig

	hash *Uint256
}

type Sig struct {
	PubKeys []*crypto.PubKey
	M       uint8
	SigData [][]byte
}

func (self *Sig) Deserialize(r io.Reader) error {
	n, err := serialization.ReadVarUint(r, 0)
	if err != nil {
		return err
	}

	self.PubKeys = make([]*crypto.PubKey, n)
	for i := 0; i < int(n); i++ {
		pubkey := new(crypto.PubKey)
		err = pubkey.DeSerialize(r)
		if err != nil {
			return err
		}
		self.PubKeys[i] = pubkey
	}

	self.M, err = serialization.ReadUint8(r)
	if err != nil {
		return err
	}

	m, err := serialization.ReadVarUint(r, 0)
	if err != nil {
		return err
	}

	self.SigData = make([][]byte, m)
	for i := 0; i < int(m); i++ {
		sig, err := serialization.ReadVarBytes(r)
		if err != nil {
			return err
		}
		self.SigData[i] = sig
	}

	return nil
}

func (self *Transaction) GetSignatureAddresses() []Uint160{
	address := make([]Uint160, 0, len(self.Sigs))
	for _, sig := range self.Sigs {
		m := int(sig.M)
		n := len(sig.PubKeys)

		if n == 1 {
			address = append(address, AddressFromPubKey(sig.PubKeys[0]))
		} else {
			addr, _ := AddressFromMultiPubKeys(sig.PubKeys, m)
			address = append(address, addr)
		}
	}

	return address
}

func (self *Sig) Serialize(w io.Writer) error {
	err := serialization.WriteVarUint(w, uint64(len(self.PubKeys)))
	if err != nil {
		return errors.New("serialize sig pubkey length failed")
	}
	for _, pubkey := range self.PubKeys {
		err = pubkey.Serialize(w)
		if err != nil {
			return err
		}
	}

	err = serialization.WriteUint8(w, self.M)
	if err != nil {
		return errors.New("serialize Sig M failed")
	}

	err = serialization.WriteVarUint(w, uint64(len(self.SigData)))
	if err != nil {
		return errors.New("serialize sig pubkey length failed")
	}

	for _, sig := range self.SigData {
		err = serialization.WriteVarBytes(w, sig)
		if err != nil {
			return err
		}
	}

	return nil
}

type Fee struct {
	Amount Fixed64
	Payer  Uint160
}

type TransactionType byte

const (
	BookKeeping    TransactionType = 0x00
	IssueAsset     TransactionType = 0x01
	BookKeeper     TransactionType = 0x02
	Claim          TransactionType = 0x03
	PrivacyPayload TransactionType = 0x20
	RegisterAsset  TransactionType = 0x40
	TransferAsset  TransactionType = 0x80
	Record         TransactionType = 0x81
	Deploy         TransactionType = 0xd0
	Invoke         TransactionType = 0xd1
	DataFile       TransactionType = 0x12
	Enrollment     TransactionType = 0x04
	Vote           TransactionType = 0x05
)

var TxName = map[TransactionType]string{
	BookKeeping:    "BookKeeping",
	IssueAsset:     "IssueAsset",
	BookKeeper:     "BookKeeper",
	Claim:          "Claim",
	PrivacyPayload: "PrivacyPayload",
	RegisterAsset:  "RegisterAsset",
	TransferAsset:  "TransferAsset",
	Record:         "Record",
	Deploy:         "Deploy",
	Invoke:         "Invoke",
	DataFile:       "DataFile",
	Enrollment:     "Enrollment",
	Vote:           "Vote",
}

//Payload define the func for loading the payload data
//base on payload type which have different struture
type Payload interface {
	//  Get payload data
	Data() []byte

	//Serialize payload data
	Serialize(w io.Writer) error

	Deserialize(r io.Reader) error
}

//Serialize the Transaction
func (tx *Transaction) Serialize(w io.Writer) error {

	err := tx.SerializeUnsigned(w)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Transaction txSerializeUnsigned Serialize failed.")
	}

	err = serialization.WriteVarUint(w, uint64(len(tx.Sigs)))
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "serialize tx sigs length failed")
	}
	for _, sig := range tx.Sigs {
		err = sig.Serialize(w)
		if err != nil {
			return err
		}
	}

	return nil
}

func (tx *Transaction) GetTotalFee() Fixed64 {
	sum := Fixed64(0)
	for _, fee := range tx.Fee {
		sum += fee.Amount
	}
	return sum
}

//Serialize the Transaction data without contracts
func (tx *Transaction) SerializeUnsigned(w io.Writer) error {
	//txType
	w.Write([]byte{byte(tx.Version), byte(tx.TxType)})
	serialization.WriteUint32(w, tx.Nonce)
	//Payload
	if tx.Payload == nil {
		return errors.New("Transaction Payload is nil.")
	}
	tx.Payload.Serialize(w)
	//[]*txAttribute
	err := serialization.WriteVarUint(w, uint64(len(tx.Attributes)))
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Transaction item txAttribute length serialization failed.")
	}
	for _, attr := range tx.Attributes {
		attr.Serialize(w)
	}

	err = serialization.WriteVarUint(w, uint64(len(tx.Fee)))
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "serialize tx fee length failed")
	}
	for _, fee := range tx.Fee {
		fee.Amount.Serialize(w)
		fee.Payer.Serialize(w)
	}

	tx.NetWorkFee.Serialize(w)

	return nil
}

//deserialize the Transaction
func (tx *Transaction) Deserialize(r io.Reader) error {
	// tx deserialize
	err := tx.DeserializeUnsigned(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "transaction Deserialize error")
	}

	// tx sigs
	length, err := serialization.ReadVarUint(r, 0)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "transaction sigs deserialize error")
	}

	tx.Sigs = make([]*Sig, 0, length)
	for i := 0; i < int(length); i++ {
		sig := new(Sig)
		err := sig.Deserialize(r)
		if err != nil {
			return errors.New("deserialize transaction failed")
		}
		tx.Sigs = append(tx.Sigs, sig)
	}

	return nil
}

func (tx *Transaction) DeserializeUnsigned(r io.Reader) error {
	var versiontype [2]byte
	r.Read(versiontype[:])
	nonce, err := serialization.ReadUint32(r)
	if err != nil {
		return err
	}
	tx.Version = versiontype[0]
	tx.TxType = TransactionType(versiontype[1])
	tx.Nonce = nonce

	switch tx.TxType {
	case Invoke:
		tx.Payload = new(payload.InvokeCode)
	case BookKeeping:
		tx.Payload = new(payload.BookKeeping)
	case Deploy:
		tx.Payload = new(payload.DeployCode)
	default:
		return fmt.Errorf("unsupported tx type %v", tx.Type())
	}

	err = tx.Payload.Deserialize(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Payload Parse error")
	}

	//attributes
	length, err := serialization.ReadVarUint(r, 0)
	if err != nil {
		return err
	}
	for i := uint64(0); i < length; i++ {
		attr := new(TxAttribute)
		err = attr.Deserialize(r)
		if err != nil {
			return err
		}
		tx.Attributes = append(tx.Attributes, attr)
	}

	length, err = serialization.ReadVarUint(r, 0)
	if err != nil {
		return err
	}
	for i := uint64(0); i < length; i++ {
		fee := new(Fee)
		err = fee.Amount.Deserialize(r)
		if err != nil {
			return err
		}
		err = fee.Payer.Deserialize(r)
		if err != nil {
			return err
		}
		tx.Fee = append(tx.Fee, fee)
	}
	err = tx.NetWorkFee.Deserialize(r)
	if err != nil {
		return err
	}

	return nil
}

func (tx *Transaction) GetMessage() []byte {
	buf := new(bytes.Buffer)
	tx.SerializeUnsigned(buf)
	return buf.Bytes()
}

func (tx *Transaction) ToArray() []byte {
	b := new(bytes.Buffer)
	tx.Serialize(b)
	return b.Bytes()
}

func (tx *Transaction) Hash() Uint256 {
	if tx.hash == nil {
		buf := bytes.Buffer{}
		tx.SerializeUnsigned(&buf)
		temp := sha256.Sum256(buf.Bytes())
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
	panic("unimplemented ")
	return nil
}

func (tx *Transaction) GetSysFee() Fixed64 {
	return Fixed64(Parameters.SystemFee[TxName[tx.TxType]])
}

func (tx *Transaction) GetNetworkFee() Fixed64 {
	return tx.NetWorkFee
}

