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
	"fmt"
	"io"
	"math"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

/* **********************************************   */
type InitContractAdminParam struct {
	AdminOntID []byte
}

func (this *InitContractAdminParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes(this.AdminOntID)
}

func (this *InitContractAdminParam) Deserialization(source *common.ZeroCopySource) error {
	var irregular, eof bool
	this.AdminOntID, _, irregular, eof = source.NextVarBytes()
	if irregular {
		return common.ErrIrregularData
	}
	if eof {
		return io.ErrUnexpectedEOF
	}
	return nil
}

/* **********************************************   */
type TransferParam struct {
	ContractAddr  common.Address
	NewAdminOntID []byte
	KeyNo         uint64
}

func (this *TransferParam) Serialization(sink *common.ZeroCopySink) {
	serializeAddress(sink, this.ContractAddr)
	sink.WriteVarBytes(this.NewAdminOntID)
	utils.EncodeVarUint(sink, this.KeyNo)
}

func (this *TransferParam) Deserialization(source *common.ZeroCopySource) error {
	var err error
	if this.ContractAddr, err = utils.DecodeAddress(source); err != nil {
		return err
	}
	var irregular, eof bool
	if this.NewAdminOntID, _, irregular, eof = source.NextVarBytes(); irregular || eof {
		return fmt.Errorf("irregular:%v, eof:%v", irregular, eof)
	}
	if this.KeyNo, err = utils.DecodeVarUint(source); err != nil {
		return err
	}
	return nil
}

/* **********************************************   */
type FuncsToRoleParam struct {
	ContractAddr common.Address
	AdminOntID   []byte
	Role         []byte
	FuncNames    []string
	KeyNo        uint64
}

func (this *FuncsToRoleParam) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeAddress(sink, this.ContractAddr)
	sink.WriteVarBytes(this.AdminOntID)
	sink.WriteVarBytes(this.Role)
	utils.EncodeVarUint(sink, uint64(len(this.FuncNames)))
	for _, fn := range this.FuncNames {
		sink.WriteString(fn)
	}
	utils.EncodeVarUint(sink, this.KeyNo)
}

func (this *FuncsToRoleParam) Deserialization(source *common.ZeroCopySource) error {
	var err error
	var fnLen uint64
	var i uint64

	if this.ContractAddr, err = utils.DecodeAddress(source); err != nil {
		return err
	}
	if this.AdminOntID, err = utils.DecodeVarBytes(source); err != nil {
		return fmt.Errorf("AdminOntID Deserialization error: %s", err)
	}
	if this.Role, err = utils.DecodeVarBytes(source); err != nil {
		return fmt.Errorf("Role Deserialization error: %s", err)
	}
	if fnLen, err = utils.DecodeVarUint(source); err != nil {
		return err
	}
	this.FuncNames = make([]string, 0)
	for i = 0; i < fnLen; i++ {
		fn, err := utils.DecodeString(source)
		if err != nil {
			return fmt.Errorf("FuncNames Deserialization error: %s", err)
		}
		this.FuncNames = append(this.FuncNames, fn)
	}
	if this.KeyNo, err = utils.DecodeVarUint(source); err != nil {
		return err
	}
	return nil
}

type OntIDsToRoleParam struct {
	ContractAddr common.Address
	AdminOntID   []byte
	Role         []byte
	Persons      [][]byte
	KeyNo        uint64
}

func (this *OntIDsToRoleParam) Serialization(sink *common.ZeroCopySink) {
	serializeAddress(sink, this.ContractAddr)
	sink.WriteVarBytes(this.AdminOntID)
	sink.WriteVarBytes(this.Role)

	utils.EncodeVarUint(sink, uint64(len(this.Persons)))
	for _, p := range this.Persons {
		sink.WriteVarBytes(p)
	}
	utils.EncodeVarUint(sink, this.KeyNo)
}

func (this *OntIDsToRoleParam) Deserialization(source *common.ZeroCopySource) error {
	var err error
	var pLen uint64
	if this.ContractAddr, err = utils.DecodeAddress(source); err != nil {
		return err
	}
	if this.AdminOntID, err = utils.DecodeVarBytes(source); err != nil {
		return fmt.Errorf("AdminOntID Deserialization error: %s", err)
	}
	if this.Role, err = utils.DecodeVarBytes(source); err != nil {
		return fmt.Errorf("Role Deserialization error: %s", err)
	}
	if pLen, err = utils.DecodeVarUint(source); err != nil {
		return err
	}
	this.Persons = make([][]byte, 0)
	for i := uint64(0); i < pLen; i++ {
		p, err := utils.DecodeVarBytes(source)
		if err != nil {
			return fmt.Errorf("Persons Deserialization error: %s", err)
		}
		this.Persons = append(this.Persons, p)
	}
	if this.KeyNo, err = utils.DecodeVarUint(source); err != nil {
		return err
	}
	return nil
}

type DelegateParam struct {
	ContractAddr common.Address
	From         []byte
	To           []byte
	Role         []byte
	Period       uint64
	Level        uint64
	KeyNo        uint64
}

func (this *DelegateParam) Serialization(sink *common.ZeroCopySink) {
	serializeAddress(sink, this.ContractAddr)
	sink.WriteVarBytes(this.From)
	sink.WriteVarBytes(this.To)
	sink.WriteVarBytes(this.Role)
	utils.EncodeVarUint(sink, this.Period)
	utils.EncodeVarUint(sink, uint64(this.Level))
	utils.EncodeVarUint(sink, this.KeyNo)
}

func (this *DelegateParam) Deserialization(source *common.ZeroCopySource) error {
	var err error
	var level uint64
	if this.ContractAddr, err = utils.DecodeAddress(source); err != nil {
		return err
	}
	if this.From, err = utils.DecodeVarBytes(source); err != nil {
		return fmt.Errorf("From Deserialization error: %s", err)
	}
	if this.To, err = utils.DecodeVarBytes(source); err != nil {
		return fmt.Errorf("To Deserialization error: %s", err)
	}
	if this.Role, err = utils.DecodeVarBytes(source); err != nil {
		return fmt.Errorf("Role Deserialization error: %s", err)
	}
	if this.Period, err = utils.DecodeVarUint(source); err != nil {
		return err
	}
	if level, err = utils.DecodeVarUint(source); err != nil {
		return err
	}
	if this.KeyNo, err = utils.DecodeVarUint(source); err != nil {
		return err
	}
	if level > math.MaxInt8 || this.Period > math.MaxUint32 {
		return fmt.Errorf("period or level too large: (%d, %d)", this.Period, level)
	}
	this.Level = level
	return nil
}

type WithdrawParam struct {
	ContractAddr common.Address
	Initiator    []byte
	Delegate     []byte
	Role         []byte
	KeyNo        uint64
}

func (this *WithdrawParam) Serialization(sink *common.ZeroCopySink) {
	serializeAddress(sink, this.ContractAddr)
	sink.WriteVarBytes(this.Initiator)
	sink.WriteVarBytes(this.Delegate)
	sink.WriteVarBytes(this.Role)
	utils.EncodeVarUint(sink, this.KeyNo)
}
func (this *WithdrawParam) Deserialization(source *common.ZeroCopySource) error {
	var err error
	if this.ContractAddr, err = utils.DecodeAddress(source); err != nil {
		return err
	}
	if this.Initiator, err = utils.DecodeVarBytes(source); err != nil {
		return fmt.Errorf("Initiator Deserialization error: %s", err)
	}
	if this.Delegate, err = utils.DecodeVarBytes(source); err != nil {
		return fmt.Errorf("Delegate Deserialization error: %s", err)
	}
	if this.Role, err = utils.DecodeVarBytes(source); err != nil {
		return fmt.Errorf("Role Deserialization error: %s", err)
	}
	if this.KeyNo, err = utils.DecodeVarUint(source); err != nil {
		return err
	}
	return nil
}

type VerifyTokenParam struct {
	ContractAddr common.Address
	Caller       []byte
	Fn           string
	KeyNo        uint64
}

func (this *VerifyTokenParam) Serialization(sink *common.ZeroCopySink) {
	serializeAddress(sink, this.ContractAddr)
	sink.WriteVarBytes(this.Caller)
	sink.WriteString(this.Fn)
	utils.EncodeVarUint(sink, this.KeyNo)
}

func (this *VerifyTokenParam) Deserialization(source *common.ZeroCopySource) error {
	var err error
	if this.ContractAddr, err = utils.DecodeAddress(source); err != nil {
		return err
	}
	if this.Caller, err = utils.DecodeVarBytes(source); err != nil {
		return fmt.Errorf("Caller Deserialization error: %s", err)
	}
	if this.Fn, err = utils.DecodeString(source); err != nil {
		return fmt.Errorf("Fn Deserialization error: %s", err)
	}
	if this.KeyNo, err = utils.DecodeVarUint(source); err != nil {
		return err
	}
	return nil
}
