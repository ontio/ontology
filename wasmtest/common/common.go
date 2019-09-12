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
	"encoding/json"
	utils2 "github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/utils"
	"github.com/ontio/ontology/smartcontract/states"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
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
	Env     TestEnv `json:"env"`
	NeedEnv bool    `json:"needenv"`
	Method  string  `json:"method"`
	Param   string  `json:"param"`
	Expect  string  `json:"expected"`
}

type TestContext struct {
	Admin   common.Address
	AddrMap map[string]common.Address
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

	if testCase.NeedEnv {
		source := common.NewZeroCopySource(tx.Payload.(*payload.InvokeCode).Code)
		contract := &states.WasmContractParam{}
		err := contract.Deserialization(source)
		if err != nil {
			return nil, err
		}
		contextParam := buildTestConext(testConext)
		contract.Args = append(contract.Args, contextParam...)

		sink := common.NewZeroCopySink(nil)
		contract.Serialization(sink)

		tx.Payload.(*payload.InvokeCode).Code = sink.Bytes()
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
	for file, addr := range addrMap {
		bf.WriteString(file)
		bf.WriteAddress(addr)
	}

	return bf.Bytes()
}
