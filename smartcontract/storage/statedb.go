/*
 * Copyright (C) 2021 The ontology Authors
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

package storage

import (
	"fmt"
	"io"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	comm "github.com/ontio/ontology/common"
	common2 "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/store/overlaydb"
	"github.com/ontio/ontology/core/types"
)

type OngBalanceHandle interface {
	SubBalance(cache *CacheDB, addr comm.Address, val *big.Int) error
	AddBalance(cache *CacheDB, addr comm.Address, val *big.Int) error
	SetBalance(cache *CacheDB, addr comm.Address, val *big.Int) error
	GetBalance(cache *CacheDB, addr comm.Address) (*big.Int, error)
}

type StateDB struct {
	cacheDB          *CacheDB
	Suicided         map[common.Address]bool
	logs             []*types.StorageLog
	thash, bhash     common.Hash
	txIndex          int
	refund           uint64
	snapshots        []*snapshot
	OngBalanceHandle OngBalanceHandle
}

func NewStateDB(cacheDB *CacheDB, thash, bhash common.Hash, balanceHandle OngBalanceHandle) *StateDB {
	return &StateDB{
		cacheDB:          cacheDB,
		Suicided:         make(map[common.Address]bool),
		logs:             nil,
		thash:            thash,
		bhash:            bhash,
		refund:           0,
		snapshots:        nil,
		OngBalanceHandle: balanceHandle,
	}
}

func (self *StateDB) Prepare(thash, bhash common.Hash) {
	self.thash = thash
	self.bhash = bhash
	//	s.accessList = newAccessList()
}

func (self *StateDB) DbErr() error {
	return self.cacheDB.backend.Error()
}

func (self *StateDB) BlockHash() common.Hash {
	return self.bhash
}

func (self *StateDB) GetLogs() []*types.StorageLog {
	return self.logs
}

func (self *StateDB) Commit() error {
	err := self.CommitToCacheDB()
	if err != nil {
		return err
	}
	self.cacheDB.Commit()
	return nil
}

func (self *StateDB) CommitToCacheDB() error {
	for addr := range self.Suicided {
		self.cacheDB.DelEthAccount(addr) //todo : check consistence with ethereum
		err := self.cacheDB.CleanContractStorageData(comm.Address(addr))
		if err != nil {
			return err
		}
	}

	self.Suicided = make(map[common.Address]bool)
	self.snapshots = self.snapshots[:0]

	return nil
}

type snapshot struct {
	changes  *overlaydb.MemDB
	suicided map[common.Address]bool
	logsSize int
	refund   uint64
}

func (self *StateDB) AddRefund(gas uint64) {
	self.refund += gas
}

// SubRefund removes gas from the refund counter.
// This method will panic if the refund counter goes below zero
func (self *StateDB) SubRefund(gas uint64) {
	if gas > self.refund {
		panic(fmt.Sprintf("Refund counter below zero (gas: %d > refund: %d)", gas, self.refund))
	}

	self.refund -= gas
}

func genKey(contract common.Address, key common.Hash) []byte {
	var result []byte
	result = append(result, contract.Bytes()...)
	result = append(result, key.Bytes()...)
	return result
}

func (self *StateDB) GetState(contract common.Address, key common.Hash) common.Hash {
	val, err := self.cacheDB.Get(genKey(contract, key))
	if err != nil {
		self.cacheDB.SetDbErr(err)
	}

	return common.BytesToHash(val)
}

// GetRefund returns the current value of the refund counter.
func (self *StateDB) GetRefund() uint64 {
	return self.refund
}

func (self *StateDB) SetState(contract common.Address, key, value common.Hash) {
	self.cacheDB.Put(genKey(contract, key), value[:])
}

func (self *StateDB) GetCommittedState(addr common.Address, key common.Hash) common.Hash {
	k := self.cacheDB.GenAccountStateKey(comm.Address(addr), key[:])
	val, err := self.cacheDB.backend.Get(k)
	if err != nil {
		self.cacheDB.SetDbErr(err)
	}

	return common.BytesToHash(val)
}

type EthAcount struct {
	Nonce    uint64
	CodeHash common.Hash
}

func (self *EthAcount) IsEmpty() bool {
	return self.Nonce == 0 && self.CodeHash == common.Hash{}
}

func (self *EthAcount) Serialization(sink *comm.ZeroCopySink) {
	sink.WriteUint64(self.Nonce)
	sink.WriteHash(comm.Uint256(self.CodeHash))
}

func (self *EthAcount) Deserialization(source *comm.ZeroCopySource) error {
	nonce, _ := source.NextUint64()
	hash, eof := source.NextHash()
	if eof {
		return io.ErrUnexpectedEOF
	}
	self.Nonce = nonce
	self.CodeHash = common.Hash(hash)

	return nil
}

func (self *CacheDB) GetEthAccount(addr common.Address) (val EthAcount, err error) {
	value, err := self.get(common2.ST_ETH_ACCOUNT, addr[:])
	if err != nil {
		return val, err
	}

	if len(value) == 0 {
		return val, nil
	}

	err = val.Deserialization(comm.NewZeroCopySource(value))

	return val, err
}

func (self *CacheDB) PutEthAccount(addr common.Address, val EthAcount) {
	var raw []byte
	if !val.IsEmpty() {
		raw = comm.SerializeToBytes(&val)
	}

	self.put(common2.ST_ETH_ACCOUNT, addr[:], raw)
}

func (self *CacheDB) DelEthAccount(addr common.Address) {
	self.put(common2.ST_ETH_ACCOUNT, addr[:], nil)
}

func (self *CacheDB) GetEthCode(codeHash common.Hash) (val []byte, err error) {
	return self.get(common2.ST_ETH_CODE, codeHash[:])
}

func (self *CacheDB) PutEthCode(codeHash common.Hash, val []byte) {
	self.put(common2.ST_ETH_CODE, codeHash[:], val)
}

func (self *StateDB) getEthAccount(addr common.Address) (val EthAcount) {
	account, err := self.cacheDB.GetEthAccount(addr)
	if err != nil {
		self.cacheDB.SetDbErr(err)
		return val
	}

	return account
}

func (self *StateDB) GetNonce(addr common.Address) uint64 {
	return self.getEthAccount(addr).Nonce
}

func (self *StateDB) SetNonce(addr common.Address, nonce uint64) {
	account := self.getEthAccount(addr)
	account.Nonce = nonce
	self.cacheDB.PutEthAccount(addr, account)
}

func (self *StateDB) GetCodeHash(addr common.Address) (hash common.Hash) {
	return self.getEthAccount(addr).CodeHash
}

func (self *StateDB) GetCode(addr common.Address) []byte {
	hash := self.GetCodeHash(addr)
	code, err := self.cacheDB.GetEthCode(hash)
	if err != nil {
		self.cacheDB.SetDbErr(err)
		return nil
	}

	return code
}

func (self *StateDB) SetCode(addr common.Address, code []byte) {
	codeHash := crypto.Keccak256Hash(code)
	account := self.getEthAccount(addr)
	account.CodeHash = codeHash
	self.cacheDB.PutEthAccount(addr, account)
	self.cacheDB.PutEthCode(codeHash, code)
}

func (self *StateDB) GetCodeSize(addr common.Address) int {
	// todo : add cache to speed up
	return len(self.GetCode(addr))
}

func (self *StateDB) Suicide(addr common.Address) bool {
	acct := self.getEthAccount(addr)
	if acct.IsEmpty() {
		return false
	}
	self.Suicided[addr] = true
	err := self.OngBalanceHandle.SetBalance(self.cacheDB, comm.Address(addr), big.NewInt(0))
	if err != nil {
		self.cacheDB.SetDbErr(err)
	}
	return true
}

func (self *StateDB) HasSuicided(addr common.Address) bool {
	return self.Suicided[addr]
}

func (self *StateDB) Exist(addr common.Address) bool {
	if self.Suicided[addr] {
		return true
	}
	acct := self.getEthAccount(addr)
	balance, err := self.OngBalanceHandle.GetBalance(self.cacheDB, comm.Address(addr))
	if err != nil {
		self.cacheDB.SetDbErr(err)
		return false
	}
	if !acct.IsEmpty() || balance.Sign() > 0 {
		return true
	}

	return false
}

func (self *StateDB) Empty(addr common.Address) bool {
	acct := self.getEthAccount(addr)

	balance, err := self.OngBalanceHandle.GetBalance(self.cacheDB, comm.Address(addr))
	if err != nil {
		self.cacheDB.SetDbErr(err)
		return false
	}

	return acct.IsEmpty() && balance.Sign() == 0
}

func (self *StateDB) AddLog(log *types.StorageLog) {
	self.logs = append(self.logs, log)
}

func (self *StateDB) AddPreimage(common.Hash, []byte) {
	// todo
}

func (self *StateDB) ForEachStorage(common.Address, func(common.Hash, common.Hash) bool) error {
	panic("todo")
}

func (self *StateDB) CreateAccount(address common.Address) {
	return
}

func (self *StateDB) Snapshot() int {
	changes := self.cacheDB.memdb.DeepClone()
	suicided := make(map[common.Address]bool)
	for k, v := range self.Suicided {
		suicided[k] = v
	}

	sn := &snapshot{
		changes:  changes,
		suicided: suicided,
		logsSize: len(self.logs),
		refund:   self.refund,
	}

	self.snapshots = append(self.snapshots, sn)

	return len(self.snapshots) - 1
}

func (self *StateDB) RevertToSnapshot(idx int) {
	if idx+1 > len(self.snapshots) {
		panic("can not to revert snapshot")
	}

	sn := self.snapshots[idx]

	self.snapshots = self.snapshots[:idx]
	self.cacheDB.memdb = sn.changes
	self.Suicided = sn.suicided
	self.refund = sn.refund
	self.logs = self.logs[:sn.logsSize]
}

func (self *StateDB) SubBalance(addr common.Address, val *big.Int) {
	err := self.OngBalanceHandle.SubBalance(self.cacheDB, comm.Address(addr), val)
	if err != nil {
		self.cacheDB.SetDbErr(err)
		return
	}
}

func (self *StateDB) AddBalance(addr common.Address, val *big.Int) {
	err := self.OngBalanceHandle.AddBalance(self.cacheDB, comm.Address(addr), val)
	if err != nil {
		self.cacheDB.SetDbErr(err)
		return
	}
}

func (self *StateDB) GetBalance(addr common.Address) *big.Int {
	balance, err := self.OngBalanceHandle.GetBalance(self.cacheDB, comm.Address(addr))
	if err != nil {
		self.cacheDB.SetDbErr(err)
		return big.NewInt(0)
	}

	return balance
}
