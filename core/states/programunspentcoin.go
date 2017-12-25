package states

import (
	"github.com/Ontology/core/transaction/utxo"
	"io"
	"github.com/Ontology/common/serialization"
)

type ProgramUnspentCoin struct {
	StateBase
	Unspents []*utxo.UTXOUnspent
}

func (this *ProgramUnspentCoin) Serialize(w io.Writer) error {
	this.StateBase.Serialize(w)
	serialization.WriteUint32(w, uint32(len(this.Unspents)))
	for _, v := range this.Unspents {
		v.Serialize(w)
	}
	return nil
}

func (this *ProgramUnspentCoin) Deserialize(r io.Reader) error {
	if this == nil {
		this = new(ProgramUnspentCoin)
	}
	err := this.StateBase.Deserialize(r)
	if err != nil {
		return err
	}
	n, err := serialization.ReadUint32(r)
	if err != nil {
		return err
	}
	for i := 0; i < int(n); i++ {
		u := new(utxo.UTXOUnspent)
		if err := u.Deserialize(r); err != nil {
			return err
		}
		this.Unspents = append(this.Unspents, u)
	}
	return nil
}
