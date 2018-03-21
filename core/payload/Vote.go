package payload

import (
	. "github.com/Ontology/common"
	"github.com/Ontology/common/serialization"
	"github.com/Ontology/crypto"
	"io"
	. "github.com/Ontology/errors"
)

const (
	MaxVoteKeys = 1024
)

type Vote struct {
	PubKeys []*crypto.PubKey // vote node list

	Account Address
}

func (self *Vote) Check() bool {
	if len(self.PubKeys) > MaxVoteKeys {
		return false
	}
	return true
}

func (self *Vote) Serialize(w io.Writer) error {
	if err := serialization.WriteUint32(w, uint32(len(self.PubKeys))); err != nil {
		return NewDetailErr(err, ErrNoCode, "Vote PubKeys length Serialize failed.")
	}
	for _, key := range self.PubKeys {
		if err := key.Serialize(w); err != nil {
			return NewDetailErr(err, ErrNoCode, "InvokeCode PubKeys Serialize failed.")
		}
	}
	if _, err := self.Account.Serialize(w); err != nil {
		return NewDetailErr(err, ErrNoCode, "InvokeCode Account Serialize failed.")
	}

	return nil
}

func (self *Vote) Deserialize(r io.Reader) error {
	length, err := serialization.ReadUint32(r)
	if err != nil {
		return err
	}
	self.PubKeys = make([]*crypto.PubKey, length)
	for i := 0; i < int(length); i++ {
		pubkey := new(crypto.PubKey)
		err := pubkey.DeSerialize(r)
		if err != nil {
			return err
		}
		self.PubKeys[i] = pubkey
	}

	err = self.Account.Deserialize(r)
	if err != nil {
		return err
	}

	return nil
}
