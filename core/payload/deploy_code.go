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

package payload

import (
	"github.com/Ontology/common/serialization"
	. "github.com/Ontology/errors"
	"io"
	"bytes"
	vmtypes "github.com/Ontology/vm/types"
)

type DeployCode struct {
	VmType      vmtypes.VmType
	Code        []byte
	NeedStorage bool
	Name        string
	Version     string
	Author      string
	Email       string
	Description string
}

func (dc *DeployCode) Serialize(w io.Writer) error {
	var err error
	err = serialization.WriteByte(w, byte(dc.VmType))
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "DeployCode VmType Serialize failed.")
	}

	err = serialization.WriteVarBytes(w, dc.Code)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "DeployCode Code Serialize failed.")
	}

	err = serialization.WriteBool(w, dc.NeedStorage)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "DeployCode NeedStorage Serialize failed.")
	}

	err = serialization.WriteVarString(w, dc.Name)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "DeployCode Name Serialize failed.")
	}

	err = serialization.WriteVarString(w, dc.Version)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "DeployCode Version Serialize failed.")
	}

	err = serialization.WriteVarString(w, dc.Author)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "DeployCode Author Serialize failed.")
	}

	err = serialization.WriteVarString(w, dc.Email)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "DeployCode Email Serialize failed.")
	}

	err = serialization.WriteVarString(w, dc.Description)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "DeployCode Description Serialize failed.")
	}

	return nil
}

func (dc *DeployCode) Deserialize(r io.Reader) error {
	vmType, err := serialization.ReadByte(r)
	if err != nil {
		return err
	}
	dc.VmType = vmtypes.VmType(vmType)

	dc.Code, err = serialization.ReadVarBytes(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "DeployCode Code Deserialize failed.")
	}

	dc.NeedStorage, err = serialization.ReadBool(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "DeployCode NeedStorage Deserialize failed.")
	}

	dc.Name, err = serialization.ReadVarString(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "DeployCode Name Deserialize failed.")
	}

	dc.Version, err = serialization.ReadVarString(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "DeployCode CodeVersion Deserialize failed.")
	}

	dc.Author, err = serialization.ReadVarString(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "DeployCode Author Deserialize failed.")
	}

	dc.Email, err = serialization.ReadVarString(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "DeployCode Email Deserialize failed.")
	}

	dc.Description, err = serialization.ReadVarString(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "DeployCode Description Deserialize failed.")
	}

	return nil
}

func (dc *DeployCode) ToArray() []byte {
	b := new(bytes.Buffer)
	dc.Serialize(b)
	return b.Bytes()
}
