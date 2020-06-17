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

package ledgerstore

import (
	"bytes"
	"fmt"
	"math"
	"strconv"

	"github.com/ontio/ontology/common"
	sysconfig "github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/store"
	scommon "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/store/overlaydb"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/service/native/global_params"
	ninit "github.com/ontio/ontology/smartcontract/service/native/init"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/ontio/ontology/smartcontract/service/neovm"
	"github.com/ontio/ontology/smartcontract/service/wasmvm"
	"github.com/ontio/ontology/smartcontract/storage"
)

func tuneGasFeeByHeight(height uint32, gas uint64, gasRound uint64, curBalance uint64) uint64 {
	gasTuneheight := sysconfig.GetGasRoundTuneHeight(sysconfig.DefConfig.P2PNode.NetworkId)
	if height > gasTuneheight {
		t := (gas + gasRound - 1) / gasRound
		if gas > math.MaxUint64-gasRound {
			return curBalance
		}

		newGas := gasRound * t

		if newGas > curBalance {
			return curBalance
		}

		return newGas
	}

	return gas
}

//HandleDeployTransaction deal with smart contract deploy transaction
func (self *StateStore) HandleDeployTransaction(store store.LedgerStore, overlay *overlaydb.OverlayDB, gasTable map[string]uint64, cache *storage.CacheDB,
	tx *types.Transaction, block *types.Block, notify *event.ExecuteNotify) error {
	deploy := tx.Payload.(*payload.DeployCode)
	var (
		notifies    []*event.NotifyEventInfo
		gasConsumed uint64
		err         error
	)

	if deploy.VmType() == payload.WASMVM_TYPE {
		_, err = wasmvm.ReadWasmModule(deploy.GetRawCode(), sysconfig.DefConfig.Common.WasmVerifyMethod)
		if err != nil {
			return err
		}
	}

	if tx.GasPrice != 0 {
		// init smart contract configuration info
		config := &smartcontract.Config{
			Time:      block.Header.Timestamp,
			Height:    block.Header.Height,
			Tx:        tx,
			BlockHash: block.Hash(),
		}
		createGasPrice, ok := gasTable[neovm.CONTRACT_CREATE_NAME]
		if !ok {
			overlay.SetError(errors.NewErr("[HandleDeployTransaction] get CONTRACT_CREATE_NAME gas failed"))
			return nil
		}

		uintCodePrice, ok := gasTable[neovm.UINT_DEPLOY_CODE_LEN_NAME]
		if !ok {
			overlay.SetError(errors.NewErr("[HandleDeployTransaction] get UINT_DEPLOY_CODE_LEN_NAME gas failed"))
			return nil
		}

		gasLimit := createGasPrice + calcGasByCodeLen(len(deploy.GetRawCode()), uintCodePrice)
		balance, err := isBalanceSufficient(tx.Payer, cache, config, store, gasLimit*tx.GasPrice)
		if err != nil {
			if err := costInvalidGas(tx.Payer, balance, config, overlay, store, notify); err != nil {
				return err
			}
			return err
		}
		if tx.GasLimit < gasLimit {
			if err := costInvalidGas(tx.Payer, tx.GasLimit*tx.GasPrice, config, overlay, store, notify); err != nil {
				return err
			}
			return fmt.Errorf("gasLimit insufficient, need:%d actual:%d", gasLimit, tx.GasLimit)

		}
		gasConsumed = gasLimit * tx.GasPrice
		notifies, err = chargeCostGas(tx.Payer, gasConsumed, config, cache, store)
		if err != nil {
			return err
		}
		cache.Commit()
	}

	address := deploy.Address()
	// store contract message
	dep, err := cache.GetContract(address)
	if err != nil {
		return err
	}
	if dep == nil {
		log.Infof("deploy contract address:%s", address.ToHexString())
		cache.PutContract(deploy)
	}
	cache.Commit()

	notify.Notify = append(notify.Notify, notifies...)
	notify.GasConsumed = gasConsumed
	notify.State = event.CONTRACT_STATE_SUCCESS
	return nil
}

//HandleInvokeTransaction deal with smart contract invoke transaction
func (self *StateStore) HandleInvokeTransaction(store store.LedgerStore, overlay *overlaydb.OverlayDB, gasTable map[string]uint64, cache *storage.CacheDB,
	tx *types.Transaction, block *types.Block, notify *event.ExecuteNotify) ([]common.Uint256, error) {
	invoke := tx.Payload.(*payload.InvokeCode)
	code := invoke.Code
	sysTransFlag := bytes.Compare(code, ninit.COMMIT_DPOS_BYTES) == 0 || block.Header.Height == 0

	isCharge := !sysTransFlag && tx.GasPrice != 0

	// init smart contract configuration info
	config := &smartcontract.Config{
		Time:      block.Header.Timestamp,
		Height:    block.Header.Height,
		Tx:        tx,
		BlockHash: block.Hash(),
	}

	var (
		costGasLimit      uint64
		costGas           uint64
		oldBalance        uint64
		newBalance        uint64
		codeLenGasLimit   uint64
		availableGasLimit uint64
		minGas            uint64
		err               error
	)

	availableGasLimit = tx.GasLimit
	if isCharge {
		uintCodeGasPrice, ok := gasTable[neovm.UINT_INVOKE_CODE_LEN_NAME]
		if !ok {
			overlay.SetError(errors.NewErr("[HandleInvokeTransaction] get UINT_INVOKE_CODE_LEN_NAME gas failed"))
			return nil, nil
		}

		oldBalance, err = getBalanceFromNative(config, cache, store, tx.Payer)
		if err != nil {
			return nil, err
		}

		minGas = neovm.MIN_TRANSACTION_GAS * tx.GasPrice

		if oldBalance < minGas {
			if err := costInvalidGas(tx.Payer, oldBalance, config, overlay, store, notify); err != nil {
				return nil, err
			}
			return nil, fmt.Errorf("balance gas: %d less than min gas: %d", oldBalance, minGas)
		}

		codeLenGasLimit = calcGasByCodeLen(len(invoke.Code), uintCodeGasPrice)

		if oldBalance < codeLenGasLimit*tx.GasPrice {
			if err := costInvalidGas(tx.Payer, oldBalance, config, overlay, store, notify); err != nil {
				return nil, err
			}
			return nil, fmt.Errorf("balance gas insufficient: balance:%d < code length need gas:%d", oldBalance, codeLenGasLimit*tx.GasPrice)
		}

		if tx.GasLimit < codeLenGasLimit {
			if err := costInvalidGas(tx.Payer, tx.GasLimit*tx.GasPrice, config, overlay, store, notify); err != nil {
				return nil, err
			}
			return nil, fmt.Errorf("invoke transaction gasLimit insufficient: need%d actual:%d", tx.GasLimit, codeLenGasLimit)
		}

		maxAvaGasLimit := oldBalance / tx.GasPrice

		if availableGasLimit > maxAvaGasLimit {
			availableGasLimit = maxAvaGasLimit
		}
	}

	//init smart contract info
	sc := smartcontract.SmartContract{
		Config:       config,
		CacheDB:      cache,
		Store:        store,
		GasTable:     gasTable,
		Gas:          availableGasLimit - codeLenGasLimit,
		WasmExecStep: sysconfig.DEFAULT_WASM_MAX_STEPCOUNT,
		PreExec:      false,
	}

	//start the smart contract executive function
	engine, _ := sc.NewExecuteEngine(invoke.Code, tx.TxType)

	_, err = engine.Invoke()
	if sc.IsInternalErr() {
		overlay.SetError(fmt.Errorf("[HandleInvokeTransaction] %s", err))
		return nil, nil
	}

	costGasLimit = availableGasLimit - sc.Gas
	if costGasLimit < neovm.MIN_TRANSACTION_GAS {
		costGasLimit = neovm.MIN_TRANSACTION_GAS
	}

	costGas = costGasLimit * tx.GasPrice
	if err != nil {
		if isCharge {
			costGas = tuneGasFeeByHeight(config.Height, costGas, tx.GasPrice*neovm.MIN_TRANSACTION_GAS, oldBalance)
			if err := costInvalidGas(tx.Payer, costGas, config, overlay, store, notify); err != nil {
				return nil, err
			}
		}
		return nil, err
	}

	var notifies []*event.NotifyEventInfo
	if isCharge {
		newBalance, err = getBalanceFromNative(config, cache, store, tx.Payer)
		if err != nil {
			return nil, err
		}

		if newBalance < costGas {
			costGas = tuneGasFeeByHeight(config.Height, costGas, tx.GasPrice*neovm.MIN_TRANSACTION_GAS, oldBalance)
			if err := costInvalidGas(tx.Payer, costGas, config, overlay, store, notify); err != nil {
				return nil, err
			}
			return nil, fmt.Errorf("gas insufficient, balance:%d < costGas:%d", newBalance, costGas)
		}

		costGas = tuneGasFeeByHeight(config.Height, costGas, tx.GasPrice*neovm.MIN_TRANSACTION_GAS, newBalance)
		notifies, err = chargeCostGas(tx.Payer, costGas, config, sc.CacheDB, store)
		if err != nil {
			return nil, err
		}
	}

	notify.Notify = append(notify.Notify, sc.Notifications...)
	notify.Notify = append(notify.Notify, notifies...)
	notify.GasConsumed = costGas
	notify.State = event.CONTRACT_STATE_SUCCESS
	sc.CacheDB.Commit()
	return sc.CrossHashes, nil
}

func SaveNotify(eventStore scommon.EventStore, txHash common.Uint256, notify *event.ExecuteNotify) error {
	if !sysconfig.DefConfig.Common.EnableEventLog {
		return nil
	}
	if err := eventStore.SaveEventNotifyByTx(txHash, notify); err != nil {
		return fmt.Errorf("SaveEventNotifyByTx error %s", err)
	}
	event.PushSmartCodeEvent(txHash, 0, event.EVENT_NOTIFY, notify)
	return nil
}

func genNativeTransferCode(from, to common.Address, value uint64) []byte {
	transfer := &ont.Transfers{States: []ont.State{{From: from, To: to, Value: value}}}
	return common.SerializeToBytes(transfer)
}

// check whether payer ong balance sufficient
func isBalanceSufficient(payer common.Address, cache *storage.CacheDB, config *smartcontract.Config, store store.LedgerStore, gas uint64) (uint64, error) {
	balance, err := getBalanceFromNative(config, cache, store, payer)
	if err != nil {
		return 0, err
	}
	if balance < gas {
		return 0, fmt.Errorf("payer gas insufficient, need %d , only have %d", gas, balance)
	}
	return balance, nil
}

func chargeCostGas(payer common.Address, gas uint64, config *smartcontract.Config,
	cache *storage.CacheDB, store store.LedgerStore) ([]*event.NotifyEventInfo, error) {

	params := genNativeTransferCode(payer, utils.GovernanceContractAddress, gas)

	sc := smartcontract.SmartContract{
		Config:  config,
		CacheDB: cache,
		Store:   store,
		Gas:     math.MaxUint64,
	}

	service, _ := sc.NewNativeService()
	_, err := service.NativeCall(utils.OngContractAddress, "transfer", params)
	if err != nil {
		return nil, err
	}
	return sc.Notifications, nil
}

func refreshGlobalParam(config *smartcontract.Config, cache *storage.CacheDB, store store.LedgerStore) error {
	sink := common.NewZeroCopySink(nil)
	utils.EncodeVarUint(sink, uint64(len(neovm.GAS_TABLE_KEYS)))
	for _, value := range neovm.GAS_TABLE_KEYS {
		sink.WriteString(value)
	}

	sc := smartcontract.SmartContract{
		Config:  config,
		CacheDB: cache,
		Store:   store,
		Gas:     math.MaxUint64,
	}

	service, _ := sc.NewNativeService()
	result, err := service.NativeCall(utils.ParamContractAddress, "getGlobalParam", sink.Bytes())
	if err != nil {
		return err
	}
	params := new(global_params.Params)
	if err := params.Deserialization(common.NewZeroCopySource(result)); err != nil {
		return fmt.Errorf("deserialize global params error:%s", err)
	}
	neovm.GAS_TABLE.Range(func(key, value interface{}) bool {
		n, ps := params.GetParam(key.(string))
		if n != -1 && ps.Value != "" {
			pu, err := strconv.ParseUint(ps.Value, 10, 64)
			if err != nil {
				log.Errorf("[refreshGlobalParam] failed to parse uint %v\n", ps.Value)
			} else {
				neovm.GAS_TABLE.Store(key, pu)
			}
		}
		return true
	})
	return nil
}

func getBalanceFromNative(config *smartcontract.Config, cache *storage.CacheDB, store store.LedgerStore, address common.Address) (uint64, error) {
	bf := common.NewZeroCopySink(nil)
	utils.EncodeAddress(bf, address)
	sc := smartcontract.SmartContract{
		Config:  config,
		CacheDB: cache,
		Store:   store,
		Gas:     math.MaxUint64,
	}

	service, _ := sc.NewNativeService()
	result, err := service.NativeCall(utils.OngContractAddress, ont.BALANCEOF_NAME, bf.Bytes())
	if err != nil {
		return 0, err
	}
	return common.BigIntFromNeoBytes(result).Uint64(), nil
}

func costInvalidGas(address common.Address, gas uint64, config *smartcontract.Config, overlay *overlaydb.OverlayDB,
	store store.LedgerStore, notify *event.ExecuteNotify) error {
	cache := storage.NewCacheDB(overlay)
	notifies, err := chargeCostGas(address, gas, config, cache, store)
	if err != nil {
		return err
	}
	cache.Commit()
	notify.GasConsumed = gas
	notify.Notify = append(notify.Notify, notifies...)
	return nil
}

func calcGasByCodeLen(codeLen int, codeGas uint64) uint64 {
	return uint64(codeLen/neovm.PER_UNIT_CODE_LEN) * codeGas
}
