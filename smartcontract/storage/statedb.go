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
	suicided         map[common.Address]bool
	logs             []*types.StorageLog
	thash, bhash     common.Hash
	txIndex          int
	refund           uint64
	dbErr            error
	snapshots        []*snapshot
	OngBalanceHandle OngBalanceHandle
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

func (s *StateDB) Prepare(thash, bhash common.Hash, ti int) {
	s.thash = thash
	s.bhash = bhash
	s.txIndex = ti
	//	s.accessList = newAccessList()
}

func (s *StateDB) DbErr() error {
	return s.dbErr
}

func (s *StateDB) BlockHash() common.Hash {
	return s.bhash
}
func (s *StateDB) TxIndex() int {
	return s.txIndex
}

func (s *StateDB) GetLogs() []*types.StorageLog {
	return s.logs
}

func (s *StateDB) Finalise() error {
	if s.dbErr != nil {
		return s.dbErr
	}

	s.cacheDB.Commit()
	return nil
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
	err := self.OngBalanceHandle.SetBalance(self.cacheDB, comm.Address(addr), big.NewInt(0))
	if err != nil {
		self.dbErr = err
	}
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
	balance, err := self.OngBalanceHandle.GetBalance(self.cacheDB, comm.Address(addr))
	if err != nil {
		self.dbErr = err
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
		self.dbErr = err
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

func (self *StateDB) SubBalance(addr common.Address, val *big.Int) {
	err := self.OngBalanceHandle.SubBalance(self.cacheDB, comm.Address(addr), val)
	if err != nil {
		self.dbErr = err
		return
	}
}

func (self *StateDB) AddBalance(addr common.Address, val *big.Int) {
	err := self.OngBalanceHandle.AddBalance(self.cacheDB, comm.Address(addr), val)
	if err != nil {
		self.dbErr = err
		return
	}
}

func (self *StateDB) GetBalance(addr common.Address) *big.Int {
	balance, err := self.OngBalanceHandle.GetBalance(self.cacheDB, comm.Address(addr))
	if err != nil {
		self.dbErr = err
		return big.NewInt(0)
	}

	return balance
}
