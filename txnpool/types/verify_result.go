package types

import (
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/common"
	vtypes "github.com/ontio/ontology/validator/types"
)

const (
	MAX_CAPACITY    = 100140                           // The tx pool's capacity that holds the verified txs
	MAX_PENDING_TXN = 4096 * 10                        // The max length of pending txs
	MAX_WORKER_NUM  = 2                                // The max concurrent workers
	MAX_RCV_TXN_LEN = MAX_WORKER_NUM * MAX_PENDING_TXN // The max length of the queue that server can hold
	MAX_RETRIES     = 0                                // The retry times to verify tx
	EXPIRE_INTERVAL = 9                                // The timeout that verify tx
	STATELESS_MASK  = 0x1                              // The mask of stateless validator
	STATEFULL_MASK  = 0x2                              // The mask of stateful validator
	VERIFY_MASK     = STATELESS_MASK | STATEFULL_MASK  // The mask that indicates tx valid
	MAX_LIMITATION  = 10000                            // The length of pending tx from net and http
)

// ActorType enumerates the kind of actor
type ActorType uint8

const (
	_              ActorType = iota
	TxStatusActor   // Actor that handles new transaction
	TxPoolActor     // Actor that handles consensus msg
	VerifyRspActor  // Actor that handles the response from valdiators
	NetActor        // Actor to send msg to the net actor
	MaxActor
)

// SenderType enumerates the kind of tx submitter
type SenderType uint8

const (
	NilSender  SenderType = iota
	NetSender   // Net sends tx req
	HttpSender  // Http sends tx req
)

func (sender SenderType) Sender() string {
	switch sender {
	case NilSender:
		return "nil sender"
	case NetSender:
		return "net sender"
	case HttpSender:
		return "http sender"
	default:
		return "unknown sender"
	}
}

// TxnStatsType enumerates the kind of tx statistics
type VerifyResultType uint8

const (
	_         VerifyResultType = iota
	Received   // The count that the tx pool receive from the actor bus
	Success    // The count that the transactions are verified successfully
	Failure    // The count that the transactions are invalid
	Duplicate  // The count that the transactions are duplicated input
	SigErr     // The count that the transactions' signature error
	StateErr   // The count that the transactions are invalid in database

	MaxStats
)

type VerifyResult struct {
	Height  uint32            // The height in which tx was verified
	Type    vtypes.VerifyType // The validator flag: stateless/stateful
	ErrCode errors.ErrCode    // Verified result
}

type TxEntry struct {
	Tx            *types.Transaction // transaction which has been verified
	Fee           common.Fixed64     // Total fee per transaction
	VerifyResults []*VerifyResult    // the result from each validator
}

// TxStatus contains the attributes of a transaction
type TxVerifyStatus struct {
	Hash          common.Uint256  // transaction hash
	VerifyResults []*VerifyResult // transaction's status
}

// VerifyTxResult returns a single transaction's verified result.
type TxResult struct {
	Height  uint32
	Tx      *types.Transaction
	ErrCode errors.ErrCode
}

// CheckBlkResult contains a verifed tx list,
// an unverified tx list and an old tx list
// to be re-verifed
type BlockVerifyStatus struct {
	VerifiedTxs   []*TxResult
	UnVerifiedTxs []*types.Transaction
	ReVerifyTxs   []*types.Transaction
}
