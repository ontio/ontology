package storage

import (
	"fmt"
	"io"
	"math/big"

	"github.com/ontio/ontology/core/store/overlaydb"

	"github.com/ethereum/go-ethereum/crypto"
	comm "github.com/ontio/ontology/common"
	common2 "github.com/ontio/ontology/core/store/common"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type OngBalanceHandle interface {
	SubBalance(common.Address, *big.Int)
	AddBalance(common.Address, *big.Int)
	GetBalance(common.Address) *big.Int
	SetBalance(common.Address, *big.Int)
}

type StateDB struct {
	cacheDB      *CacheDB
	suicided     map[common.Address]bool
	logs         []*types.Log
	thash, bhash common.Hash
	txIndex      int
	refund       uint64
	dbErr        error
	snapshots    []*snapshot
	OngBalanceHandle
}

func NewStateDB(cacheDB *CacheDB, thash, bhash common.Hash, txIndex int, balanceHandle OngBalanceHandle) *StateDB {
	return &StateDB{
		cacheDB:          cacheDB,
		suicided:         make(map[common.Address]bool),
		logs:             nil,
		thash:            thash,
		bhash:            bhash,
		txIndex:          txIndex,
		refund:           0,
		dbErr:            nil,
		snapshots:        nil,
		OngBalanceHandle: balanceHandle,
	}
}

type snapshot struct {
	changes  *overlaydb.MemDB
	suicided map[common.Address]bool
	logsSize int
	refund   uint64
}

func (s *StateDB) AddRefund(gas uint64) {
	s.refund += gas
}

// SubRefund removes gas from the refund counter.
// This method will panic if the refund counter goes below zero
func (s *StateDB) SubRefund(gas uint64) {
	if gas > s.refund {
		panic(fmt.Sprintf("Refund counter below zero (gas: %d > refund: %d)", gas, s.refund))
	}

	s.refund -= gas
}

func genKey(contract common.Address, key common.Hash) []byte {
	var result []byte
	result = append(result, contract.Bytes()...)
	result = append(result, key.Bytes()...)
	return result
}

func (s *StateDB) GetState(contract common.Address, key common.Hash) common.Hash {
	val, err := s.cacheDB.Get(genKey(contract, key))
	if err != nil {
		s.dbErr = err
	}

	return common.BytesToHash(val)
}

// GetRefund returns the current value of the refund counter.
func (s *StateDB) GetRefund() uint64 {
	return s.refund
}

func (s *StateDB) SetState(contract common.Address, key, value common.Hash) {
	s.cacheDB.Put(genKey(contract, key), value[:])
}

func (s *StateDB) GetCommittedState(addr common.Address, key common.Hash) common.Hash {
	val, err := s.cacheDB.backend.Get(genKey(addr, key))
	if err != nil {
		s.dbErr = err
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

func (self *CacheDB) GetEthCode(codeHash common.Hash) (val []byte, err error) {
	return self.get(common2.ST_ETH_CODE, codeHash[:])
}

func (self *CacheDB) PutEthCode(codeHash common.Hash, val []byte) {
	self.put(common2.ST_ETH_CODE, codeHash[:], val)
}

func (self *StateDB) getEthAccount(addr common.Address) (val EthAcount) {
	account, err := self.cacheDB.GetEthAccount(addr)
	if err != nil {
		self.dbErr = err
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
		self.dbErr = err
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
	self.suicided[addr] = true
	self.OngBalanceHandle.SetBalance(addr, big.NewInt(0))
	return true
}

func (self *StateDB) HasSuicided(addr common.Address) bool {
	return self.suicided[addr]
}

func (self *StateDB) Exist(addr common.Address) bool {
	if self.suicided[addr] {
		return true
	}
	acct := self.getEthAccount(addr)
	if !acct.IsEmpty() || self.GetBalance(addr).Sign() > 0 {
		return true
	}

	return false
}

func (self *StateDB) Empty(addr common.Address) bool {
	acct := self.getEthAccount(addr)

	return acct.IsEmpty() && self.GetBalance(addr).Sign() == 0
}

func (self *StateDB) AddLog(log *types.Log) {
	log.TxHash = self.thash
	log.BlockHash = self.bhash
	log.TxIndex = uint(self.txIndex)
	log.Index = uint(len(self.logs))
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
	for k, v := range self.suicided {
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
	self.suicided = sn.suicided
	self.refund = sn.refund
	self.logs = self.logs[:sn.logsSize]
}
