package states

import (
	"github.com/Ontology/common"
	"io"
	"github.com/Ontology/common/serialization"
	"github.com/Ontology/errors"
)

type Contract struct {
	Version byte
	Address common.Address
	Method string
	Args []byte
}

func (this *Contract) Serialize(w io.Writer) error {
	if err := serialization.WriteByte(w, this.Version); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Contract] Version serialize error!")
	}
	if err := this.Address.Serialize(w); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Contract] Address serialize error!")
	}
	if err := serialization.WriteVarBytes(w, []byte(this.Method)); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Contract] Method serialize error!")
	}
	if err := serialization.WriteVarBytes(w, this.Args); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Contract] Args serialize error!")
	}
	return nil
}

func (this *Contract) Deserialize(r io.Reader) error {
	var err error
	this.Version, err = serialization.ReadByte(r); if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Contract] Version deserialize error!")
	}

	address := new(common.Address)
	if err := address.Deserialize(r); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Contract] Address deserialize error!")
	}

	method, err := serialization.ReadVarBytes(r); if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Contract] Method deserialize error!")
	}
	this.Method = string(method)

	this.Args, err = serialization.ReadVarBytes(r); if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Contract] Args deserialize error!")
	}
	return nil
}
