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
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/store"
	scommon "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/store/statestore"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/service/native/global_params"
	ninit "github.com/ontio/ontology/smartcontract/service/native/init"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/ontio/ontology/smartcontract/service/neovm"
	"github.com/ontio/ontology/smartcontract/storage"
	ntypes "github.com/ontio/ontology/vm/neovm/types"
)

func handleDeployTransaction(store store.LedgerStore, txDB *storage.CloneCache,
	tx *types.Transaction, block *types.Block, createGasPrice, deployUintCodePrice uint64) ExecutionResult {
	var result ExecutionResult
	deploy := tx.Payload.(*payload.DeployCode)
	address := types.AddressFromVmCode(deploy.Code)

	result.charge = tx.GasPrice != 0

	if result.charge {
		// init smart contract configuration info
		config := &smartcontract.Config{
			Time:   block.Header.Timestamp,
			Height: block.Header.Height,
			Tx:     tx,
		}

		gasLimit := createGasPrice + calcGasByCodeLen(len(deploy.Code), deployUintCodePrice)
		balance, err := getBalanceFromNative(config, txDB, store, tx.Payer)
		if err != nil {
			result.status = ExecSystemError
			return result
		}

		if balance < gasLimit*tx.GasPrice {
			result.gas = balance
			result.status = ExecBalanceNotEnough
			return result
		}

		if tx.GasLimit < gasLimit {
			result.gas = tx.GasLimit * tx.GasPrice
			result.status = ExecGasLimitTooLow
			return result
		}
		result.gas = gasLimit * tx.GasPrice
	}

	log.Infof("deploy contract address:%s", address.ToHexString())
	// store contract message
	val, err := txDB.Get(scommon.ST_CONTRACT, address[:])
	if err != nil {
		result.status = ExecSystemError
		return result
	}
	if val == nil {
		txDB.Add(scommon.ST_CONTRACT, address[:], deploy)
	}

	result.status = ExecSuccess
	return result
}

type ExecutionStatus byte

const (
	ExecSuccess = ExecutionStatus(0) // execution suceess
	//ExecNeedSerial       = iota               // need serial execution, `find` api detected
	//ExecHaveGlobalLock   = iota               // global param has changed, so follow-up transaction can not execute parallel
	ExecSystemError      = iota
	ExecGasLimitTooLow   = iota
	ExecContractError    = iota
	ExecBalanceNotEnough = iota
)

type ExecutionResult struct {
	status   ExecutionStatus
	gas      uint64
	charge   bool
	notifies []*event.NotifyEventInfo
}

func handleInvokeTransaction(store store.LedgerStore, txDB *storage.CloneCache,
	tx *types.Transaction, block *types.Block, invokeUintCodeGasPrice uint64) ExecutionResult {
	var result ExecutionResult
	invoke := tx.Payload.(*payload.InvokeCode)
	code := invoke.Code
	sysTransFlag := bytes.Compare(code, ninit.COMMIT_DPOS_BYTES) == 0 || block.Header.Height == 0

	isCharge := !sysTransFlag && tx.GasPrice != 0
	result.charge = isCharge

	// init smart contract configuration info
	config := &smartcontract.Config{
		Time:   block.Header.Timestamp,
		Height: block.Header.Height,
		Tx:     tx,
	}

	var (
		oldBalance      uint64
		newBalance      uint64
		codeLenGasLimit uint64
		err             error
	)

	availableGasLimit := tx.GasLimit
	if isCharge {
		oldBalance, err = getBalanceFromNative(config, txDB, store, tx.Payer)
		if err != nil {
			// this must be system error
			result.status = ExecSystemError
			return result
		}

		minGas := neovm.MIN_TRANSACTION_GAS * tx.GasPrice

		if oldBalance < minGas {
			result.gas = oldBalance
			result.status = ExecBalanceNotEnough
			return result
		}

		codeLenGasLimit = calcGasByCodeLen(len(invoke.Code), invokeUintCodeGasPrice)

		if oldBalance < codeLenGasLimit*tx.GasPrice {
			result.gas = oldBalance
			result.status = ExecBalanceNotEnough
			return result
		}

		if tx.GasLimit < codeLenGasLimit {
			result.gas = oldBalance
			result.status = ExecGasLimitTooLow
			return result
		}

		maxAvaGasLimit := oldBalance / tx.GasPrice
		if availableGasLimit > maxAvaGasLimit {
			availableGasLimit = maxAvaGasLimit
		}
	}

	//init smart contract info
	sc := smartcontract.SmartContract{
		Config:     config,
		CloneCache: txDB,
		Store:      store,
		Gas:        availableGasLimit - codeLenGasLimit,
	}

	//start the smart contract executive function
	engine, _ := sc.NewExecuteEngine(invoke.Code)

	_, err = engine.Invoke()

	costGasLimit := availableGasLimit - sc.Gas
	if costGasLimit < neovm.MIN_TRANSACTION_GAS {
		costGasLimit = neovm.MIN_TRANSACTION_GAS
	}

	costGas := costGasLimit * tx.GasPrice
	if err != nil {
		if isCharge {
			result.gas = costGas
		}
		result.status = ExecContractError
		return result
	}

	if isCharge {
		newBalance, err = getBalanceFromNative(config, txDB, store, tx.Payer)
		if err != nil {
			// this must be system error
			result.status = ExecSystemError
			return result
		}

		if newBalance < costGas {
			result.gas = costGas
			result.status = ExecBalanceNotEnough
		}

		result.gas = costGas
	}

	result.status = ExecSuccess
	result.notifies = append(result.notifies, sc.Notifications...)

	return result
}

func partitionTransactions(txs []*types.Transaction) [][]*types.Transaction {
	var batches [][]*types.Transaction
	currBatch := make([]*types.Transaction, 0)
	currAddr := make(map[common.Address]bool)
	for _, tx := range txs {
		addresses := tx.GetSignatureAddresses()
		contain := false
		for _, addr := range addresses {
			if _, ok := currAddr[addr]; ok {
				contain = true
				break
			}
			currAddr[addr] = true
		}

		if contain {
			batches = append(batches, currBatch)
			currBatch = make([]*types.Transaction, 0)
			currAddr = make(map[common.Address]bool)

			for _, addr := range addresses {
				currAddr[addr] = true
			}
		}

		currBatch = append(currBatch, tx)
	}

	if len(currBatch) > 0 {
		batches = append(batches, currBatch)
	}

	return batches
}

func handleBlockTransactionSerial(store store.LedgerStore, stateBatch *statestore.StateBatch, block *types.Block,
	eventStore *EventStore, createGasPrice, deployUintCodePrice, invokeUintCodeGasPrice uint64) error {
	eventNotifies := make([]*event.ExecuteNotify, 0, len(block.Transactions))
	for _, tx := range block.Transactions {
		txHash := tx.Hash()
		txDB := storage.NewCloneCache(stateBatch, false) // todo: set to true when in parallel mode
		result := handleTransaction(store, txDB, block, tx,
			createGasPrice, deployUintCodePrice, invokeUintCodeGasPrice)
		if stateBatch.Error() != nil || result.status == ExecSystemError {
			return fmt.Errorf("handle transaction error. tx %s error %v", txHash.ToHexString(), stateBatch.Error())
		}

		if result.status != ExecSuccess {
			txDB.Reset()
			result.notifies = nil

			log.Debugf("handle transaction error. tx: %s, status:%d", txHash.ToHexString(), result.status)
		}

		if result.charge {
			conf := &smartcontract.Config{
				Time:   block.Header.Timestamp,
				Height: block.Header.Height,
				Tx:     tx,
			}
			notifies, err := chargeCostGas(tx.Payer, result.gas, conf, txDB, store)
			if err != nil {
				return fmt.Errorf("charge gas error. tx: %s, error: %s", txHash.ToHexString(), err.Error())
			}
			result.notifies = append(result.notifies, notifies...)
		}

		txDB.Commit()

		eventNotify := &event.ExecuteNotify{
			TxHash:      txHash,
			GasConsumed: result.gas,
			Notify:      result.notifies,
		}
		if result.status == ExecSuccess {
			eventNotify.State = event.CONTRACT_STATE_SUCCESS
		} else {
			eventNotify.State = event.CONTRACT_STATE_FAIL
		}

		eventNotifies = append(eventNotifies, eventNotify)
	}

	for _, notify := range eventNotifies {
		SaveNotify(eventStore, notify.TxHash, notify)
	}
	return nil
}

type execResult struct {
	index int
	tx    *types.Transaction
	txDB  *storage.CloneCache
	res   ExecutionResult
}

func handleBlockTransactionParallel(store store.LedgerStore, stateBatch *statestore.StateBatch, block *types.Block,
	eventStore *EventStore, createGasPrice, deployUintCodePrice, invokeUintCodeGasPrice uint64) error {
	batches := partitionTransactions(block.Transactions)
	eventNotifies := make([]*event.ExecuteNotify, 0, len(block.Transactions))
	conflict := 0
	for _, batchTxs := range batches {
		resultChan := make(chan execResult, len(batchTxs))
		for i, tx := range batchTxs {
			txDB := storage.NewCloneCache(stateBatch, true)
			go func(index int, tx *types.Transaction) {
				result := handleTransaction(store, txDB, block, tx,
					createGasPrice, deployUintCodePrice, invokeUintCodeGasPrice)
				resultChan <- execResult{
					index: index,
					tx:    tx,
					txDB:  txDB,
					res:   result,
				}
			}(i, tx)
		}

		results := make(map[int]execResult)
		curr := 0
		overlay := statestore.NewOverlayDB(stateBatch)
		for curr < len(batchTxs) {
			res := <-resultChan
			results[res.index] = res

			execResult, ok := results[curr]
			for ok {
				result := execResult.res
				tx := execResult.tx
				txDB := execResult.txDB
				txHash := tx.Hash()

				if stateBatch.Error() != nil || result.status == ExecSystemError {
					return fmt.Errorf("handle transaction error. tx %s error %v", txHash.ToHexString(), stateBatch.Error())
				}

				needRedo := false
				if txDB.FindDetected {
					needRedo = true
				} else if curr != 0 {
					// check conflict
					for _, key := range txDB.ReadSet {
						if overlay.Changed(key) {
							needRedo = true
							break
						}
					}
				}

				if needRedo {
					conflict += 1
					txDB = storage.NewCloneCache(overlay, false)
					result := handleTransaction(store, txDB, block, tx,
						createGasPrice, deployUintCodePrice, invokeUintCodeGasPrice)
					if stateBatch.Error() != nil || result.status == ExecSystemError {
						return fmt.Errorf("handle transaction error. tx %s error %v", txHash.ToHexString(), stateBatch.Error())
					}
				}

				if result.status != ExecSuccess {
					txDB.Reset()
					result.notifies = nil

					log.Debugf("handle transaction error. tx: %s, status:%d", txHash.ToHexString(), result.status)
				}

				if result.charge {
					conf := &smartcontract.Config{
						Time:   block.Header.Timestamp,
						Height: block.Header.Height,
						Tx:     tx,
					}
					notifies, err := chargeCostGas(tx.Payer, result.gas, conf, txDB, store)
					if err != nil {
						return fmt.Errorf("charge gas error. tx: %s, error: %s", txHash.ToHexString(), err.Error())
					}
					result.notifies = append(result.notifies, notifies...)
				}

				for k, v := range txDB.Cache {
					key := []byte(k)
					if v == nil {
						overlay.TryDelete(scommon.DataEntryPrefix(key[0]), key[1:])
					} else {
						overlay.TryAdd(scommon.DataEntryPrefix(key[0]), key[1:], v)
					}
				}

				eventNotify := &event.ExecuteNotify{
					TxHash:      txHash,
					GasConsumed: result.gas,
					Notify:      result.notifies,
				}
				if result.status == ExecSuccess {
					eventNotify.State = event.CONTRACT_STATE_SUCCESS
				} else {
					eventNotify.State = event.CONTRACT_STATE_FAIL
				}

				eventNotifies = append(eventNotifies, eventNotify)

				curr += 1
				execResult, ok = results[curr]
			}
		}

		overlay.CommitTo()
	}

	log.Infof("parallel execution stats: %d tx, %d batch, %d conflict\n", len(block.Transactions), len(batches), conflict)

	for _, notify := range eventNotifies {
		SaveNotify(eventStore, notify.TxHash, notify)
	}
	return nil
}

func handleTransaction(store store.LedgerStore, txDB *storage.CloneCache, block *types.Block, tx *types.Transaction,
	createGasPrice, deployUintCodePrice, invokeUintCodeGasPrice uint64) ExecutionResult {
	var result ExecutionResult
	switch tx.TxType {
	case types.Deploy:
		result = handleDeployTransaction(store, txDB, tx, block, createGasPrice, deployUintCodePrice)
	case types.Invoke:
		result = handleInvokeTransaction(store, txDB, tx, block, invokeUintCodeGasPrice)
	}

	return result
}

func SaveNotify(eventStore scommon.EventStore, txHash common.Uint256, notify *event.ExecuteNotify) error {
	if !config.DefConfig.Common.EnableEventLog {
		return nil
	}
	if err := eventStore.SaveEventNotifyByTx(txHash, notify); err != nil {
		return fmt.Errorf("SaveEventNotifyByTx error %s", err)
	}
	event.PushSmartCodeEvent(txHash, 0, event.EVENT_NOTIFY, notify)
	return nil
}

func genNativeTransferCode(from, to common.Address, value uint64) []byte {
	transfer := ont.Transfers{States: []*ont.State{{From: from, To: to, Value: value}}}
	tr := new(bytes.Buffer)
	transfer.Serialize(tr)
	return tr.Bytes()
}

func chargeCostGas(payer common.Address, gas uint64, config *smartcontract.Config,
	cache *storage.CloneCache, store store.LedgerStore) ([]*event.NotifyEventInfo, error) {

	params := genNativeTransferCode(payer, utils.GovernanceContractAddress, gas)

	sc := smartcontract.SmartContract{
		Config:     config,
		CloneCache: cache,
		Store:      store,
		Gas:        math.MaxUint64,
	}

	service, _ := sc.NewNativeService()
	_, err := service.NativeCall(utils.OngContractAddress, "transfer", params)
	if err != nil {
		return nil, err
	}
	return sc.Notifications, nil
}

func refreshGlobalParam(config *smartcontract.Config, cache *storage.CloneCache, store store.LedgerStore) error {
	bf := new(bytes.Buffer)
	if err := utils.WriteVarUint(bf, uint64(len(neovm.GAS_TABLE_KEYS))); err != nil {
		return fmt.Errorf("write gas_table_keys length error:%s", err)
	}
	for _, value := range neovm.GAS_TABLE_KEYS {
		if err := serialization.WriteString(bf, value); err != nil {
			return fmt.Errorf("serialize param name error:%s", value)
		}
	}

	sc := smartcontract.SmartContract{
		Config:     config,
		CloneCache: cache,
		Store:      store,
		Gas:        math.MaxUint64,
	}

	service, _ := sc.NewNativeService()
	result, err := service.NativeCall(utils.ParamContractAddress, "getGlobalParam", bf.Bytes())
	if err != nil {
		return err
	}
	params := new(global_params.Params)
	if err := params.Deserialize(bytes.NewBuffer(result.([]byte))); err != nil {
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

func getBalanceFromNative(config *smartcontract.Config, cache *storage.CloneCache, store store.LedgerStore, address common.Address) (uint64, error) {
	bf := new(bytes.Buffer)
	if err := utils.WriteAddress(bf, address); err != nil {
		return 0, err
	}
	sc := smartcontract.SmartContract{
		Config:     config,
		CloneCache: cache,
		Store:      store,
		Gas:        math.MaxUint64,
	}

	service, _ := sc.NewNativeService()
	result, err := service.NativeCall(utils.OngContractAddress, ont.BALANCEOF_NAME, bf.Bytes())
	if err != nil {
		return 0, err
	}
	return ntypes.BigIntFromBytes(result.([]byte)).Uint64(), nil
}

func calcGasByCodeLen(codeLen int, codeGas uint64) uint64 {
	return uint64(codeLen/neovm.PER_UNIT_CODE_LEN) * codeGas
}
