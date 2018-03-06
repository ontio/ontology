package common

import (
	"github.com/Ontology/common"
	"github.com/Ontology/core/types"
)

const (
	MAXPENDINGTXN  = 2048                         // The max length of pending txs
	MAXWORKERNUM   = 2                            // The max concurrent workers
	MAXRCVTXNLEN   = MAXWORKERNUM * MAXPENDINGTXN // The max length of the queue that server can hold
	MAXRETRIES     = 3                            // The retry times to verify tx
	EXPIREINTERVAL = 2                            // The timeout that verify tx
	STATELESSMASK  = 0x1                          // The mask of stateless validator
	STATEFULMASK   = 0x2                          // The mask of stateful validator
	VERIFYMASK     = STATELESSMASK | STATEFULMASK
)

type ActorType uint8

const (
	_ ActorType = iota
	TxActor
	TxPoolActor
	VerifyRspActor

	MAXACTOR
)

type VerifyType uint8

const (
	_ VerifyType = iota
	SignatureV
	StatefulV

	MAXVALIDATOR
)

type TxnStatsType uint8

const (
	_ TxnStatsType = iota
	RcvStats
	SuccessStats
	FailureStats
	DuplicateStats
	SigErrStats
	StateErrStats

	MAXSTATS
)

type VerifyReq struct {
	WorkerId uint8
	Txn      *types.Transaction
}

type VerifyRsp struct {
	WorkerId    uint8
	ValidatorID uint8
	Height      uint32
	TxHash      common.Uint256
	Ok          bool
}

type TxStatus struct {
	Hash  common.Uint256 // transaction hash
	Attrs []*TXAttr      // transaction's status
}

type GetTxnPoolReq struct {
	ByCount bool
}

type GetTxnPoolRsp struct {
	TxnPool []*TXEntry
}

type GetPendingTxnReq struct {
	ByCount bool
}

type GetPendingTxnRsp struct {
	Txs []*types.Transaction
}

type GetUnverifiedTxsReq struct {
	Txs []*types.Transaction
}

type GetUnverifiedTxsRsp struct {
	Txs []*types.Transaction
}

type CleanTxnPoolReq struct {
	TxnPool []*types.Transaction
}

type GetTxnReq struct {
	Hash common.Uint256
}

type GetTxnRsp struct {
	Txn *types.Transaction
}

type CheckTxnReq struct {
	Hash common.Uint256
}

type CheckTxnRsp struct {
	Ok bool
}

type GetTxnStatusReq struct {
	Hash common.Uint256
}

type GetTxnStatusRsp struct {
	Hash     common.Uint256
	TxStatus []*TXAttr
}

type GetTxnStats struct {
}

type GetTxnStatsRsp struct {
	Count *[]uint64
}

/*
 * Implement sort.Interface
 */
type LB struct {
	Size     int
	WorkerID uint8
}

type LBSlice []LB

func (this LBSlice) Len() int {
	return len(this)
}

func (this LBSlice) Swap(i, j int) {
	this[i].Size, this[j].Size = this[j].Size, this[i].Size
	this[i].WorkerID, this[j].WorkerID = this[j].WorkerID, this[i].WorkerID
}

func (this LBSlice) Less(i, j int) bool {
	return this[i].Size < this[j].Size
}
