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

package neovm

import (
	"math/big"
	"strings"

	"github.com/Ontology/common"
	"github.com/Ontology/common/log"
	"github.com/Ontology/core/states"
	"github.com/Ontology/core/store"
	"github.com/Ontology/core/types"
	"github.com/Ontology/errors"
	. "github.com/Ontology/smartcontract/common"
	"github.com/Ontology/smartcontract/event"
	trigger "github.com/Ontology/smartcontract/types"
	vm "github.com/Ontology/vm/neovm"
	vmtypes "github.com/Ontology/vm/neovm/types"
	"github.com/ontio/ontology-crypto/keypair"
)

var (
	ErrDBNotFound = "leveldb: not found"
	Notify        = "Notify"
	Log           = "Log"
)

type StateReader struct {
	serviceMap    map[string]func(*vm.ExecutionEngine) (bool, error)
	trigger       trigger.TriggerType
	Notifications []*event.NotifyEventInfo
	ldgerStore    store.ILedgerStore
}

func NewStateReader(ldgerStore store.ILedgerStore, trigger trigger.TriggerType) *StateReader {
	var stateReader StateReader
	stateReader.ldgerStore = ldgerStore
	stateReader.serviceMap = make(map[string]func(*vm.ExecutionEngine) (bool, error), 0)
	stateReader.trigger = trigger

	stateReader.Register("Neo.Runtime.GetTrigger", stateReader.RuntimeGetTrigger)
	stateReader.Register("Neo.Runtime.GetTime", stateReader.RuntimeGetTime)
	stateReader.Register("Neo.Runtime.CheckWitness", stateReader.RuntimeCheckWitness)
	stateReader.Register("Neo.Runtime.Notify", stateReader.RuntimeNotify)
	stateReader.Register("Neo.Runtime.Log", stateReader.RuntimeLog)

	stateReader.Register("Neo.Blockchain.GetHeight", stateReader.BlockChainGetHeight)
	stateReader.Register("Neo.Blockchain.GetHeader", stateReader.BlockChainGetHeader)
	stateReader.Register("Neo.Blockchain.GetBlock", stateReader.BlockChainGetBlock)
	stateReader.Register("Neo.Blockchain.GetTransaction", stateReader.BlockChainGetTransaction)
	stateReader.Register("Neo.Blockchain.GetContract", stateReader.GetContract)

	stateReader.Register("Neo.Header.GetHash", stateReader.HeaderGetHash)
	stateReader.Register("Neo.Header.GetVersion", stateReader.HeaderGetVersion)
	stateReader.Register("Neo.Header.GetPrevHash", stateReader.HeaderGetPrevHash)
	stateReader.Register("Neo.Header.GetMerkleRoot", stateReader.HeaderGetMerkleRoot)
	stateReader.Register("Neo.Header.GetIndex", stateReader.HeaderGetIndex)
	stateReader.Register("Neo.Header.GetTimestamp", stateReader.HeaderGetTimestamp)
	stateReader.Register("Neo.Header.GetConsensusData", stateReader.HeaderGetConsensusData)
	stateReader.Register("Neo.Header.GetNextConsensus", stateReader.HeaderGetNextConsensus)

	stateReader.Register("Neo.Block.GetTransactionCount", stateReader.BlockGetTransactionCount)
	stateReader.Register("Neo.Block.GetTransactions", stateReader.BlockGetTransactions)
	stateReader.Register("Neo.Block.GetTransaction", stateReader.BlockGetTransaction)

	stateReader.Register("Neo.Transaction.GetHash", stateReader.TransactionGetHash)
	stateReader.Register("Neo.Transaction.GetType", stateReader.TransactionGetType)
	stateReader.Register("Neo.Transaction.GetAttributes", stateReader.TransactionGetAttributes)

	stateReader.Register("Neo.Attribute.GetUsage", stateReader.AttributeGetUsage)
	stateReader.Register("Neo.Attribute.GetData", stateReader.AttributeGetData)

	stateReader.Register("Neo.Storage.GetScript", stateReader.ContractGetCode)
	stateReader.Register("Neo.Storage.GetContext", stateReader.StorageGetContext)
	stateReader.Register("Neo.Storage.Get", stateReader.StorageGet)

	return &stateReader
}

func (s *StateReader) Register(methodName string, handler func(*vm.ExecutionEngine) (bool, error)) bool {
	s.serviceMap[methodName] = handler
	return true
}

func (s *StateReader) GetServiceMap() map[string]func(*vm.ExecutionEngine) (bool, error) {
	return s.serviceMap
}

func (s *StateReader) RuntimeGetTrigger(e *vm.ExecutionEngine) (bool, error) {
	vm.PushData(e, int(s.trigger))
	return true, nil
}

func (s *StateReader) RuntimeGetTime(e *vm.ExecutionEngine) (bool, error) {
	hash := s.ldgerStore.GetCurrentBlockHash()
	header, err := s.ldgerStore.GetHeaderByHash(hash)
	if err != nil {
		return false, errors.NewDetailErr(err, errors.ErrNoCode, "[RuntimeGetTime] GetHeader error!.")
	}

	vm.PushData(e, header.Timestamp)
	return true, nil
}

func (s *StateReader) RuntimeNotify(e *vm.ExecutionEngine) (bool, error) {
	item := vm.PopStackItem(e)
	container := e.GetCodeContainer()
	if container == nil {
		log.Error("[RuntimeNotify] Get container fail!")
		return false, errors.NewErr("[CreateAsset] Get container fail!")
	}
	tran, ok := container.(*types.Transaction)
	if !ok {
		log.Error("[RuntimeNotify] Container not transaction!")
		return false, errors.NewErr("[CreateAsset] Container not transaction!")
	}
	context, err := e.CurrentContext()
	if err != nil {
		return false, err
	}
	hash, err := context.GetCodeHash()
	if err != nil {
		return false, err
	}
	txid := tran.Hash()
	s.Notifications = append(s.Notifications, &event.NotifyEventInfo{TxHash: txid, CodeHash: hash, States: ConvertReturnTypes(item)})
	return true, nil
}

func (s *StateReader) RuntimeLog(e *vm.ExecutionEngine) (bool, error) {
	item := vm.PopByteArray(e)
	container := e.GetCodeContainer()
	if container == nil {
		log.Error("[RuntimeLog] Get container fail!")
		return false, errors.NewErr("[CreateAsset] Get container fail!")
	}
	tran, ok := container.(*types.Transaction)
	if !ok {
		log.Error("[RuntimeLog] Container not transaction!")
		return false, errors.NewErr("[CreateAsset] Container not transaction!")
	}
	context, err := e.CurrentContext()
	if err != nil {
		return false, err
	}
	hash, err := context.GetCodeHash()
	if err != nil {
		return false, err
	}
	event.PushSmartCodeEvent(tran.Hash(), 0, Log, event.LogEventArgs{tran.Hash(), hash, string(item)})
	return true, nil
}

func (s *StateReader) CheckWitnessHash(engine *vm.ExecutionEngine, address common.Address) (bool, error) {
	tx := engine.GetCodeContainer().(*types.Transaction)
	addresses := tx.GetSignatureAddresses()
	return contains(addresses, address), nil
}

func (s *StateReader) CheckWitnessPublicKey(engine *vm.ExecutionEngine, publicKey keypair.PublicKey) (bool, error) {
	return s.CheckWitnessHash(engine, types.AddressFromPubKey(publicKey))
}

func (s *StateReader) RuntimeCheckWitness(e *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(e) < 1 {
		return false, errors.NewErr("[RuntimeCheckWitness] Too few input parameters ")
	}
	data := vm.PopByteArray(e)
	var (
		result bool
		err    error
	)
	if len(data) == 20 {
		program, err := common.AddressParseFromBytes(data)
		if err != nil {
			return false, err
		}
		result, err = s.CheckWitnessHash(e, program)
	} else if len(data) == 33 {
		publicKey, err := keypair.DeserializePublicKey(data)
		if err != nil {
			return false, err
		}
		result, err = s.CheckWitnessPublicKey(e, publicKey)
	} else {
		return false, errors.NewDetailErr(err, errors.ErrNoCode, "[RuntimeCheckWitness] data invalid.")
	}
	if err != nil {
		return false, err
	}
	vm.PushData(e, result)
	return true, nil
}

func (s *StateReader) BlockChainGetHeight(e *vm.ExecutionEngine) (bool, error) {
	var i uint32
	i = s.ldgerStore.GetCurrentBlockHeight()
	vm.PushData(e, i)
	return true, nil
}

func (s *StateReader) BlockChainGetHeader(e *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(e) < 1 {
		return false, errors.NewErr("[BlockChainGetHeader] Too few input parameters ")
	}
	data := vm.PopByteArray(e)
	var (
		header *types.Header
		err    error
	)
	l := len(data)
	if l <= 5 {
		b := new(big.Int)
		height := uint32(b.SetBytes(common.BytesReverse(data)).Int64())
		hash := s.ldgerStore.GetBlockHash(height)
		header, err = s.ldgerStore.GetHeaderByHash(hash)
		if err != nil {
			return false, errors.NewDetailErr(err, errors.ErrNoCode, "[BlockChainGetHeader] GetHeader error!.")
		}

	} else if l == 32 {
		hash, _ := common.Uint256ParseFromBytes(data)
		header, err = s.ldgerStore.GetHeaderByHash(hash)
		if err != nil {
			return false, errors.NewDetailErr(err, errors.ErrNoCode, "[BlockChainGetHeader] GetHeader error!.")
		}
	} else {
		return false, errors.NewErr("[BlockChainGetHeader] data invalid.")
	}
	vm.PushData(e, header)
	return true, nil
}

func (s *StateReader) BlockChainGetBlock(e *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(e) < 1 {
		return false, errors.NewErr("[BlockChainGetBlock] Too few input parameters ")
	}
	data := vm.PopByteArray(e)
	var (
		block *types.Block
	)
	l := len(data)
	if l <= 5 {
		b := new(big.Int)
		height := uint32(b.SetBytes(common.BytesReverse(data)).Int64())
		var err error
		block, err = s.ldgerStore.GetBlockByHeight(height)
		if err != nil {
			return false, errors.NewDetailErr(err, errors.ErrNoCode, "[BlockChainGetBlock] GetBlock error!.")
		}
	} else if l == 32 {
		hash, err := common.Uint256ParseFromBytes(data)
		if err != nil {
			return false, err
		}
		block, err = s.ldgerStore.GetBlockByHash(hash)
		if err != nil {
			return false, errors.NewDetailErr(err, errors.ErrNoCode, "[BlockChainGetBlock] GetBlock error!.")
		}
	} else {
		return false, errors.NewErr("[BlockChainGetBlock] data invalid.")
	}
	vm.PushData(e, block)
	return true, nil
}

func (s *StateReader) BlockChainGetTransaction(e *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(e) < 1 {
		return false, errors.NewErr("[BlockChainGetTransaction] Too few input parameters ")
	}
	d := vm.PopByteArray(e)
	hash, err := common.Uint256ParseFromBytes(d)
	if err != nil {
		return false, err
	}
	t, _, err := s.ldgerStore.GetTransaction(hash)
	if err != nil {
		return false, errors.NewDetailErr(err, errors.ErrNoCode, "[BlockChainGetTransaction] GetTransaction error!")
	}

	vm.PushData(e, t)
	return true, nil
}

func (s *StateReader) GetContract(e *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(e) < 1 {
		return false, errors.NewErr("[GetContract] Too few input parameters ")
	}
	hashByte := vm.PopByteArray(e)
	hash, err := common.AddressParseFromBytes(hashByte)
	if err != nil {
		return false, err
	}
	item, err := s.ldgerStore.GetContractState(hash)
	if err != nil {
		return false, errors.NewDetailErr(err, errors.ErrNoCode, "[GetContract] GetAsset error!")
	}
	vm.PushData(e, item)
	return true, nil
}

func (s *StateReader) HeaderGetHash(e *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(e) < 1 {
		return false, errors.NewErr("[HeaderGetHash] Too few input parameters ")
	}
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.NewErr("[HeaderGetHash] Pop blockdata nil!")
	}
	var data *types.Header
	if b, ok := d.(*types.Block); ok {
		data = b.Header
	} else if h, ok := d.(*types.Header); ok {
		data = h
	} else {
		return false, errors.NewErr("[HeaderGetHash] Wrong type!")
	}
	h := data.Hash()
	vm.PushData(e, h.ToArray())
	return true, nil
}

func (s *StateReader) HeaderGetVersion(e *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(e) < 1 {
		return false, errors.NewErr("[HeaderGetVersion] Too few input parameters ")
	}
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.NewErr("[HeaderGetVersion] Pop blockdata nil!")
	}
	var data *types.Header
	if b, ok := d.(*types.Block); ok {
		data = b.Header
	} else if h, ok := d.(*types.Header); ok {
		data = h
	} else {
		return false, errors.NewErr("[HeaderGetVersion] Wrong type!")
	}
	vm.PushData(e, data.Version)
	return true, nil
}

func (s *StateReader) HeaderGetPrevHash(e *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(e) < 1 {
		return false, errors.NewErr("[HeaderGetPrevHash] Too few input parameters ")
	}
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.NewErr("[HeaderGetPrevHash] Pop blockdata nil!")
	}
	var data *types.Header
	if b, ok := d.(*types.Block); ok {
		data = b.Header
	} else if h, ok := d.(*types.Header); ok {
		data = h
	} else {
		return false, errors.NewErr("[HeaderGetPrevHash] Wrong type!")
	}
	vm.PushData(e, data.PrevBlockHash.ToArray())
	return true, nil
}

func (s *StateReader) HeaderGetMerkleRoot(e *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(e) < 1 {
		return false, errors.NewErr("[HeaderGetMerkleRoot] Too few input parameters ")
	}
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.NewErr("[HeaderGetMerkleRoot] Pop blockdata nil!")
	}
	var data *types.Header
	if b, ok := d.(*types.Block); ok {
		data = b.Header
	} else if h, ok := d.(*types.Header); ok {
		data = h
	} else {
		return false, errors.NewErr("[HeaderGetMerkleRoot] Wrong type!")
	}
	vm.PushData(e, data.TransactionsRoot.ToArray())
	return true, nil
}

func (s *StateReader) HeaderGetIndex(e *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(e) < 1 {
		return false, errors.NewErr("[HeaderGetIndex] Too few input parameters ")
	}
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.NewErr("[HeaderGetIndex] Pop blockdata nil!")
	}
	var data *types.Header
	if b, ok := d.(*types.Block); ok {
		data = b.Header
	} else if h, ok := d.(*types.Header); ok {
		data = h
	} else {
		return false, errors.NewErr("[HeaderGetIndex] Wrong type!")
	}
	vm.PushData(e, data.Height)
	return true, nil
}

func (s *StateReader) HeaderGetTimestamp(e *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(e) < 1 {
		return false, errors.NewErr("[HeaderGetTimestamp] Too few input parameters ")
	}
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.NewErr("[HeaderGetTimestamp] Pop blockdata nil!")
	}
	var data *types.Header
	if b, ok := d.(*types.Block); ok {
		data = b.Header
	} else if h, ok := d.(*types.Header); ok {
		data = h
	} else {
		return false, errors.NewErr("[HeaderGetTimestamp] Wrong type!")
	}
	vm.PushData(e, data.Timestamp)
	return true, nil
}

func (s *StateReader) HeaderGetConsensusData(e *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(e) < 1 {
		return false, errors.NewErr("[HeaderGetConsensusData] Too few input parameters ")
	}
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.NewErr("[HeaderGetConsensusData] Pop blockdata nil!")
	}
	var data *types.Header
	if b, ok := d.(*types.Block); ok {
		data = b.Header
	} else if h, ok := d.(*types.Header); ok {
		data = h
	} else {
		return false, errors.NewErr("[HeaderGetConsensusData] Wrong type!")
	}
	vm.PushData(e, data.ConsensusData)
	return true, nil
}

func (s *StateReader) HeaderGetNextConsensus(e *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(e) < 1 {
		return false, errors.NewErr("[HeaderGetNextConsensus] Too few input parameters ")
	}
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.NewErr("[HeaderGetNextConsensus] Pop blockdata nil!")
	}
	var data *types.Header
	if b, ok := d.(*types.Block); ok {
		data = b.Header
	} else if h, ok := d.(*types.Header); ok {
		data = h
	} else {
		return false, errors.NewErr("[HeaderGetNextConsensus] Wrong type!")
	}
	vm.PushData(e, data.NextBookkeeper[:])
	return true, nil
}

func (s *StateReader) BlockGetTransactionCount(e *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(e) < 1 {
		return false, errors.NewErr("[BlockGetTransactionCount] Too few input parameters ")
	}
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.NewErr("[BlockGetTransactionCount] Pop blockdata nil!")
	}
	block, ok := d.(*types.Block)
	if ok == false {
		return false, errors.NewErr("[BlockGetTransactionCount] Wrong type!")
	}
	transactions := block.Transactions
	vm.PushData(e, len(transactions))
	return true, nil
}

func (s *StateReader) BlockGetTransactions(e *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(e) < 1 {
		return false, errors.NewErr("[BlockGetTransactions] Too few input parameters ")
	}
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.NewErr("[BlockGetTransactions] Pop blockdata nil!")
	}
	block, ok := d.(*types.Block)
	if ok == false {
		return false, errors.NewErr("[BlockGetTransactions] Wrong type!")
	}
	transactions := block.Transactions
	transactionList := make([]vmtypes.StackItemInterface, 0)
	for _, v := range transactions {
		transactionList = append(transactionList, vmtypes.NewInteropInterface(v))
	}
	vm.PushData(e, transactionList)
	return true, nil
}

func (s *StateReader) BlockGetTransaction(e *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(e) < 2 {
		return false, errors.NewErr("[BlockGetTransaction] Too few input parameters ")
	}
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.NewErr("[BlockGetTransaction] Pop transactions nil!")
	}
	index := vm.PopInt(e)
	if index < 0 {
		return false, errors.NewErr("[BlockGetTransaction] Pop index invalid!")
	}

	block, ok := d.(*types.Block)
	if ok == false {
		return false, errors.NewErr("[BlockGetTransactions] Wrong type!")
	}
	transactions := block.Transactions
	if index >= len(transactions) {
		return false, errors.NewErr("[BlockGetTransaction] index invalid!")
	}
	vm.PushData(e, transactions[index])
	return true, nil
}

func (s *StateReader) TransactionGetHash(e *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(e) < 1 {
		return false, errors.NewErr("[TransactionGetHash] Too few input parameters ")
	}
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.NewErr("[TransactionGetHash] Pop transaction nil!")
	}

	txn, ok := d.(*types.Transaction)
	if ok == false {
		return false, errors.NewErr("[TransactionGetHash] Wrong type!")
	}
	txHash := txn.Hash()
	vm.PushData(e, txHash.ToArray())
	return true, nil
}

func (s *StateReader) TransactionGetType(e *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(e) < 1 {
		return false, errors.NewErr("[TransactionGetType] Too few input parameters ")
	}
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.NewErr("[TransactionGetType] Pop transaction nil!")
	}
	txn, ok := d.(*types.Transaction)
	if ok == false {
		return false, errors.NewErr("[TransactionGetHash] Wrong type!")
	}
	txType := txn.TxType
	vm.PushData(e, int(txType))
	return true, nil
}

func (s *StateReader) TransactionGetAttributes(e *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(e) < 1 {
		return false, errors.NewErr("[TransactionGetAttributes] Too few input parameters ")
	}
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.NewErr("[TransactionGetAttributes] Pop transaction nil!")
	}
	txn, ok := d.(*types.Transaction)
	if ok == false {
		return false, errors.NewErr("[TransactionGetAttributes] Wrong type!")
	}
	attributes := txn.Attributes
	attributList := make([]vmtypes.StackItemInterface, 0)
	for _, v := range attributes {
		attributList = append(attributList, vmtypes.NewInteropInterface(v))
	}
	vm.PushData(e, attributList)
	return true, nil
}

func (s *StateReader) AttributeGetUsage(e *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(e) < 1 {
		return false, errors.NewErr("[AttributeGetUsage] Too few input parameters ")
	}
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.NewErr("[AttributeGetUsage] Pop txAttribute nil!")
	}
	attribute, ok := d.(*types.TxAttribute)
	if ok == false {
		return false, errors.NewErr("[AttributeGetUsage] Wrong type!")
	}
	vm.PushData(e, int(attribute.Usage))
	return true, nil
}

func (s *StateReader) AttributeGetData(e *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(e) < 1 {
		return false, errors.NewErr("[AttributeGetData] Too few input parameters ")
	}
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.NewErr("[AttributeGetData] Pop txAttribute nil!")
	}
	attribute, ok := d.(*types.TxAttribute)
	if ok == false {
		return false, errors.NewErr("[AttributeGetUsage] Wrong type!")
	}
	vm.PushData(e, attribute.Data)
	return true, nil
}

func (s *StateReader) ContractGetCode(e *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(e) < 1 {
		return false, errors.NewErr("[ContractGetCode] Too few input parameters ")
	}
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.NewErr("[ContractGetCode] Pop contractState nil!")
	}
	contractState, ok := d.(*payload.DeployCode)
	if ok == false {
		return false, errors.NewErr("[ContractGetCode] Wrong type!")
	}
	vm.PushData(e, contractState.Code)
	return true, nil
}

func (s *StateReader) StorageGetContext(e *vm.ExecutionEngine) (bool, error) {
	context, err := e.CurrentContext()
	if err != nil {
		return false, err
	}
	hash, err := context.GetCodeHash()
	if err != nil {
		return false, err
	}
	vm.PushData(e, NewStorageContext(hash))
	return true, nil
}

func (s *StateReader) StorageGet(e *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(e) < 2 {
		return false, errors.NewErr("[StorageGet] Too few input parameters ")
	}
	opInterface := vm.PopInteropInterface(e)
	if opInterface == nil {
		return false, errors.NewErr("[StorageGet] Get StorageContext error!")
	}
	context, ok := opInterface.(*StorageContext)
	if ok == false {
		return false, errors.NewErr("[StorageGet] Wrong type!")
	}
	c, err := s.ldgerStore.GetContractState(context.codeHash)
	if err != nil && !strings.EqualFold(err.Error(), ErrDBNotFound) {
		return false, err
	}
	if c == nil {
		return false, nil
	}
	key := vm.PopByteArray(e)
	item, err := s.ldgerStore.GetStorageItem(&states.StorageKey{CodeHash: context.codeHash, Key: key})
	if err != nil && !strings.EqualFold(err.Error(), ErrDBNotFound) {
		return false, err
	}
	if item == nil {
		vm.PushData(e, []byte{})
	} else {
		vm.PushData(e, item.Value)
	}
	return true, nil
}
