/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package global_params

import (
	"io"

	"encoding/json"

	"fmt"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/errors"
)

type Params map[string]string

type Admin common.Address

type ParamNameList []string

func (params *Params) Serialize(w io.Writer) error {
	paramsJsonString, err := json.Marshal(params)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "param config, serialize params error!")
	}
	if err := serialization.WriteVarBytes(w, paramsJsonString); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "param config, serialize params error!")
	}
	return nil
}

func (params *Params) Deserialize(r io.Reader) error {
	paramsJsonString, err := serialization.ReadVarBytes(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "param config, deserialize params error!")
	}
	err = json.Unmarshal(paramsJsonString, params)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "param config, deserialize params error!")
	}
	return nil
}

func (admin *Admin) Serialize(w io.Writer) error {
	_, err := w.Write(admin[:])
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "param config, serialize admin error!")
	}
	return nil
}

func (admin *Admin) Deserialize(r io.Reader) error {
	n, err := r.Read(admin[:])
	if n != len(admin[:]) || err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "param config, deserialize admin error!")
	}
	return nil
}

func (nameList *ParamNameList) Serialize(w io.Writer) error {
	nameNum := len(*nameList)
	if err := serialization.WriteVarUint(w, uint64(nameNum)); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "param config, serialize param name list length error!")
	}
	for _, value := range *nameList {
		if err := serialization.WriteString(w, value); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, fmt.Sprintf("param config, serialize param name %v error!", value))
		}
	}
	return nil
}

func (nameList *ParamNameList) Deserialize(r io.Reader) error {
	nameNum, err := serialization.ReadVarUint(r, 0)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "param config, deserialize param name list length error!")
	}
	for i := 0; uint64(i) < nameNum; i++ {
		name, err := serialization.ReadString(r)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, fmt.Sprintf("param config, deserialize param name %v error!", name))
		}
		*nameList = append(*nameList, name)
	}
	return nil
}
