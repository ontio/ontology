// Copyright 2017 The Onchain Authors
// This file is part of the Onchain library.
//
// The Onchain library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Onchain library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Onchain library. If not, see <http://www.gnu.org/licenses/>.

package smartcontract

import (
	"github.com/Ontology/common"
	"github.com/Ontology/core/contract"
	"github.com/Ontology/smartcontract/service"
	vmtypes "github.com/Ontology/vm/types"
	"reflect"
	"github.com/Ontology/vm/neovm"
	"github.com/Ontology/errors"
	"github.com/Ontology/common/log"
	"github.com/Ontology/core/store"
	scommon "github.com/Ontology/core/store/common"
	"github.com/Ontology/core/types"
)

type Context struct {
	LedgerStore store.ILedgerStore
	Code vmtypes.VmCode
	DBCache scommon.IStateStore
	TX *types.Transaction
	Time uint32
}

type SmartContract struct {
	Input          []byte
	VMType         vmtypes.VmType
}

type Engine interface {
	Create(caller common.Address, code []byte) ([]byte, error)
	Call(caller common.Address, code, input []byte) ([]byte, error)
}

func NewSmartContract(context *Context) (*SmartContract, error) {
	var e Engine
	switch context.Code.VmType {
	case vmtypes.NEOVM:
		stateMachine := service.NewStateMachine(context.LedgerStore, context.DBCache, vmtypes.Application, context.Time)
		e = neovm.NewExecutionEngine(
			context.TX,
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