package utxo

import (
	"github.com/Ontology/common"
	"io"
	"bytes"
	"github.com/Ontology/vm/neovm/interfaces"
)

type TxOutput struct {
	AssetID     common.Uint256
	Value       common.Fixed64
	ProgramHash common.Uint160
}

func (o *TxOutput) Serialize(w io.Writer) {
	o.AssetID.Serialize(w)
	o.Value.Serialize(w)
	o.ProgramHash.Serialize(w)
}

func (o *TxOutput) Deserialize(r io.Reader) {
	o.AssetID.Deserialize(r)
	o.Value.Deserialize(r)
	o.ProgramHash.Deserialize(r)
}

func (o *TxOutput) ToArray() []byte {
	bf := new(bytes.Buffer)
	o.Serialize(bf)
	return bf.Bytes()
}

func (o *TxOutput) Clone() interfaces.IInteropInterface {
	to := *o
	return &to
}


