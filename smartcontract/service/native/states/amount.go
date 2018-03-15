package states

import (
	"math/big"
	"io"
	"github.com/Ontology/common/serialization"
	"github.com/Ontology/errors"
)

type Amount struct {
	Value *big.Int
}

func(this *Amount) Serialize(w io.Writer) error {
	return serialization.WriteVarBytes(w, this.Value.Bytes())
}

func(this *Amount) Deserialize(r io.Reader) error {
	if this == nil {
		this = new(Amount)
	}
	bs, err := serialization.ReadVarBytes(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[TotalSupply Deserialize] read value error!")
	}
	this.Value = new(big.Int).SetBytes(bs)
	return nil
}
