package states

import (
	"github.com/Ontology/common"
	"github.com/Ontology/common/serialization"
	"github.com/Ontology/errors"
	"io"
	"math/big"
)

type Transfers struct {
	Version byte
	States   []*State
}

func (this *Transfers) Serialize(w io.Writer) error {
	if err := serialization.WriteByte(w, byte(this.Version)); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[TokenTransfer] Serialize version error!")
	}
	if err := serialization.WriteVarUint(w, uint64(len(this.States))); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[TokenTransfer] Serialize States length error!")
	}
	for _, v := range this.States {
		if err := v.Serialize(w); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[TokenTransfer] Serialize States error!")
		}
	}
	return nil
}

func (this *Transfers) Deserialize(r io.Reader) error {
	version, err := serialization.ReadByte(r); if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[TokenTransfer] Deserialize version error!")
	}
	this.Version = version

	n, err := serialization.ReadVarUint(r, 0); if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[TokenTransfer] Deserialize states length error!")
	}
	for i := 0; uint64(i) < n; i++ {
		state := new(State)
		if err := state.Deserialize(r); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[TokenTransfer] Deserialize states error!")
		}
		this.States = append(this.States, state)
	}
	return nil
}

type State struct {
	Version byte
	From  common.Address
	To    common.Address
	Value *big.Int
}

func (this *State) Serialize(w io.Writer) error {
	if err := serialization.WriteByte(w, byte(this.Version)); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[State] Serialize version error!")
	}
	if err := this.From.Serialize(w); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[State] Serialize From error!")
	}
	if err := this.To.Serialize(w); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[State] Serialize To error!")
	}
	if err := serialization.WriteVarBytes(w, this.Value.Bytes()); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[State] Serialize Value error!")
	}
	return nil
}

func (this *State) Deserialize(r io.Reader) error {
	version, err := serialization.ReadByte(r); if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[State] Deserialize version error!")
	}
	this.Version = version

	from := new(common.Address)
	if err := from.Deserialize(r); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[State] Deserialize from error!")
	}
	this.From = *from

	to := new(common.Address)
	if err := to.Deserialize(r); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[State] Deserialize to error!")
	}
	this.To = *to

	value, err := serialization.ReadVarBytes(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[State] Deserialize value error!")
	}

	this.Value = new(big.Int).SetBytes(value)
	return nil
}
