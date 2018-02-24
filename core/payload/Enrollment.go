package payload

import (
	"github.com/Ontology/crypto"
	"io"
)

type Enrollment struct {
	PublicKey *crypto.PubKey
}

func (e *Enrollment) Data() []byte {
	return []byte{0}
}

func (e *Enrollment) Serialize(w io.Writer) error {
	if err := e.PublicKey.Serialize(w); err != nil {
		return err
	}
	return nil
}

func (e *Enrollment) Deserialize(r io.Reader) error {
	pk := new(crypto.PubKey)
	if err := pk.DeSerialize(r); err != nil {
		return err
	}
	e.PublicKey = pk
	return nil
}
