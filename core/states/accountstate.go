package states

import (
	"io"
	"bytes"
	"github.com/Ontology/common/serialization"
	. "github.com/Ontology/common"
	. "github.com/Ontology/errors"
)

type AccountState struct {
	StateBase
	ProgramHash Uint160
	IsFrozen    bool
	Balances    []*Balance
}

type Balance struct {
	StateBase
	AssetId Uint256
	Amount Fixed64
}

func (this *Balance) Serialize(w io.Writer) error {
	var err error
	err = this.StateBase.Serialize(w)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[Balance] StateBase Serialize failed.")
	}
	_, err = this.AssetId.Serialize(w)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[Balance] AssetId Serialize failed.")
	}
	err = this.Amount.Serialize(w)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[Balance] Amount Serialize failed.")
	}
	return nil
}

func (this *Balance) Deserialize(r io.Reader) error {
	if this == nil {
		this = new(Balance)
	}
	err := this.StateBase.Deserialize(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[Balance] StateBase Deserialize failed.")
	}
	u := new(Uint256)
	if err = u.Deserialize(r); err != nil {
		return NewDetailErr(err, ErrNoCode, "[Balance] AssetId Deserialize failed.")
	}
	f := new(Fixed64)
	if err = f.Deserialize(r); err != nil {
		return NewDetailErr(err, ErrNoCode, "[Balance] Amount Deserialize failed.")
	}
	return nil
}

func (this *AccountState) Serialize(w io.Writer) error {
	var err error
	err = this.StateBase.Serialize(w)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[AccountState] StateBase Serialize failed.")
	}
	_, err = this.ProgramHash.Serialize(w)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[AccountState] ProgramHash Serialize failed.")
	}
	err = serialization.WriteBool(w, this.IsFrozen)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[AccountState] IsFrozen Serialize failed.")
	}
	err = serialization.WriteUint64(w, uint64(len(this.Balances)))
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[AccountState] Balances Length Serialize failed.")
	}
	for _, v := range this.Balances {
		err = v.Serialize(w)
		if err != nil {
			return NewDetailErr(err, ErrNoCode, "[AccountState] Balances Serialize failed.")
		}
	}
	return nil
}

func (this *AccountState) Deserialize(r io.Reader) error {
	if this == nil {
		this = new(AccountState)
	}
	err := this.StateBase.Deserialize(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[AccountState] StateBase Deserialize failed.")
	}
	err = this.ProgramHash.Deserialize(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[AccountState] ProgramHash Deserialize failed.")
	}
	isFrozen, err := serialization.ReadBool(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[AccountState] IsFrozen Deserialize failed.")
	}
	this.IsFrozen = isFrozen
	l, err := serialization.ReadUint64(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[AccountState] Balances Length Deserialize failed.")
	}
	for i := 0; i < int(l); i++ {
		b := new(Balance)
		err := b.Deserialize(r)
		if err != nil {
			return NewDetailErr(err, ErrNoCode, "[AccountState] Balances Deserialize failed.")
		}
		this.Balances = append(this.Balances, b)
	}
	return nil
}

func (accountState *AccountState) ToArray() []byte {
	b := new(bytes.Buffer)
	accountState.Serialize(b)
	return b.Bytes()
}


