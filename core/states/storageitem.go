package states

import (
	"io"
	"bytes"
	"github.com/Ontology/common/serialization"
	"github.com/Ontology/vm/neovm/interfaces"
)

type StorageItem struct {
	StateBase
	Value []byte
}

func (this *StorageItem) Serialize(w io.Writer) error {
	this.StateBase.Serialize(w)
	serialization.WriteVarBytes(w, this.Value)
	return nil
}

func (this *StorageItem) Deserialize(r io.Reader) error {
	if this == nil {
		this = new(StorageItem)
	}
	err := this.StateBase.Deserialize(r)
	if err != nil {
		return err
	}
	value, err := serialization.ReadVarBytes(r)
	if err != nil {
		return err
	}
	this.Value = value
	return nil
}

func (storageItem *StorageItem) ToArray() []byte {
	b := new(bytes.Buffer)
	storageItem.Serialize(b)
	return b.Bytes()
}

func (storageItem *StorageItem) Clone() interfaces.IInteropInterface {
	si := *storageItem
	return &si
}
