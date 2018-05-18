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
package smartcontract

import (
	"bytes"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/store"
	scommon "github.com/ontio/ontology/core/store/common"
	ctypes "github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/context"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/service/native"
	_ "github.com/ontio/ontology/smartcontract/service/native/init"
	"github.com/ontio/ontology/smartcontract/service/neovm"
	"github.com/ontio/ontology/smartcontract/service/wasmvm"
	"github.com/ontio/ontology/smartcontract/states"
	"github.com/ontio/ontology/smartcontract/storage"
	stypes "github.com/ontio/ontology/smartcontract/types"
	vm "github.com/ontio/ontology/vm/neovm"
)

var (
	CONTRACT_NOT_EXIST    = errors.NewErr("[AppCall] Get contract context nil")
	DEPLOYCODE_TYPE_ERROR = errors.NewErr("[AppCall] DeployCode type error!")
	INVOKE_CODE_EXIST     = errors.NewErr("[AppCall] Invoke codes exist!")
	ENGINE_NOT_SUPPORT    = errors.NewErr("[Execute] Engine doesn't support!")
)

// SmartContract describe smart contract execute engine
type SmartContract struct {
	Contexts      []*context.Context  // all execute smart contract context
	CloneCache    *storage.CloneCache // state cache
	Store         store.LedgerStore   // ledger store
	Config        *Config
	Engine        Engine
	Code          stypes.VmCode
	Notifications []*event.NotifyEventInfo // all execute smart contract event notify info
	Gas           uint64
}

// Config describe smart contract need parameters configuration
type Config struct {
	Time   uint32              // current block timestamp
	Height uint32              // current block height
	Tx     *ctypes.Transaction // current transaction
}

type Engine interface {
	Invoke() (interface{}, error)
}

// PushContext push current context to smart contract
func (this *SmartContract) PushContext(context *context.Context) {
	this.Contexts = append(this.Contexts, context)
}

// CurrentContext return smart contract current context
func (this *SmartContract) CurrentContext() *context.Context {
	if len(this.Contexts) < 1 {
		return nil
	}
	return this.Contexts[len(this.Contexts)-1]
}

// CallingContext return smart contract caller context
func (this *SmartContract) CallingContext() *context.Context {
	if len(this.Contexts) < 2 {
		return nil
	}
	return this.Contexts[len(this.Contexts)-2]
}

// EntryContext return smart contract entry entrance context
func (this *SmartContract) EntryContext() *context.Context {
	if len(this.Contexts) < 1 {
		return nil
	}
	return this.Contexts[0]
}

// PopContext pop smart contract current context
func (this *SmartContract) PopContext() {
	if len(this.Contexts) > 1 {
		this.Contexts = this.Contexts[:len(this.Contexts)-1]
	}
}

// PushNotifications push smart contract event info
func (this *SmartContract) PushNotifications(notifications []*event.NotifyEventInfo) {
	this.Notifications = append(this.Notifications, notifications...)
}

func (this *SmartContract) CheckUseGas(gas uint64) bool {
	if this.Gas < gas {
		return false
	}
	this.Gas -= gas
	return true
}

// Execute is smart contract execute manager
// According different vm type to launch different service
func (this *SmartContract) Execute() (interface{}, error) {
	var engine Engine
	switch this.Code.VmType {
	case stypes.Native:
		engine = &native.NativeService{
			CloneCache: this.CloneCache,
			Code:       this.Code.Code,
			Tx:         this.Config.Tx,
			Height:     this.Config.Height,
			Time:       this.Config.Time,
			ContextRef: this,
			ServiceMap: make(map[string]native.Handler),
		}
	case stypes.NEOVM:
		engine = &neovm.NeoVmService{
			Store:      this.Store,
			CloneCache: this.CloneCache,
			ContextRef: this,
			Code:       this.Code.Code,
			Tx:         this.Config.Tx,
			Time:       this.Config.Time,
		}
	case stypes.WASMVM:
		engine = &wasmvm.WasmVmService{
			Store:      this.Store,
			CloneCache: this.CloneCache,
			ContextRef: this,
			Code:       this.Code.Code,
			Tx:         this.Config.Tx,
			Time:       this.Config.Time,
		}
	default:
		return nil, ENGINE_NOT_SUPPORT
	}
	return engine.Invoke()
}

// AppCall a smart contract, if contract exist on blockchain, you should set the address
// Param address: invoke smart contract on blockchain according contract address
// Param method: invoke smart contract method name
// Param codes: invoke smart contract off blockchain
// Param args: invoke smart contract args
func (this *SmartContract) AppCall(address common.Address, method string, codes, args []byte) (interface{}, error) {
	var code []byte
	vmType := stypes.VmType(address[0])
	switch vmType {
	case stypes.Native:
		bf := new(bytes.Buffer)
		c := states.Contract{
			Address: address,
			Method:  method,
			Args:    args,
		}
		if err := c.Serialize(bf); err != nil {
			return nil, err
		}
		code = bf.Bytes()
	case stypes.NEOVM:
		c, err := this.loadCode(address, codes)
		if err != nil {
			return nil, err
		}
		var temp []byte
		build := vm.NewParamsBuilder(new(bytes.Buffer))
		if method != "" {
			build.EmitPushByteArray([]byte(method))
		}
		temp = append(args, build.ToArray()...)
		code = append(temp, c...)
		vmCode := stypes.VmCode{Code: c, VmType: stypes.NEOVM}
		this.PushContext(&context.Context{ContractAddress: vmCode.AddressFromVmCode()})
	case stypes.WASMVM:
		c, err := this.loadCode(address, codes)
		if err != nil {
			return nil, err
		}
		bf := new(bytes.Buffer)
		contract := states.Contract{
			Version: 1, //fix to > 0
			Address: address,
			Method:  method,
			Args:    args,
			Code:    c,
		}
		if err := contract.Serialize(bf); err != nil {
			return nil, err
		}
		code = bf.Bytes()
	}

	this.Code = stypes.VmCode{Code: code, VmType: vmType}
	res, err := this.Execute()
	if err != nil {
		return nil, err
	}

	return res, nil
}

// CheckWitness check whether authorization correct
// If address is wallet address, check whether in the signature addressed list
// Else check whether address is calling contract address
// Param address: wallet address or contract address
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

// loadCode load smart contract execute code
// Param address, invoke on blockchain smart contract address
// Param codes, invoke off blockchain smart contract code
// If you invoke off blockchain smart contract, you can set address is codes address
// But this address doesn't deployed on blockchain
func (this *SmartContract) loadCode(address common.Address, codes []byte) ([]byte, error) {
	isLoad := false
	if len(codes) == 0 {
		isLoad = true
	}
	item, err := this.getContract(address[:])
	if err != nil {
		return nil, err
	}
	if isLoad {
		if item == nil {
			return nil, CONTRACT_NOT_EXIST
		}
		contract, ok := item.Value.(*payload.DeployCode)
		if !ok {
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
	item, err := this.CloneCache.Store.TryGet(scommon.ST_CONTRACT, address[:])
	if err != nil {
		return nil, errors.NewErr("[getContract] Get contract context error!")
	}
	return item, nil
}
