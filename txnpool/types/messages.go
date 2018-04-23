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
package types

import (
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/errors"
)



// TxReq specifies the api that how to submit a new transaction.
// Input: transacton and submitter type
type AppendTxReq struct {
	Tx     *types.Transaction
	Sender SenderType
}

// TxRsp returns the result of submitting tx, including
// a transaction hash and error code.
type AppendTxRsp struct {
	Hash    common.Uint256
	ErrCode errors.ErrCode
}

// restful api

// GetTxnReq specifies the api that how to get the transaction.
// Input: a transaction hash
type GetTxFromPoolReq struct {
	Hash common.Uint256
}

// GetTxnRsp returns a transaction for the specified tx hash.
type GetTxFromPoolRsp struct {
	Txn *types.Transaction
}

// CheckTxnReq specifies the api that how to check whether a
// transaction in the pool.
// Input: a transaction hash
type IsTxInPoolReq struct {
	Hash common.Uint256
}

// CheckTxnRsp returns a value for the CheckTxnReq, if the
// transaction in the pool, value is true, or false.
type IsTxInPoolRsp struct {
	Ok bool
}

// GetTxnStatusReq specifies the api that how to get a transaction
// status.
// Input: a transaction hash.
type GetTxVerifyResultReq struct {
	Hash common.Uint256
}

// GetTxnStatusRsp returns a transaction status for GetTxnStatusReq.
// Output: a transaction hash and it's verified result.
type GetTxVerifyResultRsp struct {
	Hash          common.Uint256
	VerifyResults []*VerifyResult
}

// GetTxnStats specifies the api that how to get the tx statistics.
type GetTxVerifyResultStaticsReq struct {
}

// GetTxnStatsRso returns the tx statistics.
type GetTxVerifyResultStaticsRsp struct {
	Count []uint64
}

// GetPendingTxnReq specifies the api that how to get a pending tx list
// in the pool.
type GetPendingTxnReq struct {
	ByCount bool
}

// GetPendingTxnRsp returns a transaction list for GetPendingTxnReq.
type GetPendingTxnRsp struct {
	Txs []*types.Transaction
}

// consensus messages
// GetTxnPoolReq specifies the api that how to get the valid transaction list.
type GetTxnPoolReq struct {
	ByCount bool
	Height  uint32
}

// GetTxnPoolRsp returns a transaction list for GetTxnPoolReq.
type GetTxnPoolRsp struct {
	TxnPool []*TxEntry
}

// VerifyBlockReq specifies that api that how to verify a block from consensus.
type VerifyBlockReq struct {
	Height uint32
	Txs    []*types.Transaction
}

// VerifyBlockRsp returns a verified result for VerifyBlockReq.
type VerifyBlockRsp struct {
	TxResults []*TxResult
}
