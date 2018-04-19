package states

import (
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/errors"
	"io"
)

type Param struct { // ontology network environment variables
	Version byte
	K       string
	V       string
}

type Params struct {
	Version   byte
	ParamList []*Param
}

func (param *Param) Serialize(w io.Writer) error {
	if err := serialization.WriteByte(w, byte(param.Version)); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Param Config] Serialize version error!")
	}

	if err := serialization.WriteString(w, param.K); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Param Config] Serialize param name error!")
	}

	if err := serialization.WriteString(w, param.V); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Param Config] Serialize param value error!")
	}
	return nil
}

func (param *Param) Deserialize(r io.Reader) error {
	version, err := serialization.ReadByte(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Param Config] Deserialize version error!")
	}
	param.Version = version

	var key string
	key, err = serialization.ReadString(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Param Config] Deserialize parm name error!")
	}
	param.K = key

	var value string
	value, err = serialization.ReadString(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode,
			"[Param Config] Deserialize param value error!")
	}
	param.V = value
	return nil
}

func (params *Params) Serialize(w io.Writer) error {
	if err := serialization.WriteByte(w, params.Version); err != nil {
		return err
	}

	if err := serialization.WriteVarUint(w, uint64(len(params.ParamList))); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode,
			"[Param Config] Serialize param num error!")
	}

	for _, param := range params.ParamList {
		if err := param.Serialize(w); err != nil {
			return err
		}
	}
	return nil
}

func (params *Params) Deserialize(r io.Reader) error {
	version, err := serialization.ReadByte(r)
	if err != nil {
		return err
	}
	params.Version = version

	paramNum, err := serialization.ReadVarUint(r, 0)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode,
			"[Param Config] Deserialize param num error!")
	}

	var i uint64
	for i = 0; i < paramNum; i++ {
		param := new(Param)
		if err := param.Deserialize(r); err != nil {
			return err
		}
		params.ParamList = append(params.ParamList, param)
	}
	return nil
}
