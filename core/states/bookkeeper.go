package states

import (
	"io"
	"bytes"
	"github.com/Ontology/common/serialization"
	"github.com/Ontology/crypto"
)

type BookKeeperState struct {
	StateBase
	CurrBookKeeper []*crypto.PubKey
	NextBookKeeper []*crypto.PubKey
}

func (this *BookKeeperState) Serialize(w io.Writer) error {
	this.StateBase.Serialize(w)
	serialization.WriteUint32(w, uint32(len(this.CurrBookKeeper)))
	for _, v := range this.CurrBookKeeper {
		v.Serialize(w)
	}
	serialization.WriteUint32(w, uint32(len(this.NextBookKeeper)))
	for _, v := range this.NextBookKeeper {
		v.Serialize(w)
	}
	return nil
}

func (this *BookKeeperState) Deserialize(r io.Reader) error {
	if this == nil {
		this = new(BookKeeperState)
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
		p := new(crypto.PubKey)
		err = p.DeSerialize(r)
		if err != nil {
			return err
		}
		this.CurrBookKeeper = append(this.CurrBookKeeper, p)
	}

	n, err = serialization.ReadUint32(r)
	if err != nil {
		return err
	}
	for i := 0; i < int(n); i++ {
		p := new(crypto.PubKey)
		err = p.DeSerialize(r)
		if err != nil {
			return err
		}
		this.NextBookKeeper = append(this.NextBookKeeper, p)
	}
	return nil
}

func (v *BookKeeperState) ToArray() []byte {
	b := new(bytes.Buffer)
	v.Serialize(b)
	return b.Bytes()
}

