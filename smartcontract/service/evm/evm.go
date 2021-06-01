// Copyright (C) 2021 The Ontology Authors
// Copyright 2016 The go-ethereum Authors
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
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ontio/ontology/core/store"
	"github.com/ontio/ontology/vm/evm"
)

// NewEVMBlockContext creates a new context for use in the EVM.
func NewEVMBlockContext(height, timestamp uint32, chain store.LedgerStore) evm.BlockContext {
	return evm.BlockContext{
		CanTransfer: CanTransfer,
		Transfer:    Transfer,
		GetHash:     GetHashFn(chain),
		Coinbase:    common.Address{}, // todo
		BlockNumber: big.NewInt(int64(height)),
		Time:        big.NewInt(int64(timestamp)),
		Difficulty:  big.NewInt(0),
		GasLimit:    math.MaxUint64, // todo
	}
}

// NewEVMTxContext creates a new transaction context for a single transaction.
func NewEVMTxContext(msg Message) evm.TxContext {
	return evm.TxContext{
		Origin:   msg.From(),
		GasPrice: new(big.Int).Set(msg.GasPrice()),
	}
}

// GetHashFn returns a GetHashFunc which retrieves header hashes by number
func GetHashFn(chain store.LedgerStore) func(n uint64) common.Hash {
	return func(n uint64) common.Hash {
		hash := chain.GetBlockHash(uint32(n))
		return common.Hash(hash)
	}
}

// CanTransfer checks whether there are enough funds in the address' account to make a transfer.
// This does not take the necessary gas in to account to make the transfer valid.
func CanTransfer(db evm.StateDB, addr common.Address, amount *big.Int) bool {
	return db.GetBalance(addr).Cmp(amount) >= 0
}

// Transfer subtracts amount from sender and adds amount to recipient using the given Db
func Transfer(db evm.StateDB, sender, recipient common.Address, amount *big.Int) {
	db.SubBalance(sender, amount)
	db.AddBalance(recipient, amount)
}
