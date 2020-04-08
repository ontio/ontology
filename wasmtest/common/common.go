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
package common

import (
	"bytes"
	"encoding/json"

	utils2 "github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/utils"
	common2 "github.com/ontio/ontology/http/base/common"
	"github.com/ontio/ontology/smartcontract/states"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/vm/neovm"
)

type TestEnv struct {
	Witness []common.Address `json:"witness"`
}

func (self TestEnv) MarshalJSON() ([]byte, error) {
	var witness []string
	for _, addr := range self.Witness {
		witness = append(witness, addr.ToBase58())
	}
	env := struct {
		Witness []string `json:"witness"`
	}{Witness: witness}

	return json.Marshal(env)
}

func (self *TestEnv) UnmarshalJSON(buf []byte) error {
	env := struct {
		Witness []string `json:"witness"`
	}{}
	err := json.Unmarshal(buf, &env)
	if err != nil {
		return err
	}
	var witness []common.Address
	for _, addr := range env.Witness {
		wit, err := common.AddressFromBase58(addr)
		if err != nil {
			return err
		}

		witness = append(witness, wit)
	}

	self.Witness = witness
	return nil
}

type TestCase struct {
	Env         TestEnv `json:"env"`
	NeedContext bool    `json:"needcontext"`
	Method      string  `json:"method"`
	Param       string  `json:"param"`
	Expect      string  `json:"expected"`
	Notify      string  `json:"notify"`
}

type ConAddr struct {
	File    string
	Address common.Address
}

type TestContext struct {
	Admin   common.Address
	AddrMap []ConAddr
}

func GenWasmTransaction(testCase TestCase, contract common.Address, testConext *TestContext) (*types.Transaction, error) {
	params, err := utils2.ParseParams(testCase.Param)
	if err != nil {
		return nil, err
	}
	allParam := append([]interface{}{}, testCase.Method)
	allParam = append(allParam, params...)
	tx, err := utils.NewWasmVMInvokeTransaction(0, 100000000, contract, allParam)
	if err != nil {
		return nil, err
	}

	if testCase.NeedContext {
		source := common.NewZeroCopySource(tx.Payload.(*payload.InvokeCode).Code)
		contract := &states.WasmContractParam{}
		err := contract.Deserialization(source)
		if err != nil {
			return nil, err
		}
		contextParam := buildTestConext(testConext)
		contract.Args = append(contract.Args, contextParam...)

		tx.Payload.(*payload.InvokeCode).Code = common.SerializeToBytes(contract)
	}

	imt, err := tx.IntoImmutable()
	if err != nil {
		return nil, err
	}

	imt.SignedAddr = append(imt.SignedAddr, testCase.Env.Witness...)
	imt.SignedAddr = append(imt.SignedAddr, testConext.Admin)

	return imt, nil
}

// when need pass testConext to neovm contract, must write contract as def Main(operation, args) api. and args need be a list.
func buildTestConextForNeo(testConext *TestContext) []byte {
	addrMap := testConext.AddrMap
	builder := neovm.NewParamsBuilder(new(bytes.Buffer))

	// [args, operation]
	builder.Emit(neovm.SWAP)
	// [operation, args]
	builder.Emit(neovm.TOALTSTACK)
	// [operation]

	// construct [admin, map] array
	builder.EmitPushByteArray(testConext.Admin[:])
	builder.Emit(neovm.NEWMAP)
	for _, item := range addrMap {
		file := item.File
		addr := item.Address
		builder.Emit(neovm.DUP)
		builder.EmitPushByteArray(addr[:])
		builder.Emit(neovm.SWAP)
		builder.EmitPushByteArray([]byte(file))
		builder.Emit(neovm.ROT)
		builder.Emit(neovm.SETITEM)
	}
	builder.Emit(neovm.PUSH2)
	builder.Emit(neovm.PACK)
	// end [addmin, map] array construct

	// [operation, [admin, map]]
	builder.Emit(neovm.FROMALTSTACK)
	// [operation, [admin, map], args]
	builder.Emit(neovm.UNPACK)
	builder.Emit(neovm.PUSH1)
	builder.Emit(neovm.ADD)
	builder.Emit(neovm.PACK)
	// [operation, [args,[admin, map]]]
	builder.Emit(neovm.SWAP)
	// the second list of last elt is the testConext
	// [[args,[admin, map]], operation] ==> topof the stack.
	return builder.ToArray()
}

func GenNeoVMTransaction(testCase TestCase, contract common.Address, testConext *TestContext) (*types.Transaction, error) {
	params, err := utils2.ParseParams(testCase.Param)
	if err != nil {
		return nil, err
	}
	allParam := append([]interface{}{}, testCase.Method)
	allParam = append(allParam, params...)
	tx, err := common2.NewNeovmInvokeTransaction(0, 100000000, contract, allParam)
	if err != nil {
		return nil, err
	}

	if testCase.NeedContext {
		args := buildTestConextForNeo(testConext)
		codelen := uint32(len(tx.Payload.(*payload.InvokeCode).Code))
		tx.Payload.(*payload.InvokeCode).Code = append(tx.Payload.(*payload.InvokeCode).Code[:codelen-(common.ADDR_LEN+1)], args...)
		tx.Payload.(*payload.InvokeCode).Code = append(tx.Payload.(*payload.InvokeCode).Code, 0x67)
		tx.Payload.(*payload.InvokeCode).Code = append(tx.Payload.(*payload.InvokeCode).Code, contract[:]...)
		//neovms.Dumpcode(tx.Payload.(*payload.InvokeCode).Code[:], "")
	}

	imt, err := tx.IntoImmutable()
	if err != nil {
		return nil, err
	}

	imt.SignedAddr = append(imt.SignedAddr, testCase.Env.Witness...)
	imt.SignedAddr = append(imt.SignedAddr, testConext.Admin)

	return imt, nil
}

func buildTestConext(testConext *TestContext) []byte {
	bf := common.NewZeroCopySink(nil)
	addrMap := testConext.AddrMap

	bf.WriteAddress(testConext.Admin)
	bf.WriteVarUint(uint64(len(addrMap)))
	for _, item := range addrMap {
		file := item.File
		addr := item.Address
		bf.WriteString(file)
		bf.WriteAddress(addr)
	}

	return bf.Bytes()
}
