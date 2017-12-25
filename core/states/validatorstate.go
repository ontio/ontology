package states

import (
	"github.com/Ontology/crypto"
	"io"
)

type ValidatorState struct {
	StateBase
	PublicKey *crypto.PubKey
}

func (this *ValidatorState) Serialize(w io.Writer) error {
	this.StateBase.Serialize(w)
	if err := this.PublicKey.Serialize(w); err != nil {
		return err
	}
	return nil
}

func (this *ValidatorState) Deserialize(r io.Reader) error {
	if this == nil {
		this = new(ValidatorState)
	}
	err := this.StateBase.Deserialize(r)
	if err != nil {
		return err
	}
	pk := new(crypto.PubKey)
	if err := pk.DeSerialize(r); err != nil {
		return err
	}
	this.PublicKey = pk
	return nil
}