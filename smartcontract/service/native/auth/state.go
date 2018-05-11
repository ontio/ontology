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

package auth

import (
	"bytes"
	"fmt"
	"github.com/ontio/ontology/common/serialization"
	"io"
	"math"
)

/* **********************************************   */
type InitContractAdminParam struct {
	AdminOntID []byte
}

func (this *InitContractAdminParam) Serialize(w io.Writer) error {
	if err := serialization.WriteVarBytes(w, this.AdminOntID); err != nil {
		return err
	}
	return nil
}

func (this *InitContractAdminParam) Deserialize(rd io.Reader) error {
	var err error
	if this.AdminOntID, err = serialization.ReadVarBytes(rd); err != nil {
		return err
	}
	return nil
}

/* **********************************************   */
type TransferParam struct {
	ContractAddr  []byte
	NewAdminOntID []byte
	KeyNo         uint32
}

func (this *TransferParam) Serialize(w io.Writer) error {
	if err := serialization.WriteVarBytes(w, this.ContractAddr); err != nil {
		return err
	}
	if err := serialization.WriteVarBytes(w, this.NewAdminOntID); err != nil {
		return err
	}
	if err := serialization.WriteUint32(w, this.KeyNo); err != nil {
		return nil
	}
	return nil
}

func (this *TransferParam) Deserialize(rd io.Reader) error {
	var err error
	if this.ContractAddr, err = serialization.ReadVarBytes(rd); err != nil {
		return err
	}
	if this.NewAdminOntID, err = serialization.ReadVarBytes(rd); err != nil {
		return err
	}
	if this.KeyNo, err = serialization.ReadUint32(rd); err != nil {
		return err
	}
	return nil
}

/* **********************************************   */
type FuncsToRoleParam struct {
	ContractAddr []byte
	AdminOntID   []byte
	Role         []byte
	FuncNames    []string
	KeyNo        uint32
}

func (this *FuncsToRoleParam) Serialize(w io.Writer) error {
	if err := serialization.WriteVarBytes(w, this.ContractAddr); err != nil {
		return err
	}
	if err := serialization.WriteVarBytes(w, this.AdminOntID); err != nil {
		return err
	}
	if err := serialization.WriteVarBytes(w, this.Role); err != nil {
		return err
	}
	if err := serialization.WriteVarUint(w, uint64(len(this.FuncNames))); err != nil {
		return err
	}
	for _, fn := range this.FuncNames {
		if err := serialization.WriteString(w, fn); err != nil {
			return err
		}
	}
	if err := serialization.WriteUint32(w, this.KeyNo); err != nil {
		return nil
	}
	return nil
}

func (this *FuncsToRoleParam) Deserialize(rd io.Reader) error {
	var err error
	var fnLen uint64
	var i uint64

	if this.ContractAddr, err = serialization.ReadVarBytes(rd); err != nil {
		return err
	}
	if this.AdminOntID, err = serialization.ReadVarBytes(rd); err != nil {
		return err
	}
	if this.Role, err = serialization.ReadVarBytes(rd); err != nil {
		return err
	}
	if fnLen, err = serialization.ReadVarUint(rd, 0); err != nil {
		return err
	}
	this.FuncNames = make([]string, fnLen)
	for i = 0; i < fnLen; i++ {
		fn, err := serialization.ReadString(rd)
		if err != nil {
			return err
		}
		this.FuncNames[i] = fn
	}
	if this.KeyNo, err = serialization.ReadUint32(rd); err != nil {
		return err
	}
	return nil
}

type OntIDsToRoleParam struct {
	ContractAddr []byte
	AdminOntID   []byte
	Role         []byte
	Persons      [][]byte
	KeyNo        uint32
}

func (this *OntIDsToRoleParam) Serialize(w io.Writer) error {
	if err := serialization.WriteVarBytes(w, this.ContractAddr); err != nil {
		return err
	}
	if err := serialization.WriteVarBytes(w, this.AdminOntID); err != nil {
		return err
	}
	if err := serialization.WriteVarBytes(w, this.Role); err != nil {
		return err
	}
	if err := serialization.WriteVarUint(w, uint64(len(this.Persons))); err != nil {
		return err
	}
	for _, p := range this.Persons {
		if err := serialization.WriteVarBytes(w, p); err != nil {
			return err
		}
	}
	if err := serialization.WriteUint32(w, this.KeyNo); err != nil {
		return nil
	}
	return nil
}

func (this *OntIDsToRoleParam) Deserialize(rd io.Reader) error {
	var err error
	var pLen uint64
	if this.ContractAddr, err = serialization.ReadVarBytes(rd); err != nil {
		return err
	}
	if this.AdminOntID, err = serialization.ReadVarBytes(rd); err != nil {
		return err
	}
	if this.Role, err = serialization.ReadVarBytes(rd); err != nil {
		return err
	}
	if pLen, err = serialization.ReadVarUint(rd, 0); err != nil {
		return err
	}
	this.Persons = make([][]byte, pLen)
	for i := uint64(0); i < pLen; i++ {
		p, err := serialization.ReadVarBytes(rd)
		if err != nil {
			return err
		}
		this.Persons[i] = p
	}
	if this.KeyNo, err = serialization.ReadUint32(rd); err != nil {
		return err
	}
	return nil
}

type DelegateParam struct {
	ContractAddr []byte
	From         []byte
	To           []byte
	Role         []byte
	Period       uint32
	Level        uint
	KeyNo        uint32
}

func (this *DelegateParam) Serialize(w io.Writer) error {
	if err := serialization.WriteVarBytes(w, this.ContractAddr); err != nil {
		return err
	}
	if err := serialization.WriteVarBytes(w, this.From); err != nil {
		return err
	}
	if err := serialization.WriteVarBytes(w, this.To); err != nil {
		return err
	}
	if err := serialization.WriteVarBytes(w, this.Role); err != nil {
		return err
	}
	if err := serialization.WriteUint32(w, this.Period); err != nil {
		return err
	}
	if err := serialization.WriteVarUint(w, uint64(this.Level)); err != nil {
		return err
	}
	if err := serialization.WriteUint32(w, this.KeyNo); err != nil {
		return err
	}
	return nil
}

func (this *DelegateParam) Deserialize(rd io.Reader) error {
	var err error
	var period uint32
	var level uint64
	if this.ContractAddr, err = serialization.ReadVarBytes(rd); err != nil {
		return err
	}
	if this.From, err = serialization.ReadVarBytes(rd); err != nil {
		return err
	}
	if this.To, err = serialization.ReadVarBytes(rd); err != nil {
		return err
	}
	if this.Role, err = serialization.ReadVarBytes(rd); err != nil {
		return err
	}
	if this.Period, err = serialization.ReadUint32(rd); err != nil {
		return err
	}
	if level, err = serialization.ReadVarUint(rd, 0); err != nil {
		return err
	}
	if this.KeyNo, err = serialization.ReadUint32(rd); err != nil {
		return err
	}
	if level > math.MaxInt8 {
		return fmt.Errorf("period or level too large: (%d, %d)", period, level)
	}
	this.Level = uint(level)
	return nil
}

type WithdrawParam struct {
	ContractAddr []byte
	Initiator    []byte
	Delegate     []byte
	Role         []byte
	KeyNo        uint32
}

func (this *WithdrawParam) Serialize(w io.Writer) error {
	if err := serialization.WriteVarBytes(w, this.ContractAddr); err != nil {
		return err
	}
	if err := serialization.WriteVarBytes(w, this.Initiator); err != nil {
		return err
	}
	if err := serialization.WriteVarBytes(w, this.Delegate); err != nil {
		return err
	}
	if err := serialization.WriteVarBytes(w, this.Role); err != nil {
		return err
	}
	if err := serialization.WriteUint32(w, this.KeyNo); err != nil {
		return err
	}
	return nil
}
func (this *WithdrawParam) Deserialize(rd io.Reader) error {
	var err error
	if this.ContractAddr, err = serialization.ReadVarBytes(rd); err != nil {
		return err
	}
	if this.Initiator, err = serialization.ReadVarBytes(rd); err != nil {
		return err
	}
	if this.Delegate, err = serialization.ReadVarBytes(rd); err != nil {
		return err
	}
	if this.Role, err = serialization.ReadVarBytes(rd); err != nil {
		return err
	}
	if this.KeyNo, err = serialization.ReadUint32(rd); err != nil {
		return err
	}
	return nil
}

type VerifyTokenParam struct {
	ContractAddr []byte
	Caller       []byte
	Fn           []byte
	KeyNo        uint32
}

func (this *VerifyTokenParam) Serialize(w io.Writer) error {
	if err := serialization.WriteVarBytes(w, this.ContractAddr); err != nil {
		return err
	}
	if err := serialization.WriteVarBytes(w, this.Caller); err != nil {
		return err
	}
	if err := serialization.WriteVarBytes(w, this.Fn); err != nil {
		return err
	}
	if err := serialization.WriteUint32(w, this.KeyNo); err != nil {
		return err
	}
	return nil
}

func (this *VerifyTokenParam) Deserialize(rd io.Reader) error {
	var err error
	if this.ContractAddr, err = serialization.ReadVarBytes(rd); err != nil {
		return err
	}
	if this.Caller, err = serialization.ReadVarBytes(rd); err != nil {
		return err //deserialize caller error
	}
	if this.Fn, err = serialization.ReadVarBytes(rd); err != nil {
		return err
	}
	if this.KeyNo, err = serialization.ReadUint32(rd); err != nil {
		return err
	}
	return nil
}

type AuthToken struct {
	expireTime uint32
	level      uint
}

func (this *AuthToken) Serialize(w io.Writer) error {
	//bf := new(bytes.Buffer)
	if err := serialization.WriteVarUint(w, uint64(this.expireTime)); err != nil {
		return err
	}
	if err := serialization.WriteVarUint(w, uint64(this.level)); err != nil {
		return err
	}
	return nil
}

func (this *AuthToken) Deserialize(rd io.Reader) error {
	//rd := bytes.NewReader(data)
	expireTime, err := serialization.ReadVarUint(rd, 0)
	if err != nil {
		return err
	}
	level, err := serialization.ReadVarUint(rd, 0)
	if err != nil {
		return err
	}
	this.expireTime = uint32(expireTime)
	this.level = uint(level)
	return nil
}

func (this *AuthToken) serialize() ([]byte, error) {
	bf := new(bytes.Buffer)
	err := this.Serialize(bf)
	if err != nil {
		return nil, err
	}
	return bf.Bytes(), nil
}

func (this *AuthToken) deserialize(data []byte) error {
	rd := bytes.NewReader(data)
	return this.Deserialize(rd)
}
