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

package common

import (
	"math/big"
	"sync/atomic"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	types2 "github.com/ontio/ontology/validator/types"
)

const (
	MAX_CAPACITY     = 100140      // The tx pool's capacity that holds the verified txs
	MAX_PENDING_TXN  = 4096 * 10   // The max length of pending txs
	MAX_LIMITATION   = 10000       // The length of pending tx from net and http
	UPDATE_FREQUENCY = 100         // The frequency to update gas price from global params
	MAX_TX_SIZE      = 1024 * 1024 // The max size of a transaction to prevent DOS attacks
)

// SenderType enumerates the kind of tx submitter
type SenderType uint8

const (
	NilSender  SenderType = iota
	NetSender             // Net sends tx req
	HttpSender            // Http sends tx req
)

// CheckBlkResult contains a verifed tx list,
// an unverified tx list and an old tx list
// to be re-verifed
type CheckBlkResult struct {
	VerifiedTxs   []*VerifyTxResult
	UnverifiedTxs []*types.Transaction
	OldTxs        []*types.Transaction
}

// TxStatus contains the attributes of a transaction
type TxStatus struct {
	Hash  common.Uint256 // transaction hash
	Attrs []*TXAttr      // transaction's status
}

type TxResult struct {
	Err  errors.ErrCode
	Hash common.Uint256
	Desc string
}

type CheckingStatus struct {
	PassedStateless uint32 // actually bool, use uint32 for atomic operation
	PassedStateful  uint32 // actually bool, use uint32 for atomic operation
	CheckHeight     uint32
}

func (s *CheckingStatus) SetStateless() {
	atomic.StoreUint32(&s.PassedStateless, 1)
}

func (s *CheckingStatus) GetStateless() bool {
	val := atomic.LoadUint32(&s.PassedStateless)
	return val == 1
}

func (s *CheckingStatus) GetStateful() bool {
	val := atomic.LoadUint32(&s.PassedStateful)
	return val == 1
}

func (s *CheckingStatus) SetStateful(height uint32) {
	if s.CheckHeight < height {
		s.CheckHeight = height
	}
	atomic.StoreUint32(&s.PassedStateful, 1)
}

func (self *CheckingStatus) GetTxAttr() []*TXAttr {
	var res []*TXAttr
	if self.GetStateless() {
		res = append(res,
			&TXAttr{
				Height:  0,
				Type:    types2.Stateless,
				ErrCode: errors.ErrNoError,
			})
	}
	if self.GetStateful() {
		res = append(res,
			&TXAttr{
				Height:  self.CheckHeight,
				Type:    types2.Stateful,
				ErrCode: errors.ErrNoError,
			})
	}

	return res
}

// consensus messages
// GetTxnPoolReq specifies the api that how to get the valid transaction list.
type GetTxnPoolReq struct {
	ByCount bool
	Height  uint32
}

// GetTxnPoolRsp returns a transaction list for GetTxnPoolReq.
type GetTxnPoolRsp struct {
	TxnPool []*VerifiedTx
}

type TxPoolService interface {
	GetTransaction(hash common.Uint256) *types.Transaction
	GetTransactionStatus(hash common.Uint256) *TxStatus
	GetTxAmount() []uint32
	GetTxList() []common.Uint256
	AppendTransaction(sender SenderType, txn *types.Transaction) *TxResult
	AppendTransactionAsync(sender SenderType, txn *types.Transaction)
}

// VerifyBlockReq specifies that api that how to verify a block from consensus.
type VerifyBlockReq struct {
	Height uint32
	Txs    []*types.Transaction
}

// VerifyTxResult returns a single transaction's verified result.
type VerifyTxResult struct {
	Height  uint32
	Tx      *types.Transaction
	ErrCode errors.ErrCode
}

// VerifyBlockRsp returns a verified result for VerifyBlockReq.
type VerifyBlockRsp struct {
	TxnPool []*VerifyTxResult
}

type OrderByNetWorkFee []*VerifiedTx

func (n OrderByNetWorkFee) Len() int { return len(n) }

func (n OrderByNetWorkFee) Swap(i, j int) { n[i], n[j] = n[j], n[i] }

func (n OrderByNetWorkFee) Less(i, j int) bool {
	//sort the eip tx
	if n[i].Tx.TxType != n[j].Tx.TxType {
		return byte(n[i].Tx.TxType) > byte(n[j].Tx.TxType)
	}

	if n[i].Tx.TxType == types.EIP155 && n[j].Tx.TxType == types.EIP155 {
		if n[i].Tx.Payer == n[j].Tx.Payer {
			return n[i].Tx.Nonce < n[j].Tx.Nonce
		}
	}

	return n[j].Tx.GasPrice < n[i].Tx.GasPrice
}

func GetOngBalance(account common.Address) (*big.Int, error) {
	cache := ledger.DefLedger.GetStore().GetCacheDB()
	balanceKey := ont.GenBalanceKey(utils.OngContractAddress, account)
	amount, err := utils.GetStorageUInt64(cache, balanceKey)
	if err != nil {
		return nil, err
	}

	return big.NewInt(0).SetUint64(amount), nil
}
