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

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/errors"
)

const (
	NEOVM_TYPE  byte = 1
	WASMVM_TYPE byte = 3
)

// DeployCode is an implementation of transaction payload for deploy smartcontract
type DeployCode struct {
	Code []byte
	//modify for define contract type
	VmType      byte
	Name        string
	Version     string
	Author      string
	Email       string
	Description string

	address common.Address
}

func (dc *DeployCode) Address() common.Address {
	if dc.address == common.ADDRESS_EMPTY {
		dc.address = common.AddressFromVmCode(dc.Code)
	}
	return dc.address
}

func (dc *DeployCode) Serialize(w io.Writer) error {
	var err error

	err = serialization.WriteVarBytes(w, dc.Code)
	if err != nil {
		return fmt.Errorf("DeployCode Code Serialize failed: %s", err)
	}

	err = serialization.WriteByte(w, dc.VmType)
	if err != nil {
		return fmt.Errorf("DeployCode NeedStorage Serialize failed: %s", err)
	}

	err = serialization.WriteString(w, dc.Name)
	if err != nil {
		return fmt.Errorf("DeployCode Name Serialize failed: %s", err)
	}

	err = serialization.WriteString(w, dc.Version)
	if err != nil {
		return fmt.Errorf("DeployCode Version Serialize failed: %s", err)
	}

	err = serialization.WriteString(w, dc.Author)
	if err != nil {
		return fmt.Errorf("DeployCode Author Serialize failed: %s", err)
	}

	err = serialization.WriteString(w, dc.Email)
	if err != nil {
		return fmt.Errorf("DeployCode Email Serialize failed: %s", err)
	}

	err = serialization.WriteString(w, dc.Description)
	if err != nil {
		return fmt.Errorf("DeployCode Description Serialize failed: %s", err)
	}

	return nil
}

func (dc *DeployCode) Deserialize(r io.Reader) error {
	code, err := serialization.ReadVarBytes(r)
	if err != nil {
		return fmt.Errorf("DeployCode Code Deserialize failed: %s", err)
	}
	dc.Code = code

	dc.VmType, err = serialization.ReadByte(r)
	if err != nil {
		return fmt.Errorf("DeployCode NeedStorage Deserialize failed: %s", err)
	}

	dc.Name, err = serialization.ReadString(r)
	if err != nil {
		return fmt.Errorf("DeployCode Name Deserialize failed: %s", err)
	}

	dc.Version, err = serialization.ReadString(r)
	if err != nil {
		return fmt.Errorf("DeployCode CodeVersion Deserialize failed: %s", err)
	}

	dc.Author, err = serialization.ReadString(r)
	if err != nil {
		return fmt.Errorf("DeployCode Author Deserialize failed: %s", err)
	}

	dc.Email, err = serialization.ReadString(r)
	if err != nil {
		return fmt.Errorf("DeployCode Email Deserialize failed: %s", err)
	}

	dc.Description, err = serialization.ReadString(r)
	if err != nil {
		return fmt.Errorf("DeployCode Description Deserialize failed: %s", err)
	}

	err = validateDeployCode(dc)
	if err != nil {
		return err
	}

	return nil
}

func (dc *DeployCode) ToArray() []byte {
	b := new(bytes.Buffer)
	dc.Serialize(b)
	return b.Bytes()
}

func (dc *DeployCode) Serialization(sink *common.ZeroCopySink) error {
	sink.WriteVarBytes(dc.Code)
	sink.WriteByte(dc.VmType)
	sink.WriteString(dc.Name)
	sink.WriteString(dc.Version)
	sink.WriteString(dc.Author)
	sink.WriteString(dc.Email)
	sink.WriteString(dc.Description)

	return nil
}

//note: DeployCode.Code has data reference of param source
func (dc *DeployCode) Deserialization(source *common.ZeroCopySource) error {
	var eof, irregular bool
	dc.Code, _, irregular, eof = source.NextVarBytes()
	if irregular {
		return common.ErrIrregularData
	}

	dc.VmType, eof = source.NextByte()

	dc.Name, _, irregular, eof = source.NextString()
	if irregular {
		return common.ErrIrregularData
	}

	dc.Version, _, irregular, eof = source.NextString()
	if irregular {
		return common.ErrIrregularData
	}

	dc.Author, _, irregular, eof = source.NextString()
	if irregular {
		return common.ErrIrregularData
	}

	dc.Email, _, irregular, eof = source.NextString()
	if irregular {
		return common.ErrIrregularData
	}

	dc.Description, _, irregular, eof = source.NextString()
	if irregular {
		return common.ErrIrregularData
	}

	if eof {
		return io.ErrUnexpectedEOF
	}

	err := validateDeployCode(dc)
	if err != nil {
		return err
	}

	return nil
}

func validateDeployCode(dep *DeployCode) error {
	if len(dep.Code) > 1024*1024 {
		return errors.NewErr("[Contract] Code too long!")
	}

	if len(dep.Name) > 252 {
		return errors.NewErr("[Contract] name too long!")
	}

	if len(dep.Version) > 252 {
		return errors.NewErr("[Contract] version too long!")
	}

	if len(dep.Author) > 252 {
		return errors.NewErr("[author] version too long!")
	}

	if len(dep.Email) > 252 {
		return errors.NewErr("[author] emailPtr too long!")
	}

	if len(dep.Description) > 65536 {
		return errors.NewErr("[descPtr] emailPtr too long!")
	}

	if dep.VmType != WASMVM_TYPE && dep.VmType != NEOVM_TYPE {
		return errors.NewErr("[descPtr] VmType invalid!")
	}
	return nil
}

func CreateDeployCode(code []byte,
	vmType uint32,
	name []byte,
	version []byte,
	author []byte,
	email []byte,
	desc []byte) (*DeployCode, error) {

	contract := &DeployCode{
		Code:        code,
		VmType:      byte(vmType),
		Name:        string(name),
		Version:     string(version),
		Author:      string(author),
		Email:       string(email),
		Description: string(desc),
	}

	err := validateDeployCode(contract)
	if err != nil {
		return nil, err
	}
	return contract, nil
}
