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
package utils

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	svrneo "github.com/ontio/ontology/smartcontract/service/neovm"
	svrneovm "github.com/ontio/ontology/smartcontract/service/neovm"
	"github.com/ontio/ontology/vm/neovm"
	"strings"
)

type InvokeCodeInfo struct {
	Contract string
	Method   string
	Version  int64
	Params   []interface{}
}

var (
	nativeCallCode string
)

func init() {
	builder := neovm.NewParamsBuilder(new(bytes.Buffer))
	builder.Emit(neovm.SYSCALL)
	builder.EmitPushByteArray([]byte(svrneovm.NATIVE_INVOKE_NAME))
	nativeCallCode = hex.EncodeToString(builder.ToArray())
}

func ParseInvokeCode(code string) (*InvokeCodeInfo, error) {
	defer func() {
		err := recover()
		if err != nil {
			fmt.Printf("ParseInvokeCode recover error:%s\n", err)
		}
	}()
	if !strings.HasSuffix(code, nativeCallCode) {
		return nil, nil
	}
	code = strings.TrimSuffix(code, nativeCallCode)
	codeData, err := hex.DecodeString(code)
	if err != nil {
		return nil, fmt.Errorf("hex.DecodeString error:%s", err)
	}
	config := &smartcontract.Config{
		Tx: &types.Transaction{},
	}
	sc := &smartcontract.SmartContract{
		Config:     config,
		Gas:        10000,
		CloneCache: nil,
	}

	invokeInfo := &InvokeCodeInfo{}
	engine, err := sc.NewExecuteEngine(codeData)
	if err != nil {
		return nil, fmt.Errorf("NewExecuteEngine error:%s", err)
	}
	_, err = engine.Invoke()
	if err != nil {
		return nil, fmt.Errorf("engine.Invoke code:%s error:%s", code, err)
	}
	evaStack := engine.(*svrneo.NeoVmService).Engine.EvaluationStack
	version, err := evaStack.Peek(0).GetBigInteger()
	if err != nil {
		return nil, fmt.Errorf("get version error:%s", err)
	}
	invokeInfo.Version = version.Int64()

	contract, err := evaStack.Peek(1).GetByteArray()
	if err != nil {
		return nil, fmt.Errorf("get contract address:%s", err)
	}
	contractAddr, _ := common.AddressParseFromBytes(contract)
	invokeInfo.Contract = contractAddr.ToHexString()

	method, err := evaStack.Peek(2).GetByteArray()
	if err != nil {
		return nil, fmt.Errorf("get method error:%s", err)
	}
	invokeInfo.Method = string(method)

	var params []interface{}
	switch invokeInfo.Method {
	case ont.TRANSFER_NAME:
		params, err = ParseNativeTransferParams(evaStack)
		if err != nil {
			return nil, fmt.Errorf("ParseNativeTransferParams error:%s", err)
		}
	case ont.APPROVE_NAME:
		params, err = ParseNativeApproveParams(evaStack)
		if err != nil {
			return nil, fmt.Errorf("ParseNativeTransferParams error:%s", err)
		}
	case ont.TRANSFERFROM_NAME:
		params, err = ParseNativeTransferFromParams(evaStack)
		if err != nil {
			return nil, fmt.Errorf("ParseNativeTransferFromParams error:%s", err)
		}
	}
	invokeInfo.Params = params
	return invokeInfo, nil
}

type NativeTransferState struct {
	From  string
	To    string
	Value uint64
}

func ParseNativeTransferParams(evaStack *neovm.RandomAccessStack) ([]interface{}, error) {
	states, err := evaStack.Peek(3).GetArray()
	if err != nil {
		return nil, fmt.Errorf("get states error:%s", err)
	}
	params := make([]interface{}, 0, 1)
	statesParam := make([]*NativeTransferState, 0, len(states))
	for i := 0; i < len(states); i++ {
		state, err := states[i].GetStruct()
		if err != nil {
			return nil, fmt.Errorf("get state error:%s", err)
		}
		if len(state) != 3 {
			return nil, fmt.Errorf("tranfer states params len:%d != 3", len(state))
		}
		fromData, err := state[0].GetByteArray()
		if err != nil {
			return nil, fmt.Errorf("get from error:%s", err)
		}
		frAddr, err := common.AddressParseFromBytes(fromData)
		if err != nil {
			return nil, fmt.Errorf("AddressParseFromBytes from error:%s", err)
		}
		toData, err := state[1].GetByteArray()
		if err != nil {
			return nil, fmt.Errorf("get to error:%s", err)
		}
		toAddr, err := common.AddressParseFromBytes(toData)
		if err != nil {
			return nil, fmt.Errorf("AddressParseFromBytes to error:%s", err)
		}
		amount, err := state[2].GetBigInteger()
		if err != nil {
			return nil, fmt.Errorf("get amount error:%s", err)
		}
		statesParam = append(statesParam, &NativeTransferState{
			From:  frAddr.ToBase58(),
			To:    toAddr.ToBase58(),
			Value: amount.Uint64(),
		})
	}
	params = append(params, statesParam)
	return params, nil
}

func ParseNativeApproveParams(evaStack *neovm.RandomAccessStack) ([]interface{}, error) {
	state, err := evaStack.Peek(3).GetStruct()
	if err != nil {
		return nil, fmt.Errorf("get state error:%s", err)
	}
	if err != nil {
		return nil, fmt.Errorf("get state error:%s", err)
	}
	if len(state) != 3 {
		return nil, fmt.Errorf("approve states params len:%d != 3", len(state))
	}
	fromData, err := state[0].GetByteArray()
	if err != nil {
		return nil, fmt.Errorf("get from error:%s", err)
	}
	frAddr, err := common.AddressParseFromBytes(fromData)
	if err != nil {
		return nil, fmt.Errorf("AddressParseFromBytes from error:%s", err)
	}
	toData, err := state[1].GetByteArray()
	if err != nil {
		return nil, fmt.Errorf("get to error:%s", err)
	}
	toAddr, err := common.AddressParseFromBytes(toData)
	if err != nil {
		return nil, fmt.Errorf("AddressParseFromBytes to error:%s", err)
	}
	amount, err := state[2].GetBigInteger()
	if err != nil {
		return nil, fmt.Errorf("get amount error:%s", err)
	}
	params := make([]interface{}, 0, 1)
	params = append(params, &NativeTransferState{
		From:  frAddr.ToBase58(),
		To:    toAddr.ToBase58(),
		Value: amount.Uint64(),
	})
	return params, nil
}

type NativeTransferFromState struct {
	Sender string
	From   string
	To     string
	Value  uint64
}

func ParseNativeTransferFromParams(evaStack *neovm.RandomAccessStack) ([]interface{}, error) {
	state, err := evaStack.Peek(3).GetStruct()
	if err != nil {
		return nil, fmt.Errorf("get state error:%s", err)
	}
	if err != nil {
		return nil, fmt.Errorf("get state error:%s", err)
	}
	if len(state) != 4 {
		return nil, fmt.Errorf("tranferfrom states params len:%d != 3", len(state))
	}
	senderData, err := state[0].GetByteArray()
	if err != nil {
		return nil, fmt.Errorf("get sender error:%s", err)
	}
	senderAddr, err := common.AddressParseFromBytes(senderData)
	if err != nil {
		return nil, fmt.Errorf("AddressParseFromBytes error:%s", err)
	}
	fromData, err := state[1].GetByteArray()
	if err != nil {
		return nil, fmt.Errorf("get from error:%s", err)
	}
	frAddr, err := common.AddressParseFromBytes(fromData)
	if err != nil {
		return nil, fmt.Errorf("AddressParseFromBytes from error:%s", err)
	}
	toData, err := state[2].GetByteArray()
	if err != nil {
		return nil, fmt.Errorf("get to error:%s", err)
	}
	toAddr, err := common.AddressParseFromBytes(toData)
	if err != nil {
		return nil, fmt.Errorf("AddressParseFromBytes to error:%s", err)
	}
	amount, err := state[3].GetBigInteger()
	if err != nil {
		return nil, fmt.Errorf("get amount error:%s", err)
	}
	params := make([]interface{}, 0, 1)
	params = append(params, &NativeTransferFromState{
		Sender: senderAddr.ToBase58(),
		From:   frAddr.ToBase58(),
		To:     toAddr.ToBase58(),
		Value:  amount.Uint64(),
	})
	return params, nil
}
