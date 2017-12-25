package utxo

import (
	"github.com/Ontology/common"
	"github.com/Ontology/common/serialization"
	"io"
	"bytes"
)

type UTXOUnspent struct {
	Txid  common.Uint256
	Index uint32
	Value common.Fixed64
}

func (uu *UTXOUnspent) Serialize(w io.Writer) {
	uu.Txid.Serialize(w)
	serialization.WriteUint32(w, uu.Index)
	uu.Value.Serialize(w)
}

func (uu *UTXOUnspent) Deserialize(r io.Reader) error {
	uu.Txid.Deserialize(r)

	index, err := serialization.ReadUint32(r)
	uu.Index = uint32(index)
	if err != nil {
		return err
	}

	uu.Value.Deserialize(r)

	return nil
}

func (uu *UTXOUnspent) ToArray() []byte {
	bf := new(bytes.Buffer)
	uu.Serialize(bf)
	return bf.Bytes()
}