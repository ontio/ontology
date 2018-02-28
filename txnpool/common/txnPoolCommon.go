package common

import (
	"github.com/Ontology/common"
	"github.com/Ontology/core/types"
)

const (
	MAXPENDINGTXN  = 2048
	MAXWORKERNUM   = 2
	MAXRCVTXNLEN   = MAXWORKERNUM * MAXPENDINGTXN
	MAXRETRIES     = 3
	EXPIREINTERVAL = 2
	SIGNATUREMASK  = 0x1
	STATEFULMASK   = 0x2
	VERIFYMASK     = SIGNATUREMASK | STATEFULMASK
	TOPIC          = "TXN"
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
	TxnHash     common.Uint256
	Ok          bool
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
