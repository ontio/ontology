package payload

import (
	"bytes"
	. "github.com/Ontology/common"
	"github.com/Ontology/common/serialization"
	"github.com/Ontology/crypto"
	"io"
)

const (
	MaxVoteKeys = 1024
)

type Vote struct {
	PubKeys []*crypto.PubKey // vote node list

	Account Uint160
}

func (self *Vote) Check() bool {
	if len(self.PubKeys) > MaxVoteKeys {
		return false
	}
	return true
}

func (self *Vote) Data() []byte {
	var buf bytes.Buffer
	serialization.WriteUint32(&buf, uint32(len(self.PubKeys)))
	for _, key := range self.PubKeys {
		key.Serialize(&buf)
	}
	self.Account.Serialize(&buf)

	return buf.Bytes()
}

func (self *Vote) Serialize(w io.Writer) error {
	_, err := w.Write(self.Data())

	return err
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
