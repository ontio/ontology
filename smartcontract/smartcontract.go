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
	vmtypes "github.com/Ontology/vm/types"
	"github.com/Ontology/vm/neovm/interfaces"
	ctypes "github.com/Ontology/core/types"
	"github.com/Ontology/smartcontract/service/native"
	scommon "github.com/Ontology/core/store/common"
	sneovm "github.com/Ontology/smartcontract/service/neovm"
	"github.com/Ontology/core/store"
	stypes "github.com/Ontology/smartcontract/types"
	"github.com/Ontology/vm/neovm"
	"github.com/Ontology/smartcontract/context"
	"github.com/Ontology/smartcontract/event"
)

type SmartContract struct {
	Context []*context.Context
	Config *Config
	Engine Engine
	Notifications []*event.NotifyEventInfo
}

type Config struct {
	Time uint32
	Height uint32
	Tx *ctypes.Transaction
	Table interfaces.ICodeTable
	DBCache scommon.IStateStore
	Store store.ILedgerStore
}

type Engine interface {
	StepInto()
}

//put current context to smart contract
func(sc *SmartContract) LoadContext(context *context.Context) {
	sc.Context = append(sc.Context, context)
}

//get smart contract current context
func(sc *SmartContract) CurrentContext() *context.Context {
	if len(sc.Context) < 1 {
		return nil
	}
	return sc.Context[len(sc.Context) - 1]
}

//get smart contract caller context
func(sc *SmartContract) CallingContext() *context.Context {
	if len(sc.Context) < 2 {
		return nil
	}
	return sc.Context[len(sc.Context) - 2]
}

//get smart contract entry entrance context
func(sc *SmartContract) EntryContext() *context.Context {
	if len(sc.Context) < 1 {
		return nil
	}
	return sc.Context[0]
}

func (sc *SmartContract) Execute() error {
	ctx := sc.CurrentContext()
	switch ctx.Code.VmType {
	case vmtypes.Native:
		service := native.NewNativeService(sc.Config.DBCache, sc.Config.Height, sc.Config.Tx, sc)
		if err := service.Invoke(); err != nil {
			return err
		}
		sc.Notifications = append(sc.Notifications, service.Notifications...)
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
		sc.Notifications = append(sc.Notifications, stateMachine.Notifications...)
	case vmtypes.WASMVM:
	}
	return nil
}
