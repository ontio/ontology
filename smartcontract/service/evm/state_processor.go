// Copyright (C) 2021 The Ontology Authors
// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package evm

import (
	"math/big"

	common2 "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/store"
	otypes "github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/storage"
	"github.com/ontio/ontology/vm/evm"
	"github.com/ontio/ontology/vm/evm/params"
)

func applyTransaction(msg types.Message, statedb *storage.StateDB, blockHeight uint32, tx *types.Transaction, usedGas *uint64, evm *evm.EVM, feeReceiver common.Address) (*ExecutionResult, *otypes.Receipt, error) {
	// Create a new context to be used in the EVM environment
	txContext := NewEVMTxContext(msg)
	// Add addresses to access list if applicable
	/* todo
	if config.IsYoloV2(header.Number) {
		statedb.AddAddressToAccessList(msg.From())
		if dst := msg.To(); dst != nil {
			statedb.AddAddressToAccessList(*dst)
			// If it's a create-tx, the destination will be added inside evm.create
		}
		for _, addr := range evm.ActivePrecompiles() {
			statedb.AddAddressToAccessList(addr)
		}
	}
	*/

	// Update the evm with the new transaction context.
	evm.Reset(txContext, statedb)
	// Apply the transaction to the current state (included in the env)
	result, err := ApplyMessage(evm, msg, common2.Address(feeReceiver))
	if err != nil {
		return nil, nil, err
	}
	// flush changes to overlay db
	err = statedb.Commit()
	if err != nil {
		return nil, nil, err
	}
	*usedGas += result.UsedGas

	// Create a new receipt for the transaction, storing the intermediate root and gas used by the tx
	// based on the eip phase, we're passing whether the root touch-delete accounts.
	receipt := otypes.NewReceipt(result.Failed(), *usedGas)
	receipt.TxHash = tx.Hash()
	receipt.GasUsed = result.UsedGas
	receipt.GasPrice = tx.GasPrice().Uint64() // safe since ong should be in uint64 range
	// if the transaction created a contract, store the creation address in the receipt.
	if msg.To() == nil {
		receipt.ContractAddress = crypto.CreateAddress(evm.TxContext.Origin, tx.Nonce())
	}
	// Set the receipt logs and create a bloom for filtering
	receipt.Logs = statedb.GetLogs()
	receipt.BlockHash = statedb.BlockHash()
	receipt.BlockNumber = big.NewInt(int64(blockHeight))

	return result, receipt, err
}

// ApplyTransaction attempts to apply a transaction to the given state database
// and uses the input parameters for its environment. It returns the receipt
// for the transaction, gas used and an error if the transaction failed,
// indicating the block was invalid.
func ApplyTransaction(config *params.ChainConfig, bc store.LedgerStore, statedb *storage.StateDB, blockHeight, timestamp uint32, tx *types.Transaction, usedGas *uint64, feeReceiver common.Address, cfg evm.Config) (*ExecutionResult, *otypes.Receipt, error) {
	signer := types.NewEIP155Signer(config.ChainID)
	msg, err := tx.AsMessage(signer)
	if err != nil {
		return nil, nil, err
	}
	// Create a new context to be used in the EVM environment
	blockContext := NewEVMBlockContext(blockHeight, timestamp, bc)
	vmenv := evm.NewEVM(blockContext, evm.TxContext{}, statedb, config, cfg)
	return applyTransaction(msg, statedb, blockHeight, tx, usedGas, vmenv, feeReceiver)
}
