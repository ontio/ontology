// Copyright 2017 The Ontology Authors
// This file is part of the Ontology library.
//
// The Ontology library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Ontology library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Ontology library. If not, see <http://www.gnu.org/licenses/>.

package smartcontract

import (
	"github.com/Ontology/common"
	"github.com/Ontology/core/contract"
	sig "github.com/Ontology/core/signature"
	"github.com/Ontology/smartcontract/service"
	"github.com/Ontology/smartcontract/types"
	"github.com/Ontology/vm/neovm"
	"github.com/Ontology/vm/neovm/interfaces"
	"math/big"
	"github.com/Ontology/core/store"
	"github.com/Ontology/errors"
	"github.com/Ontology/common/log"
	scommon "github.com/Ontology/smartcontract/common"
	"reflect"
)

type SmartContract struct {
	Engine         Engine
	Code           []byte
	Input          []byte
	ParameterTypes []contract.ContractParameterType
	Caller         common.Uint160
	CodeHash       common.Uint160
	VMType         types.VmType
	ReturnType     contract.ContractParameterType
}

type Context struct {
	VmType         types.VmType
	Caller         common.Uint160
	StateMachine   *service.StateMachine
	DBCache        store.IStateStore
	Code           []byte
	Input          []byte
	CodeHash       common.Uint160
	Time           *big.Int
	BlockNumber    *big.Int
	CacheCodeTable interfaces.ICodeTable
	SignableData   sig.SignableData
	Gas            common.Fixed64
	ReturnType     contract.ContractParameterType
	ParameterTypes []contract.ContractParameterType
}

type Engine interface {
	Create(caller common.Uint160, code []byte) ([]byte, error)
	Call(caller common.Uint160, code, input []byte) ([]byte, error)
}

func NewSmartContract(context *Context) (*SmartContract, error) {
	var e Engine
	switch context.VmType {
	case types.NEOVM:
		e = neovm.NewExecutionEngine(
			context.SignableData,
			new(neovm.ECDsaCrypto),
			context.CacheCodeTable,
			context.StateMachine,
		)
	default:
		return nil, errors.NewErr("[NewSmartContract] Invalid vm type!")
	}
	return &SmartContract{
		Engine:         e,
		Code:           context.Code,
		CodeHash:       context.CodeHash,
		Input:          context.Input,
		Caller:         context.Caller,
		VMType:         context.VmType,
		ReturnType:     context.ReturnType,
		ParameterTypes: context.ParameterTypes,
	}, nil
}

func (sc *SmartContract) DeployContract() ([]byte, error) {
	return sc.Engine.Create(sc.Caller, sc.Code)
}

func (sc *SmartContract) InvokeContract() (interface{}, error) {
	_, err := sc.Engine.Call(sc.Caller, sc.Code, sc.Input)
	if err != nil {
		return nil, err
	}
	return sc.InvokeResult()
}


func (sc *SmartContract) InvokeResult() (interface{}, error) {
	switch sc.VMType {
	case types.NEOVM:
		engine := sc.Engine.(*neovm.ExecutionEngine)
		if engine.GetEvaluationStackCount() > 0 && neovm.Peek(engine).GetStackItem() != nil {
			switch sc.ReturnType {
			case contract.Boolean:
				return neovm.PopBoolean(engine), nil
			case contract.Integer:
				log.Error(reflect.TypeOf(neovm.Peek(engine).GetStackItem().GetByteArray()))
				return neovm.PopBigInt(engine).Int64(), nil
			case contract.ByteArray:
				return common.ToHexString(neovm.PopByteArray(engine)), nil
			case contract.String:
				return string(neovm.PopByteArray(engine)), nil
			case contract.Hash160, contract.Hash256:
				return common.ToHexString(neovm.PopByteArray(engine)), nil
			case contract.PublicKey:
				return common.ToHexString(neovm.PopByteArray(engine)), nil
			case contract.InteropInterface:
				if neovm.PeekInteropInterface(engine) != nil {
					return common.ToHexString(neovm.PopInteropInterface(engine).ToArray()), nil
				}
				return nil, nil
			case contract.Array:
				var states []interface{}
				arr := neovm.PeekArray(engine)
				for _, v := range arr {
					states = append(states, scommon.ConvertReturnTypes(v)...)
				}
				return states, nil
			default:
				return common.ToHexString(neovm.PopByteArray(engine)), nil
			}
		}
	}
	return nil, nil
}