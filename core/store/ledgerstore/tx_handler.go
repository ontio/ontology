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
	"math/big"
)

//HandleDeployTransaction deal with smart contract deploy transaction
func (self *StateStore) HandleDeployTransaction(store store.LedgerStore, stateBatch *statestore.StateBatch,
	tx *types.Transaction, block *types.Block, eventStore scommon.EventStore) error {
	deploy := tx.Payload.(*payload.DeployCode)
	txHash := tx.Hash()
	address := types.AddressFromVmCode(deploy.Code)
	var (
		notifies []*event.NotifyEventInfo
		err      error
	)

	if tx.GasPrice != 0 {
		// init smart contract configuration info
		config := &smartcontract.Config{
			Time:   block.Header.Timestamp,
			Height: block.Header.Height,
			Tx:     tx,
		}
		cache := storage.NewCloneCache(stateBatch)
		gasLimit := neovm.GAS_TABLE[neovm.CONTRACT_CREATE_NAME] + calcGasByCodeLen(len(deploy.Code), neovm.GAS_TABLE[neovm.UINT_DEPLOY_CODE_LEN_NAME])
		balance, err := isBalanceSufficient(tx.Payer, cache, config, store, gasLimit*tx.GasPrice)
		if err != nil {
			if err := costInvalidGas(tx.Payer, balance, config, stateBatch, store, eventStore, txHash); err != nil {
				return err
			}
			return err
		}
		if tx.GasLimit < gasLimit {
			log.Errorf("gasLimit insufficient, need:%d actual:%d", gasLimit, tx.GasLimit)
			if err := costInvalidGas(tx.Payer, tx.GasLimit*tx.GasPrice, config, stateBatch, store, eventStore, txHash); err != nil {
				return err
			}
		}
		notifies, err = costGas(tx.Payer, gasLimit*tx.GasPrice, config, cache, store)
		if err != nil {
			return err
		}
		cache.Commit()
	}

	log.Infof("deploy contract address:%s", address.ToHexString())
	// store contract message
	err = stateBatch.TryGetOrAdd(scommon.ST_CONTRACT, address[:], deploy)
	if err != nil {
		return err
	}

	SaveNotify(eventStore, txHash, notifies, true)
	return nil
}

//HandleInvokeTransaction deal with smart contract invoke transaction
func (self *StateStore) HandleInvokeTransaction(store store.LedgerStore, stateBatch *statestore.StateBatch,
	tx *types.Transaction, block *types.Block, eventStore scommon.EventStore) error {
	invoke := tx.Payload.(*payload.InvokeCode)
	txHash := tx.Hash()
	code := invoke.Code
	sysTransFlag := bytes.Compare(code, ninit.COMMIT_DPOS_BYTES) == 0 || block.Header.Height == 0

	isCharge := !sysTransFlag && tx.GasPrice != 0

	// init smart contract configuration info
	config := &smartcontract.Config{
		Time:   block.Header.Timestamp,
		Height: block.Header.Height,
		Tx:     tx,
	}

	var (
		codeLenGas uint64
		gasLimit   uint64
		gas        uint64
		balance    uint64
		err        error
	)
	cache := storage.NewCloneCache(stateBatch)
	if isCharge {
		codeLenGas = calcGasByCodeLen(len(invoke.Code), neovm.GAS_TABLE[neovm.UINT_INVOKE_CODE_LEN_NAME])
		balance, err := isBalanceSufficient(tx.Payer, cache, config, store, gasLimit*tx.GasPrice)
		if err != nil {
			if err := costInvalidGas(tx.Payer, balance, config, stateBatch, store, eventStore, txHash); err != nil {
				return err
			}
			return err
		}

		if tx.GasLimit < codeLenGas {
			if err := costInvalidGas(tx.Payer, tx.GasLimit*tx.GasPrice, config, stateBatch, store, eventStore, txHash); err != nil {
				return err
			}
			return fmt.Errorf("transaction gas: %d less than code length gas: %d", tx.GasLimit, codeLenGas)
		}
	}

	//init smart contract info
	sc := smartcontract.SmartContract{
		Config:     config,
		CloneCache: cache,
		Store:      store,
		Gas:        tx.GasLimit - codeLenGas,
	}

	//start the smart contract executive function
	engine, _ := sc.NewExecuteEngine(invoke.Code)

	_, err = engine.Invoke()

	if isCharge {
		gasLimit = tx.GasLimit - sc.Gas
		gas = gasLimit*tx.GasPrice
		balance, err = getBalance(config, cache, store, tx.Payer)
		if err != nil {
			return err
		}
		if balance < gas {
			if err := costInvalidGas(tx.Payer, balance, config, stateBatch, store, eventStore, txHash); err != nil {
				return err
			}
		}
	}

	if err != nil {
		if isCharge {
			if err := costInvalidGas(tx.Payer, gas, config, stateBatch, store, eventStore, txHash); err != nil {
				return err
			}
		}
		return err
	}

	var notifies []*event.NotifyEventInfo
	if isCharge {
		mixGas := neovm.MIN_TRANSACTION_GAS
		if gasLimit < mixGas {
			if balance < mixGas*tx.GasPrice {
				if err := costInvalidGas(tx.Payer, balance, config, stateBatch, store, eventStore, txHash); err != nil {
					return err
				}
			}
			gas = mixGas * tx.GasPrice
		}
		notifies, err = costGas(tx.Payer, gas, config, sc.CloneCache, store)
		if err != nil {
			return err
		}

	}

	SaveNotify(eventStore, txHash, append(sc.Notifications, notifies...), true)
	sc.CloneCache.Commit()
	return nil
}

func SaveNotify(eventStore scommon.EventStore, txHash common.Uint256, notifies []*event.NotifyEventInfo, execSucc bool) error {
	if !config.DefConfig.Common.EnableEventLog {
		return nil
	}
	var notifyInfo *event.ExecuteNotify
	if execSucc {
		notifyInfo = &event.ExecuteNotify{TxHash: txHash,
			State: event.CONTRACT_STATE_SUCCESS, Notify: notifies}
	} else {
		notifyInfo = &event.ExecuteNotify{TxHash: txHash,
			State: event.CONTRACT_STATE_FAIL, Notify: notifies}
	}
	if err := eventStore.SaveEventNotifyByTx(txHash, notifyInfo); err != nil {
		return fmt.Errorf("SaveEventNotifyByTx error %s", err)
	}
	event.PushSmartCodeEvent(txHash, 0, event.EVENT_NOTIFY, notifyInfo)
	return nil
}

func genNativeTransferCode(from, to common.Address, value uint64) []byte {
	transfer := ont.Transfers{States: []*ont.State{{From: from, To: to, Value: value}}}
	tr := new(bytes.Buffer)
	transfer.Serialize(tr)
	return tr.Bytes()
}

// check whether payer ong balance sufficient
func isBalanceSufficient(payer common.Address, cache *storage.CloneCache, config *smartcontract.Config, store store.LedgerStore, gas uint64) (uint64, error) {
	balance, err := getBalance(config, cache, store, payer)
	if err != nil {
		return 0, err
	}
	if balance < gas {
		return 0, fmt.Errorf("payer gas insufficient, need %d , only have %d", gas, balance)
	}
	return balance, nil
}

func costGas(payer common.Address, gas uint64, config *smartcontract.Config,
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

	for k, _ := range neovm.GAS_TABLE {
		n, ps := params.GetParam(k)
		if n != -1 && ps.Value != "" {
			pu, err := strconv.ParseUint(ps.Value, 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse uint %v", err)
			}
			neovm.GAS_TABLE[k] = pu
		}
	}
	return nil
}

func getBalance(config *smartcontract.Config, cache *storage.CloneCache, store store.LedgerStore, address common.Address) (uint64, error) {
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
	return new(big.Int).SetBytes(result.([]byte)).Uint64(), nil
}

func costInvalidGas(address common.Address, gas uint64, config *smartcontract.Config, stateBatch *statestore.StateBatch,
	store store.LedgerStore, eventStore scommon.EventStore, txHash common.Uint256) error {
	cache := storage.NewCloneCache(stateBatch)
	notifies, err := costGas(address, gas, config, cache, store)
	if err != nil {
		return err
	}
	cache.Commit()
	SaveNotify(eventStore, txHash, notifies, false)
	return nil
}

func calcGasByCodeLen(codeLen int, codeGas uint64) uint64 {
	return uint64(codeLen/neovm.PER_UNIT_CODE_LEN) * codeGas
}
