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

package contract

import (
	"bytes"
	"io"

	"github.com/Ontology/common"
	"github.com/Ontology/common/serialization"
	vm "github.com/Ontology/vm/neovm"
)

//Contract address is the hash of contract program .
//which be used to control asset or indicate the smart contract address

//Contract include the program codes with parameters which can be executed on specific evnrioment
type Contract struct {
	//the contract program code,which will be run on VM or specific envrionment
	Code            []byte

	//the Contract Parameter type list
	// describe the number of contract program parameters and the parameter type
	Parameters      []ContractParameterType

	//The program hash as contract address
	ProgramHash common.Address

	//owner's pubkey hash indicate the owner of contract
	OwnerPubkeyHash common.Address
}

func (c *Contract) IsStandard() bool {
	if len(c.Code) != 35 {
		return false
	}
	if c.Code[0] != 33 || c.Code[34] != byte(vm.CHECKSIG) {
		return false
	}
	return true
}

func (c *Contract) IsMultiSigContract() bool {
	var m int16 = 0
	var n int16 = 0
	i := 0

	if len(c.Code) < 37 {
		return false
	}
	if c.Code[i] > byte(vm.PUSH16) {
		return false
	}
	if c.Code[i] < byte(vm.PUSH1) && c.Code[i] != 1 && c.Code[i] != 2 {
		return false
	}

	switch c.Code[i] {
	case 1:
		i++
		m = int16(c.Code[i])
		i++
		break
	case 2:
		i++
		m = common.BytesToInt16(c.Code[i:])
		i += 2
		break
	default:
		m = int16(c.Code[i]) - 80
		i++
		break
	}

	if m < 1 || m > 1024 {
		return false
	}

	for c.Code[i] == 33 {
		i += 34
		if len(c.Code) <= i {
			return false
		}
		n++
	}
	if n < m || n > 1024 {
		return false
	}

	switch c.Code[i] {
	case 1:
		i++
		if n != int16(c.Code[i]) {
			return false
		}
		i++
		break
	case 2:
		i++
		if n != common.BytesToInt16(c.Code[i:]) {
			return false
		}
		i += 2
		break
	default:
		if n != (int16(c.Code[i]) - 80) {
			return false
		}
		i++
		break
	}

	if c.Code[i] != byte(vm.CHECKMULTISIG) {
		return false
	}
	i++
	if len(c.Code) != i {
		return false
	}

	return true
}

func (c *Contract) GetType() ContractType {
	if c.IsStandard() {
		return SignatureContract
	}
	if c.IsMultiSigContract() {
		return MultiSigContract
	}
	return CustomContract
}

func (c *Contract) Deserialize(r io.Reader) error {
	c.OwnerPubkeyHash.Deserialize(r)

	p, err := serialization.ReadVarBytes(r)
	if err != nil {
		return err
	}
	c.Parameters = ByteToContractParameterType(p)

	c.Code, err = serialization.ReadVarBytes(r)
	if err != nil {
		return err
	}

	return nil
}

func (c *Contract) Serialize(w io.Writer) error {
	err := c.OwnerPubkeyHash.Serialize(w)
	if err != nil {
		return err
	}

	err = serialization.WriteVarBytes(w, ContractParameterTypeToByte(c.Parameters))
	if err != nil {
		return err
	}

	err = serialization.WriteVarBytes(w, c.Code)
	if err != nil {
		return err
	}

	return nil
}

func (c *Contract) ToArray() []byte {
	w := new(bytes.Buffer)
	c.Serialize(w)

	return w.Bytes()
}

func ContractParameterTypeToByte(c []ContractParameterType) []byte {
	b := make([]byte, len(c))

	for i := 0; i < len(c); i++ {
		b[i] = byte(c[i])
	}

	return b
}

func ByteToContractParameterType(b []byte) []ContractParameterType {
	c := make([]ContractParameterType, len(b))

	for i := 0; i < len(b); i++ {
		c[i] = ContractParameterType(b[i])
	}

	return c
}
