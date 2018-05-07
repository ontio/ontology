package states

import (
	"encoding/json"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/errors"
	"io"
)

type Params map[string]string

type Admin common.Address

func (params *Params) Serialize(w io.Writer) error {
	paramsJsonString, err := json.Marshal(params)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Param Config] Serialize params error!")
	}
	if err := serialization.WriteVarBytes(w, paramsJsonString); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Param Config] Serialize params error!")
	}
	return nil
}

func (params *Params) Deserialize(r io.Reader) error {
	paramsJsonString, err := serialization.ReadVarBytes(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Param Config] Deserialize params error!")
	}
	err = json.Unmarshal(paramsJsonString, params)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Param Config] Deserialize params error!")
	}
	return nil
}

func (admin *Admin) Serialize(w io.Writer) error {
	_, err := w.Write(admin[:])
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Param Config] Serialize admin error!")
	}
	return nil
}

func (admin *Admin)Deserialize(r io.Reader) error {
	n, err := r.Read(admin[:])
	if n != len(admin[:]) || err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Param Config] Deserialize params error!")
	}
	return nil
}