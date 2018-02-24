package payload

import (
	"io"

	"github.com/Ontology/common/serialization"
)

type BookKeeping struct {
	Nonce uint64
}

func (a *BookKeeping) Data() []byte {
	return []byte{0}
}

func (a *BookKeeping) Serialize(w io.Writer) error {
	err := serialization.WriteUint64(w, a.Nonce)
	if err != nil {
		return err
	}
	return nil
}

func (a *BookKeeping) Deserialize(r io.Reader) error {
	var err error
	a.Nonce, err = serialization.ReadUint64(r)
	if err != nil {
		return err
	}
	return nil
}
