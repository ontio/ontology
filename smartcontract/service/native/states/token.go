package states

import (
	"github.com/Ontology/common"
	"github.com/Ontology/common/serialization"
	"github.com/Ontology/errors"
	"io"
	"math/big"
)

type Transfers struct {
	Params []*TokenTransfer
}

func (this *Transfers) Serialize(w io.Writer) error {
	if err := serialization.WriteVarUint(w, uint64(len(this.Params))); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Transfers Serialize] TokenTransfer length error!")
	}
	for _, v := range this.Params {
		if err := v.Serialize(w); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[Transfers Serialize] TokenTransfer error!")
		}
	}
	return nil
}

func (this *Transfers) Deserialize(r io.Reader) error {
	n, err := serialization.ReadVarUint(r, 0)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Transfers Deserialize] TokenTransfer length error!")
	}
	for i := 0; uint64(i) < n; i++ {
		tokenTransfer := new(TokenTransfer)
		if err := tokenTransfer.Deserialize(r); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[Transfers Deserialize] TokenTransfer error!")
		}
		this.Params = append(this.Params, tokenTransfer)
	}
	return nil
}

type TokenTransfer struct {
	Contract common.Uint160
	States   []*State
}

func (this *TokenTransfer) Serialize(w io.Writer) error {
	if _, err := this.Contract.Serialize(w); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[TokenTransfer Serialize] Contract error!")
	}
	if err := serialization.WriteVarUint(w, uint64(len(this.States))); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[TokenTransfer Serialize] States length error!")
	}
	for _, v := range this.States {
		if err := v.Serialize(w); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[TokenTransfer Serialize] States error!")
		}
	}
	return nil
}

func (this *TokenTransfer) Deserialize(r io.Reader) error {
	contract := new(common.Uint160)
	if err := contract.Deserialize(r); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[TokenTransfer Deserialize] Contract error!")
	}
	n, err := serialization.ReadVarUint(r, 0)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[TokenTransfer Deserialize] States length error!")
	}
	for i := 0; uint64(i) < n; i++ {
		state := new(State)
		if err := state.Deserialize(r); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[TokenTransfer Deserialize] States error!")
		}
		this.States = append(this.States, state)
	}
	return nil
}

type State struct {
	From  common.Uint160
	To    common.Uint160
	Value *big.Int
}

func (this *State) Serialize(w io.Writer) error {
	if _, err := this.From.Serialize(w); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[State Serialize] From error!")
	}
	if _, err := this.To.Serialize(w); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[State Serialize] To error!")
	}
	if err := serialization.WriteVarBytes(w, this.Value.Bytes()); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[State Serialize] Value error!")
	}
	return nil
}

func (this *State) Deserialize(r io.Reader) error {
	from := new(common.Uint160)
	if err := from.Deserialize(r); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[State Deserialize] From error!")
	}
	this.From = *from

	to := new(common.Uint160)
	if err := from.Deserialize(r); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[State Deserialize] To error!")
	}
	this.From = *to

	value, err := serialization.ReadVarBytes(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[State Deserialize] Value error!")
	}

	this.Value = new(big.Int).SetBytes(value)
	return nil
}
