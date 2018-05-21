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

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/core/genesis"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/core/store"
	scommon "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/store/statestore"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/service/native/governance"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	neovm "github.com/ontio/ontology/smartcontract/service/neovm"
	sstates "github.com/ontio/ontology/smartcontract/states"
	"github.com/ontio/ontology/smartcontract/storage"
	stypes "github.com/ontio/ontology/smartcontract/types"
	vmtype "github.com/ontio/ontology/smartcontract/types"
)

//HandleDeployTransaction deal with smart contract deploy transaction
func (self *StateStore) HandleDeployTransaction(stateBatch *statestore.StateBatch, tx *types.Transaction) error {
	deploy := tx.Payload.(*payload.DeployCode)

	originAddress := deploy.Code.AddressFromVmCode()

	// mapping native contract origin address to target address
	if deploy.Code.VmType == stypes.Native {
		targetAddress, err := common.AddressParseFromBytes(deploy.Code.Code)
		if err != nil {
			return fmt.Errorf("Invalid native contract address:%v", err)

		}
		originAddress = targetAddress
	}

	// store contract message
	if err := stateBatch.TryGetOrAdd(
		scommon.ST_CONTRACT,
		originAddress[:],
		deploy); err != nil {
		return fmt.Errorf("TryGetOrAdd contract error %s", err)
	}
	return nil
}

//HandleInvokeTransaction deal with smart contract invoke transaction
func (self *StateStore) HandleInvokeTransaction(store store.LedgerStore, stateBatch *statestore.StateBatch, tx *types.Transaction, block *types.Block, eventStore scommon.EventStore) error {
	invoke := tx.Payload.(*payload.InvokeCode)
	txHash := tx.Hash()

	sysTransFlag := bytes.Compare(invoke.Code.Code, governance.COMMIT_DPOS_BYTES) == 0 || bytes.Compare(invoke.Code.Code, governance.INIT_CONFIG_BYTES) == 0

	if !sysTransFlag {
		// check payer ong balance
		balance, err := GetBalance(stateBatch, tx.Payer, genesis.OngContractAddress)
		if err != nil {
			return err
		}
		if balance < tx.GasLimit*tx.GasPrice {
			return fmt.Errorf("payer gas insufficient, need %d , only have %d", tx.GasLimit*tx.GasPrice, balance)
		}
	}

	// init smart contract configuration info
	config := &smartcontract.Config{
		Time:   block.Header.Timestamp,
		Height: block.Header.Height,
		Tx:     tx,
	}

	cache := storage.NewCloneCache(stateBatch)
	//init smart contract info
	sc := smartcontract.SmartContract{
		Config:     config,
		CloneCache: cache,
		Store:      store,
		Code:       invoke.Code,
		Gas:        tx.GasLimit - neovm.TRANSACTION_GAS,
	}

	//start the smart contract executive function
	_, err := sc.Execute()

	if !sysTransFlag {
		totalGas := (tx.GasLimit - sc.Gas) * tx.GasPrice
		nativeTransferCode := genNativeTransferCode(genesis.OngContractAddress, tx.Payer,
			genesis.GovernanceContractAddress, totalGas)
		transContract := smartcontract.SmartContract{
			Config:     config,
			CloneCache: cache,
			Store:      store,
			Code:       nativeTransferCode,
			Gas:        math.MaxUint64,
		}
		if err != nil {
			cache = storage.NewCloneCache(stateBatch)
			transContract.CloneCache = cache
			if _, err := transContract.Execute(); err != nil {
				return err
			}
			cache.Commit()
			if err := saveNotify(eventStore, txHash, []*event.NotifyEventInfo{}, false); err != nil {
				return err
			}
			return err
		}
		if _, err := transContract.Execute(); err != nil {
			return err
		}
		if err := saveNotify(eventStore, txHash, sc.Notifications, true); err != nil {
			return err
		}
	} else {
		if err != nil {
			if err := saveNotify(eventStore, txHash, []*event.NotifyEventInfo{}, false); err != nil {
				return err
			}
			return err
		}
		if err := saveNotify(eventStore, txHash, []*event.NotifyEventInfo{}, true); err != nil {
			return err
		}
	}
	sc.CloneCache.Commit()

	return nil
}

func saveNotify(eventStore scommon.EventStore, txHash common.Uint256, notifies []*event.NotifyEventInfo, execSucc bool) error {
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

//HandleClaimTransaction deal with ong claim transaction
func (self *StateStore) HandleClaimTransaction(stateBatch *statestore.StateBatch, tx *types.Transaction) error {
	//TODO
	return nil
}

//HandleVoteTransaction deal with vote transaction
func (self *StateStore) HandleVoteTransaction(stateBatch *statestore.StateBatch, tx *types.Transaction) error {
	vote := tx.Payload.(*payload.Vote)
	buf := new(bytes.Buffer)
	vote.Account.Serialize(buf)
	stateBatch.TryAdd(scommon.ST_VOTE, buf.Bytes(), &states.VoteState{PublicKeys: vote.PubKeys})
	return nil
}

func genNativeTransferCode(contract, from, to common.Address, value uint64) vmtype.VmCode {
	transfer := ont.Transfers{States: []*ont.State{{From: from, To: to, Value: value}}}
	tr := new(bytes.Buffer)
	transfer.Serialize(tr)
	trans := &sstates.Contract{
		Address: contract,
		Method:  "transfer",
		Args:    tr.Bytes(),
	}
	ts := new(bytes.Buffer)
	trans.Serialize(ts)
	return vmtype.VmCode{Code: ts.Bytes(), VmType: vmtype.Native}

}

func GetBalance(stateBatch *statestore.StateBatch, address, contract common.Address) (uint64, error) {
	bl, err := stateBatch.TryGet(scommon.ST_STORAGE, append(contract[:], address[:]...))
	if err != nil {
		return 0, err
	}
	if bl == nil || bl.Value == nil {
		return 0, err
	}
	item, ok := bl.Value.(*states.StorageItem)
	if !ok {
		return 0, fmt.Errorf("%s", "[GetStorageItem] instance doesn't StorageItem!")
	}
	balance, err := serialization.ReadUint64(bytes.NewBuffer(item.Value))
	if err != nil {
		return 0, err
	}
	return balance, nil
}
