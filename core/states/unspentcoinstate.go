package states

import (
	"io"
	"github.com/Ontology/common/serialization"
)

type UnspentCoinState struct {
	StateBase
	Item []CoinState
}

func (this *UnspentCoinState) Serialize(w io.Writer) error {
	this.StateBase.Serialize(w)
	serialization.WriteUint32(w, uint32(len(this.Item)))
	for _, v := range this.Item {
		serialization.WriteByte(w, byte(v))
	}
	return nil
}

func (this *UnspentCoinState) Deserialize(r io.Reader) error {
	if this == nil {
		this = new(UnspentCoinState)
	}
	err := this.StateBase.Deserialize(r)
	if err != nil {
		return err
	}
	n, err := serialization.ReadUint32(r)
	if err != nil {
		return err
	}
	for i := uint32(0); i < n; i++ {
		state, err := serialization.ReadByte(r)
		if err != nil {
			return err
		}
		this.Item = append(this.Item, CoinState(state))
	}
	return nil
}
