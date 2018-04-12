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
	"fmt"
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
	"github.com/ontio/ontology/smartcontract/service/wasm"
	stypes "github.com/ontio/ontology/smartcontract/types"
	"github.com/ontio/ontology/vm/wasmvm/exec"
	"github.com/ontio/ontology/vm/wasmvm/util"
	"github.com/ontio/ontology/smartcontract/service/neovm"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/smartcontract/states"
	vm "github.com/ontio/ontology/vm/neovm"
)

var (
	CONTRACT_NOT_EXIST = errors.NewErr("[AppCall] Get contract context nil")
	DEPLOYCODE_TYPE_ERROR = errors.NewErr("[AppCall] DeployCode type error!")
	INVOKE_CODE_EXIST = errors.NewErr("[AppCall] Invoke codes exist!")
)

type SmartContract struct {
	Contexts      []*context.Context       // all execute smart contract context
	Config        *Config
	Engine        Engine
	Notifications []*event.NotifyEventInfo // all execute smart contract event notify info
}

type Config struct {
	Time    uint32              // current block timestamp
	Height  uint32              // current block height
	Tx      *ctypes.Transaction // current transaction
	DBCache scommon.StateStore  // db states cache
	Store   store.LedgerStore   // ledger store
}

type Engine interface {
	Invoke()
}

//put current context to smart contract
func (this *SmartContract) PushContext(context *context.Context) {
	this.Contexts = append(this.Contexts, context)
}

//get smart contract current context
func (this *SmartContract) CurrentContext() *context.Context {
	if len(this.Contexts) < 1 {
		return nil
	}
	return this.Contexts[len(this.Contexts) - 1]
}

//get smart contract caller context
func (this *SmartContract) CallingContext() *context.Context {
	if len(this.Contexts) < 2 {
		return nil
	}
	return this.Contexts[len(this.Contexts) - 2]
}

//get smart contract entry entrance context
func (this *SmartContract) EntryContext() *context.Context {
	if len(this.Contexts) < 1 {
		return nil
	}
	return this.Contexts[0]
}

//pop smart contract current context
func (this *SmartContract) PopContext() {
	if len(this.Contexts) > 0 {
		this.Contexts = this.Contexts[:len(this.Contexts) - 1]
	}
}

// push smart contract event info
func (this *SmartContract) PushNotifications(notifications []*event.NotifyEventInfo) {
	this.Notifications = append(this.Notifications, notifications...)
}

func (this *SmartContract) Execute() error {
	ctx := this.CurrentContext()
	switch ctx.Code.VmType {
	case stypes.Native:
		service := native.NewNativeService(this.Config.DBCache, this.Config.Height, this.Config.Tx, this)
		if err := service.Invoke(); err != nil {
			return err
		}
	case stypes.NEOVM:
		service := neovm.NewNeoVmService(this.Config.Store, this.Config.DBCache, this.Config.Tx, this.Config.Time, this)
		if err := service.Invoke(); err != nil {
			fmt.Println("execute neovm error:", err)
			return err
		}
	case stypes.WASMVM:
		stateMachine := wasm.NewWasmStateMachine(this.Config.Store, this.Config.DBCache, this.Config.Time)

		engine := exec.NewExecutionEngine(
			this.Config.Tx,
			new(util.ECDsaCrypto),
			stateMachine,
		)

		contract := &states.Contract{}
		contract.Deserialize(bytes.NewBuffer(ctx.Code.Code))
		addr := contract.Address

		dpcode, err := stateMachine.GetContractCodeFromAddress(addr)
		if err != nil {
			return errors.NewErr("get contract  error")
		}

		var caller common.Address
		if this.CallingContext() == nil {
			caller = common.Address{}
		} else {
			caller = this.CallingContext().ContractAddress
		}
		res, err := engine.Call(caller, dpcode, contract.Method, contract.Args, contract.Version)

		if err != nil {
			return err
		}

		//get the return message
		_, err = engine.GetVM().GetPointerMemory(uint64(binary.LittleEndian.Uint32(res)))
		if err != nil {
			return err
		}

		stateMachine.CloneCache.Commit()
		this.Notifications = append(this.Notifications, stateMachine.Notifications...)
	}
	return nil
}

// When you want to call a contract use this function, if contract exist in block chain, you should set isLoad true,
// Otherwise, you can set execute code, and set isLoad false.
// param address: smart contract address
// param method: invoke smart contract method name
// param codes: invoke smart contract whether need to load code
// param args: invoke smart contract args
func (this *SmartContract) AppCall(address common.Address, method string, codes, args []byte) error {
	var code []byte

	vmType := stypes.VmType(address[0])

	switch vmType {
	case stypes.Native:
		bf := new(bytes.Buffer)
		c := states.Contract{
			Address: address,
			Method: method,
			Args: args,
		}
		if err := c.Serialize(bf); err != nil {
			return err
		}
		code = bf.Bytes()
	case stypes.NEOVM:
		code, err := this.loadCode(address, codes);
		if err != nil {
			return nil
		}
		var temp []byte
		build := vm.NewParamsBuilder(new(bytes.Buffer))
		if method != "" {
			build.EmitPushByteArray([]byte(method))
		}
		temp = append(args, build.ToArray()...)
		code = append(temp, code...)
	case stypes.WASMVM:
	}

	this.PushContext(&context.Context{
		Code: stypes.VmCode{
			Code: code,
			VmType: vmType,
		},
		ContractAddress: address,
	})

	if err := this.Execute(); err != nil {
		return err
	}

	this.PopContext()
	return nil
}

// check authorization correct
// if address is wallet address, check whether in the signature addressed list
// else check whether address is calling contract address
// param address: wallet address or contract address
func (this *SmartContract) CheckWitness(address common.Address) bool {
	if stypes.IsVmCodeAddress(address) {
		if this.CallingContext() != nil && this.CallingContext().ContractAddress == address {
			return true
		}
	} else {
		addresses := this.Config.Tx.GetSignatureAddresses()
		for _, v := range addresses {
			if v == address {
				return true
			}
		}
	}

	return false
}

// load smart contract execute code
// param address, invoke onchain smart contract address
// param codes, invoke offchain smart contract code
// if you invoke offchain smart contract, you can set address is codes address
// but this address cann't find in blockchain
func (this *SmartContract) loadCode(address common.Address, codes []byte) ([]byte, error) {
	isLoad := false
	if len(codes) == 0 {
		isLoad = true
	}
	item, err := this.getContract(address[:]); if err != nil {
		return nil, err
	}
	if isLoad {
		if item == nil {
			return nil, CONTRACT_NOT_EXIST
		}
		contract, ok := item.Value.(*payload.DeployCode); if !ok {
			return nil, DEPLOYCODE_TYPE_ERROR
		}
		return contract.Code.Code, nil
	} else {
		if item != nil {
			return nil, INVOKE_CODE_EXIST
		}
		return codes, nil
	}
}

func (this *SmartContract) getContract(address []byte) (*scommon.StateItem, error) {
	item, err := this.Config.DBCache.TryGet(scommon.ST_CONTRACT, address[:]);
	if err != nil {
		return nil, errors.NewErr("[getContract] Get contract context error!")
	}
	return item, nil
}