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
	"bytes"
	"fmt"
	"io"

	"github.com/ontio/ontology/common/serialization"
	vmtypes "github.com/ontio/ontology/vm/types"
)

type DeployCode struct {
	Code        *vmtypes.VmCode
	NeedStorage bool
	Name        string
	Version     string
	Author      string
	Email       string
	Description string
}

func (dc *DeployCode) Serialize(w io.Writer) error {
	var err error

	err = dc.Code.Serialize(w)
	if err != nil {
		return fmt.Errorf("DeployCode Code Serialize failed: %s", err)
	}

	err = serialization.WriteBool(w, dc.NeedStorage)
	if err != nil {
		return fmt.Errorf("DeployCode NeedStorage Serialize failed: %s", err)
	}

	err = serialization.WriteVarString(w, dc.Name)
	if err != nil {
		return fmt.Errorf("DeployCode Name Serialize failed: %s", err)
	}

	err = serialization.WriteVarString(w, dc.Version)
	if err != nil {
		return fmt.Errorf("DeployCode Version Serialize failed: %s", err)
	}

	err = serialization.WriteVarString(w, dc.Author)
	if err != nil {
		return fmt.Errorf("DeployCode Author Serialize failed: %s", err)
	}

	err = serialization.WriteVarString(w, dc.Email)
	if err != nil {
		return fmt.Errorf("DeployCode Email Serialize failed: %s", err)
	}

	err = serialization.WriteVarString(w, dc.Description)
	if err != nil {
		return fmt.Errorf("DeployCode Description Serialize failed: %s", err)
	}

	return nil
}

func (dc *DeployCode) Deserialize(r io.Reader) error {
	var err error
	code := new(vmtypes.VmCode)

	err = code.Deserialize(r)
	if err != nil {
		return fmt.Errorf("DeployCode Code Deserialize failed: %s", err)
	}
	dc.Code = code

	dc.NeedStorage, err = serialization.ReadBool(r)
	if err != nil {
		return fmt.Errorf("DeployCode NeedStorage Deserialize failed: %s", err)
	}

	dc.Name, err = serialization.ReadVarString(r)
	if err != nil {
		return fmt.Errorf("DeployCode Name Deserialize failed: %s", err)
	}

	dc.Version, err = serialization.ReadVarString(r)
	if err != nil {
		return fmt.Errorf("DeployCode CodeVersion Deserialize failed: %s", err)
	}

	dc.Author, err = serialization.ReadVarString(r)
	if err != nil {
		return fmt.Errorf("DeployCode Author Deserialize failed: %s", err)
	}

	dc.Email, err = serialization.ReadVarString(r)
	if err != nil {
		return fmt.Errorf("DeployCode Email Deserialize failed: %s", err)
	}

	dc.Description, err = serialization.ReadVarString(r)
	if err != nil {
		return fmt.Errorf("DeployCode Description Deserialize failed: %s", err)
	}

	return nil
}

func (dc *DeployCode) ToArray() []byte {
	b := new(bytes.Buffer)
	dc.Serialize(b)
	return b.Bytes()
}
