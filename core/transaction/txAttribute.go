package transaction

import (
	"io"

	"DNA/common/serialization"
	. "DNA/errors"
)

type TransactionAttributeUsage byte

const (
	Script         TransactionAttributeUsage = 0x20
	DescriptionUrl TransactionAttributeUsage = 0x81
	Description    TransactionAttributeUsage = 0x90
)

type TxAttribute struct {
	Usage TransactionAttributeUsage
	Data  []byte
	Size  uint32
}

func NewTxAttribute(u TransactionAttributeUsage, d []byte) TxAttribute {
	tx := TxAttribute{u, d, 0}
	tx.Size = tx.GetSize()
	return tx
}

func (u *TxAttribute) GetSize() uint32 {
	if u.Usage == DescriptionUrl {
		return uint32(len([]byte{(byte(0xff))}) + len([]byte{(byte(0xff))}) + len(u.Data))
	}
	return 0
}

func (tx *TxAttribute) Serialize(w io.Writer) error {
	err := serialization.WriteUint8(w, byte(tx.Usage))
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Transaction attribute Usage serialization error.")
	}
	// TODO other transaction attribute type
	if tx.Usage == Description || tx.Usage == DescriptionUrl {
		err := serialization.WriteVarBytes(w, tx.Data)
		if err != nil {
			return NewDetailErr(err, ErrNoCode, "Transaction attribute Data serialization error.")
		}
	} else {
		return NewDetailErr(err, ErrNoCode, "Unsupported attribute Description.")
	}
	return nil
}

func (tx *TxAttribute) Deserialize(r io.Reader) error {
	val, err := serialization.ReadBytes(r, 1)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Transaction attribute Usage deserialization error.")
	}
	tx.Usage = TransactionAttributeUsage(val[0])
	if tx.Usage == Description || tx.Usage == DescriptionUrl {
		tx.Data, err = serialization.ReadVarBytes(r)
		if err != nil {
			return NewDetailErr(err, ErrNoCode, "Transaction attribute Data deserialization error.")
		}
	} else {
		return NewDetailErr(err, ErrNoCode, "Unsupported attribute description.")
	}
	return nil
}
