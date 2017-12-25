package states

import (
	"github.com/Ontology/crypto"
	"github.com/Ontology/common"
	"io"
	"github.com/Ontology/common/serialization"
)

type VoteState struct {
	StateBase
	PublicKeys []*crypto.PubKey
	Count      common.Fixed64
}

func (this *VoteState) Serialize(w io.Writer) error {
	this.StateBase.Serialize(w)
	serialization.WriteUint32(w, uint32(len(this.PublicKeys)))
	for _, v := range this.PublicKeys {
		err := v.Serialize(w)
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *VoteState) Deserialize(r io.Reader) error {
	if this == nil {
		this = new(VoteState)
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
		pk := new(crypto.PubKey)
		if err := pk.DeSerialize(r); err != nil {
			return err
		}
		this.PublicKeys = append(this.PublicKeys, pk)
	}
	return nil
}


