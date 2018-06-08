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
package wasmvm

import (
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/vm/wasmvm/exec"
	"github.com/ontio/ontology/vm/wasmvm/util"
)

func (this *WasmVmService) blockGetCurrentHeaderHash(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()
	vm.RestoreCtx()

	headerHash := this.Store.GetCurrentHeaderHash()
	//change hash to hexstring format
	idx, err := vm.SetPointerMemory(common.ToHexString(headerHash.ToArray()))
	if err != nil {
		return false, err
	}
	vm.RestoreCtx()
	vm.PushResult(uint64(idx))
	return true, nil
}

func (this *WasmVmService) blockGetCurrentHeaderHeight(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()
	vm.RestoreCtx()
	headerHight := this.Store.GetCurrentHeaderHeight()
	vm.RestoreCtx()
	vm.PushResult(uint64(headerHight))
	return true, nil
}

func (this *WasmVmService) blockGetCurrentBlockHash(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()
	vm.RestoreCtx()

	bHash := this.Store.GetCurrentBlockHash()
	//change hash to hexstring format
	idx, err := vm.SetPointerMemory(common.ToHexString(bHash.ToArray()))
	if err != nil {
		return false, err
	}
	vm.RestoreCtx()
	vm.PushResult(uint64(idx))
	return true, nil
}

func (this *WasmVmService) blockGetBlockByHash(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 1 {
		return false, errors.NewErr("[blockGetBlockByHash]parameter count error ")
	}

	//it's a hexstring
	hashbytes, err := vm.GetPointerMemory(params[0])
	if err != nil {
		return false, err
	}

	hexbytes, err := common.HexToBytes(util.TrimBuffToString(hashbytes))
	if err != nil {
		return false, err
	}
	hash, err := common.Uint256ParseFromBytes(hexbytes)
	if err != nil {
		return false, err
	}
	block, err := this.Store.GetBlockByHash(hash)
	if err != nil {
		return false, err
	}
	//change hash to hexstring format
	idx, err := vm.SetPointerMemory(block.ToArray())
	if err != nil {
		return false, err
	}
	vm.RestoreCtx()
	vm.PushResult(uint64(idx))
	return true, nil
}

func (this *WasmVmService) blockGetBlockByHeight(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 1 {
		return false, errors.NewErr("[blockGetBlockByHeight]parameter count error ")
	}

	block, err := this.Store.GetBlockByHeight(uint32(params[0]))
	if err != nil {
		return false, err
	}
	//change hash to hexstring format
	idx, err := vm.SetPointerMemory(common.ToHexString(block.ToArray()))
	if err != nil {
		return false, err
	}
	vm.RestoreCtx()
	vm.PushResult(uint64(idx))
	return true, nil
}

func (this *WasmVmService) blockGetCurrentBlockHeight(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()
	vm.RestoreCtx()
	bHight := this.Store.GetCurrentBlockHeight()
	vm.RestoreCtx()
	vm.PushResult(uint64(bHight))
	return true, nil
}

func (this *WasmVmService) blockGetTransactionByHash(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 1 {
		return false, errors.NewErr("[RuntimeLog]parameter count error ")
	}

	//it's a hexstring
	hashbytes, err := vm.GetPointerMemory(params[0])
	if err != nil {
		return false, err
	}

	hexbytes, err := common.HexToBytes(util.TrimBuffToString(hashbytes))
	if err != nil {
		return false, err
	}
	//hextobytes
	thash, err := common.Uint256ParseFromBytes(hexbytes)

	if err != nil {
		return false, err
	}
	tx, _, err := this.Store.GetTransaction(thash)

	txbytes := tx.ToArray()

	idx, err := vm.SetPointerMemory(txbytes)
	if err != nil {
		return false, err
	}
	vm.RestoreCtx()
	vm.PushResult(uint64(idx))
	return true, nil

}

// BlockGetTransactionCount put block's transactions count to vm stack
func (this *WasmVmService) blockGetTransactionCountByBlkHash(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 1 {
		return false, errors.NewErr("[RuntimeLog]parameter count error ")
	}

	blockhash, err := vm.GetPointerMemory(params[0])
	if err != nil {
		return false, err
	}

	hexbytes, err := common.HexToBytes(util.TrimBuffToString(blockhash))
	if err != nil {
		return false, err
	}
	hash, err := common.Uint256ParseFromBytes(hexbytes)
	if err != nil {
		return false, err
	}

	block, err := this.Store.GetBlockByHash(hash)
	if err != nil {
		return false, err
	}

	length := len(block.Transactions)

	vm.RestoreCtx()
	vm.PushResult(uint64(length))
	return true, nil
}

// BlockGetTransactionCount put block's transactions count to vm stack
func (this *WasmVmService) blockGetTransactionCountByBlkHeight(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 1 {
		return false, errors.NewErr("[RuntimeLog]parameter count error ")
	}

	block, err := this.Store.GetBlockByHeight(uint32(params[0]))
	if err != nil {
		return false, err
	}

	length := len(block.Transactions)

	vm.RestoreCtx()
	vm.PushResult(uint64(length))
	return true, nil
}

// BlockGetTransactions put block's transactions to vm stack
func (this *WasmVmService) blockGetTransactionsByBlkHash(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 1 {
		return false, errors.NewErr("[RuntimeLog]parameter count error ")
	}

	blockhash, err := vm.GetPointerMemory(params[0])
	if err != nil {
		return false, err
	}

	hexbytes, err := common.HexToBytes(util.TrimBuffToString(blockhash))
	if err != nil {
		return false, err
	}
	hash, err := common.Uint256ParseFromBytes(hexbytes)
	if err != nil {
		return false, err
	}

	block, err := this.Store.GetBlockByHash(hash)
	if err != nil {
		return false, err
	}

	transactionList := make([]string, len(block.Transactions))
	for i, tx := range block.Transactions {
		hash := tx.Hash()

		transactionList[i] = common.ToHexString(hash.ToArray())
	}

	idx, err := vm.SetPointerMemory(transactionList)
	if err != nil {
		return false, err
	}
	vm.RestoreCtx()
	vm.PushResult(uint64(idx))

	return true, nil
}

// BlockGetTransactions put block's transactions to vm stack
func (this *WasmVmService) blockGetTransactionsByBlkHeight(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 1 {
		return false, errors.NewErr("[RuntimeLog]parameter count error ")
	}

	block, err := this.Store.GetBlockByHeight(uint32(params[0]))
	if err != nil {
		return false, err
	}

	transactionList := make([]string, len(block.Transactions))
	for i, tx := range block.Transactions {
		hash := tx.Hash()
		transactionList[i] = common.ToHexString(hash.ToArray())
	}

	idx, err := vm.SetPointerMemory(transactionList)
	if err != nil {
		return false, err
	}
	vm.RestoreCtx()
	vm.PushResult(uint64(idx))

	return true, nil
}
