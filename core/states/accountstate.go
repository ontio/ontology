package states

import (
	"io"
	"bytes"
	"github.com/Ontology/common/serialization"
	. "github.com/Ontology/common"
)

type AccountState struct {
	StateBase
	ProgramHash Uint160
	IsFrozen    bool
	Balances    map[Uint256]Fixed64
}

func NewAccountState() *AccountState {
	return &AccountState{
		Balances: make(map[Uint256]Fixed64),
	}
}

func (this *AccountState) Serialize(w io.Writer) error {
	this.StateBase.Serialize(w)
	this.ProgramHash.Serialize(w)
	serialization.WriteBool(w, this.IsFrozen)
	serialization.WriteUint64(w, uint64(len(this.Balances)))
	for k, v := range this.Balances {
		k.Serialize(w)
		v.Serialize(w)
	}
	return nil
}

func (this *AccountState) Deserialize(r io.Reader) error {
	if this == nil {
		this = NewAccountState()
	}
	err := this.StateBase.Deserialize(r)
	if err != nil {
		return err
	}
	this.ProgramHash.Deserialize(r)
	isFrozen, err := serialization.ReadBool(r)
	if err != nil {
		return err
	}
	this.IsFrozen = isFrozen
	l, err := serialization.ReadUint64(r)
	if err != nil {
		return err
	}
	balances := make(map[Uint256]Fixed64)
	u := new(Uint256)
	f := new(Fixed64)
	for i := 0; i < int(l); i++ {
		if err = u.Deserialize(r); err != nil {
			return err
		}
		if err = f.Deserialize(r); err != nil {
			return err
		}
		balances[*u] = *f
	}
	this.Balances = balances
	return nil
}

func (accountState *AccountState) ToArray() []byte {
	b := new(bytes.Buffer)
	accountState.Serialize(b)
	return b.Bytes()
}


