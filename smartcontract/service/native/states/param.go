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