package states

import (
	"github.com/Ontology/common"
	"io"
	. "github.com/Ontology/errors"
)

type ContractMapping struct {
	OriginAddress common.Address
	TargetAddress common.Address
}

func (this *ContractMapping) Serialize(w io.Writer) error {
	if err := this.OriginAddress.Serialize(w); err != nil {
		return NewDetailErr(err, ErrNoCode, "[ContractMapping] OriginAddress serialize failed.")
	}
	if err := this.TargetAddress.Serialize(w); err != nil {
		return NewDetailErr(err, ErrNoCode, "[ContractMapping] TargetAddress serialize failed.")
	}
	return nil
}

func (this *ContractMapping) Deserialize(r io.Reader) error {
	origin := new(common.Address)
	if err := origin.Deserialize(r); err != nil {
		return NewDetailErr(err, ErrNoCode, "[ContractMapping] OriginAddress deserialize failed.")
	}

	target := new(common.Address)
	if err := target.Deserialize(r); err != nil {
		return NewDetailErr(err, ErrNoCode, "[ContractMapping] TargetAddress deserialize failed.")
	}
	return nil
}