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
	"bytes"
	"encoding/binary"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/store"
	scommon "github.com/ontio/ontology/core/store/common"
	ctypes "github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/context"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/service/native"
	sneovm "github.com/ontio/ontology/smartcontract/service/neovm"
	"github.com/ontio/ontology/smartcontract/service/wasm"
	stypes "github.com/ontio/ontology/smartcontract/types"
	"github.com/ontio/ontology/vm/neovm"
	"github.com/ontio/ontology/vm/neovm/interfaces"
	vmtypes "github.com/ontio/ontology/vm/types"
	"github.com/ontio/ontology/vm/wasmvm/exec"
	"github.com/ontio/ontology/vm/wasmvm/util"
)

type SmartContract struct {
	Context       []*context.Context
	Config        *Config
	Engine        Engine
	Notifications []*event.NotifyEventInfo
}

type Config struct {
	Time    uint32
	Height  uint32
	Tx      *ctypes.Transaction
	Table   interfaces.CodeTable
	DBCache scommon.StateStore
	Store   store.LedgerStore
}

type Engine interface {
	StepInto()
}

//put current context to smart contract
func (sc *SmartContract) PushContext(context *context.Context) {
	sc.Context = append(sc.Context, context)
}

//get smart contract current context
func (sc *SmartContract) CurrentContext() *context.Context {
	if len(sc.Context) < 1 {
		return nil
	}
	return sc.Context[len(sc.Context)-1]
}

//get smart contract caller context
func (sc *SmartContract) CallingContext() *context.Context {
	if len(sc.Context) < 2 {
		return nil
	}
	return sc.Context[len(sc.Context)-2]
}

//get smart contract entry entrance context
func (sc *SmartContract) EntryContext() *context.Context {
	if len(sc.Context) < 1 {
		return nil
	}
	return sc.Context[0]
}

//pop smart contract current context
func (sc *SmartContract) PopContext() {
	sc.Context = sc.Context[:len(sc.Context)-1]
}

func (sc *SmartContract) PushNotifications(notifications []*event.NotifyEventInfo) {
	sc.Notifications = append(sc.Notifications, notifications...)
}

func (sc *SmartContract) Execute() error {
	ctx := sc.CurrentContext()
	switch ctx.Code.VmType {
	case vmtypes.Native:
		service := native.NewNativeService(sc.Config.DBCache, sc.Config.Height, sc.Config.Tx, sc)
		if err := service.Invoke(); err != nil {
			return err
		}
	case vmtypes.NEOVM:
		stateMachine := sneovm.NewStateMachine(sc.Config.Store, sc.Config.DBCache, stypes.Application, sc.Config.Time)
		engine := neovm.NewExecutionEngine(
			sc.Config.Tx,
			new(neovm.ECDsaCrypto),
			sc.Config.Table,
			stateMachine,
		)
		engine.LoadCode(ctx.Code.Code, false)
		if err := engine.Execute(); err != nil {
			return err
		}
		stateMachine.CloneCache.Commit()
		sc.Notifications = append(sc.Notifications, stateMachine.Notifications...)
	case vmtypes.WASMVM:
		//todo refactor following code to match Neovm
		stateMachine := wasm.NewWasmStateMachine(sc.Config.Store, sc.Config.DBCache, stypes.Application, sc.Config.Time)
		engine := exec.NewExecutionEngine(
			sc.Config.Tx,
			new(util.ECDsaCrypto),
			sc.Config.Table,
			stateMachine,
			"product",
		)

		tmpcodes := bytes.Split(ctx.Code.Code, []byte(exec.PARAM_SPLITER))
		if len(tmpcodes) != 3 {
			return errors.NewErr("Wasm paramter count error")
		}
		contractCode := tmpcodes[0]

		addr, err := common.AddressParseFromBytes(contractCode)
		if err != nil {
			return errors.NewErr("get contract address error")
		}

		dpcode, err := stateMachine.GetContractCodeFromAddress(addr)
		if err != nil {
			return errors.NewErr("get contract  error")
		}

		input := ctx.Code.Code[len(contractCode)+1:]
		res, err := engine.Call(ctx.ContractAddress, dpcode, input)
		if err != nil {
			return err
		}

		//todo how to deal with the result???
		_, err = engine.GetVM().GetPointerMemory(uint64(binary.LittleEndian.Uint32(res)))
		if err != nil {
			return err
		}

		stateMachine.CloneCache.Commit()
		sc.Notifications = append(sc.Notifications, stateMachine.Notifications...)
	}
	return nil
}

func (sc *SmartContract) CheckWitness(address common.Address) bool {
	if vmtypes.IsVmCodeAddress(address) {
		for _, v := range sc.Context {
			if v.ContractAddress == address {
				return true
			}
		}
	} else {
		addresses := sc.Config.Tx.GetSignatureAddresses()
		for _, v := range addresses {
			if v == address {
				return true
			}
		}
	}

	return false
}
