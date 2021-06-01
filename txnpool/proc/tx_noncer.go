// Copyright (C) 2021 The Ontology Authors
// Copyright 2019 The go-ethereum Authors
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

package proc

import (
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/smartcontract/storage"
	"sync"

	ethcomm "github.com/ethereum/go-ethereum/common"
	//"github.com/ethereum/go-ethereum/core/state"
)

// txNoncer is a tiny virtual state database to manage the executable nonces of
// accounts in the pool, falling back to reading from a real state database if
// an account is unknown.
type txNoncer struct {
	fallback *storage.CacheDB
	nonces   map[common.Address]uint64
	lock     sync.Mutex
}

// newTxNoncer creates a new virtual state database to track the pool nonces.
func newTxNoncer(cachedb *storage.CacheDB) *txNoncer {
	return &txNoncer{
		// do we really need  copy of stateDB???
		fallback: cachedb,
		nonces:   make(map[common.Address]uint64),
	}
}

// get returns the current nonce of an account, falling back to a real state
// database if the account is unknown.
func (txn *txNoncer) get(addr common.Address) uint64 {
	// We use mutex for get operation is the underlying
	// state will mutate db even for read access.
	txn.lock.Lock()
	defer txn.lock.Unlock()

	if _, ok := txn.nonces[addr]; !ok {
		//txn.nonces[addr] = txn.fallback.GetNonce()
		ethacct, err := txn.fallback.GetEthAccount(ethcomm.BytesToAddress(addr[:]))
		if err != nil {
			log.Error(err)
			//todo return the default nonce???
			txn.nonces[addr] = 0
		} else {
			txn.nonces[addr] = ethacct.Nonce
		}
	}
	return txn.nonces[addr]
}

// set inserts a new virtual nonce into the virtual state database to be returned
// whenever the pool requests it instead of reaching into the real state database.
func (txn *txNoncer) set(addr common.Address, nonce uint64) {
	txn.lock.Lock()
	defer txn.lock.Unlock()

	txn.nonces[addr] = nonce
}

// setIfLower updates a new virtual nonce into the virtual state database if the
// the new one is lower.
func (txn *txNoncer) setIfLower(addr common.Address, nonce uint64) {
	txn.lock.Lock()
	defer txn.lock.Unlock()

	if _, ok := txn.nonces[addr]; !ok {
		//txn.nonces[addr] = txn.fallback.GetNonce(ethcomm.BytesToAddress(addr[:]))
		ethacct, err := txn.fallback.GetEthAccount(ethcomm.BytesToAddress(addr[:]))
		if err != nil {
			log.Error(err)
			//todo return the default nonce???
			txn.nonces[addr] = 0
		} else {
			txn.nonces[addr] = ethacct.Nonce
		}
	}
	if txn.nonces[addr] <= nonce {
		return
	}
	txn.nonces[addr] = nonce
}
