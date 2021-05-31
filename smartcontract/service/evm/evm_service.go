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
package evm

import (
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/store"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/context"
	"github.com/ontio/ontology/smartcontract/service/native/ong"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/ontio/ontology/smartcontract/storage"
	"github.com/ontio/ontology/vm/evm"
	"github.com/ontio/ontology/vm/evm/params"

	ethcom "github.com/ethereum/go-ethereum/common"
)

type EvmService struct {
	Store      store.LedgerStore
	Code       []byte
	CacheDB    *storage.CacheDB
	Tx         *types.Transaction
	Time       uint32
	Height     uint32
	BlockHash  common.Uint256
	ContextRef context.ContextRef
}

func (this *EvmService) Invoke() (interface{}, error) {
	config := params.MainnetChainConfig //todo use config based on network
	txhash := this.Tx.Hash()
	thash := ethcom.BytesToHash(txhash[:])
	statedb := storage.NewStateDB(this.CacheDB, thash, ethcom.BytesToHash(this.BlockHash[:]),ong.OngBalanceHandle{})
	usedGas := uint64(0)

	block, err := this.Store.GetBlockByHash(this.BlockHash)
	if err != nil {
		return nil, err
	}
	eiptx, err := this.Tx.GetEIP155Tx()
	if err != nil {
		return nil, err
	}
	res, _, err := ApplyTransaction(config, this.Store, statedb, block.Header, eiptx, &usedGas, utils.GovernanceContractAddress, evm.Config{})
	this.ContextRef.CheckUseGas(usedGas)

	return res, err
}
