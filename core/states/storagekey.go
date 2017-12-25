package states

import (
	"io"
	"github.com/Ontology/common"
	"github.com/Ontology/common/serialization"
	"bytes"
)

type StorageKey struct {
	CodeHash common.Uint160
	Key      []byte
}

func (this *StorageKey) Serialize(w io.Writer) (int, error) {
	this.CodeHash.Serialize(w)
	serialization.WriteVarBytes(w, this.Key)
	return 0, nil
}

func (this *StorageKey) Deserialize(r io.Reader) error {
	u := new(common.Uint160)
	err := u.Deserialize(r)
	if err != nil {
		return err
	}
	this.CodeHash = *u
	key, err := serialization.ReadVarBytes(r)
	if err != nil {
		return err
	}
	this.Key = key
	return nil
}

func (this *StorageKey) ToArray() []byte {
	b := new(bytes.Buffer)
	this.Serialize(b)
	return b.Bytes()
}

