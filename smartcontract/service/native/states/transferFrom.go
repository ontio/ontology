package states

import (
	"github.com/Ontology/common"
	"math/big"
	"io"
	"github.com/Ontology/errors"
	"github.com/Ontology/common/serialization"
)

type TransferFrom struct {
	Version byte
	Sender common.Address
	From common.Address
	To common.Address
	Value *big.Int
}

func (this *TransferFrom) Serialize(w io.Writer) error {
	if err := serialization.WriteByte(w, byte(this.Version)); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[TransferFrom] Serialize version error!")
	}
	if err := this.Sender.Serialize(w); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[TransferFrom] Serialize sender error!")
	}
	if err := this.From.Serialize(w); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[TransferFrom] Serialize from error!")
	}
	if err := this.To.Serialize(w); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[TransferFrom] Serialize to error!")
	}
	if err := serialization.WriteVarBytes(w, this.Value.Bytes()); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[TransferFrom] Serialize value error!")
	}
	return nil
}

func (this *TransferFrom) Deserialize(r io.Reader) error {
	version, err := serialization.ReadByte(r); if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[TransferFrom] Deserialize version error!")
	}
	this.Version = version

	sender := new(common.Address)
	if err := sender.Deserialize(r); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[TransferFrom] Deserialize sender error!")
	}
	this.Sender = *sender

	from := new(common.Address)
	if err := from.Deserialize(r); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[TransferFrom] Deserialize from error!")
	}
	this.From = *from

	to := new(common.Address)
	if err := to.Deserialize(r); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[TransferFrom] Deserialize to error!")
	}
	this.To = *to

	value, err := serialization.ReadVarBytes(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[TransferFrom] Deserialize value error!")
	}

	this.Value = new(big.Int).SetBytes(value)
	return nil
}
