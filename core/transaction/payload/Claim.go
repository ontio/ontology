package payload

import (
	"errors"
	"github.com/Ontology/common/serialization"
	"github.com/Ontology/core/transaction/utxo"
	. "github.com/Ontology/errors"
	"io"
)

const ClaimPayloadVersion byte = 0x00

type Claim struct {
	Claims []*utxo.UTXOTxInput
}

func (a *Claim) Data(version byte) []byte {
	//TODO: implement RegisterClaim.Data()
	return []byte{0}
}

// Serialize is the implement of SignableData interface.
func (this *Claim) Serialize(w io.Writer, version byte) error {
	count := uint32(len(this.Claims))
	err := serialization.WriteUint32(w, count)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[Claim], Claim serialize failed.")
	}
	for _, v := range this.Claims {
		v.Serialize(w)
	}
	return nil
}

// Deserialize is the implement of SignableData interface.
func (a *Claim) Deserialize(r io.Reader, version byte) error {
	if a == nil {
		a = &Claim{[]*utxo.UTXOTxInput{}}
	}
	count, err := serialization.ReadUint32(r)
	if err != nil {
		return NewDetailErr(errors.New("[Claim], Claim deserialize failed."), ErrNoCode, "")
	}

	for i := 0; i < int(count); i++ {
		claim_ := new(utxo.UTXOTxInput)
		err := claim_.Deserialize(r)
		if err != nil {
			return err
		}
		a.Claims = append(a.Claims, claim_)
	}
	return nil
}