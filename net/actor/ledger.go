package actor

import (
	"github.com/Ontology/common"
	"github.com/Ontology/common/log"
	"github.com/Ontology/core/types"
	"github.com/Ontology/errors"
	"github.com/Ontology/eventbus/actor"
	"time"
)

var defLedgerPid *actor.PID

type AddHeaderReq struct {
	Header *[]types.Header
}
type AddHeaderRsp struct {
	BlockHash *common.Uint256
	Error     error
}

type AddBlockReq struct {
	Block *types.Block
}
type AddBlockRsp struct {
	BlockHash *common.Uint256
	Error     error
}

type GetTransactionReq struct {
	TxHash common.Uint256
}
type GetTransactionRsp struct {
	Tx    *types.Transaction
	Error error
}

type GetBlockByHashReq struct {
	BlockHash common.Uint256
}
type GetBlockByHashRsp struct {
	Block *types.Block
	Error error
}

type GetBlockByHeightReq struct {
	Height uint32
}
type GetBlockByHeightRsp struct {
	Block *types.Block
	Error error
}

type GetHeaderByHashReq struct {
	BlockHash common.Uint256
}
type GetHeaderByHashRsp struct {
	Header *types.Header
	Error  error
}

type GetHeaderByHeightReq struct {
	Height uint32
}
type GetHeaderByHeightRsp struct {
	Header *types.Header
	Error  error
}

type GetCurrentBlockHashReq struct{}
type GetCurrentBlockHashRsp struct {
	BlockHash common.Uint256
	Error     error
}

type GetCurrentBlockHeightReq struct{}
type GetCurrentBlockHeightRsp struct {
	Height uint32
	Error  error
}

type GetCurrentHeaderHeightReq struct{}
type GetCurrentHeaderHeightRsp struct {
	Height uint32
	Error  error
}

type GetBlockHashReq struct {
	Height uint32
}
type GetBlockHashRsp struct {
	BlockHash common.Uint256
	Error     error
}

type IsContainBlockReq struct {
	BlockHash common.Uint256
}
type IsContainBlockRsp struct {
	IsContain bool
	Error     error
}

func SetLedgerPid(ledgePid *actor.PID){
	defLedgerPid = ledgePid
}

func AddHeader(header *[]types.Header) {
	defLedgerPid.Tell(&AddHeaderReq{Header: header})
}

func AddBlock(block *types.Block) {
	defLedgerPid.Tell(&AddBlockReq{Block: block})
}

func GetTxnFromLedger(hash common.Uint256) (*types.Transaction, error) {
	future := defLedgerPid.RequestFuture(&GetTransactionReq{hash}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return nil, err
	}
	return result.(GetTransactionRsp).Tx, result.(GetTransactionRsp).Error
}

func GetCurrentBlockHash() (common.Uint256, error) {
	future := defLedgerPid.RequestFuture(&GetCurrentBlockHashReq{}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return common.Uint256{}, err
	}
	return result.(GetCurrentBlockHashRsp).BlockHash, result.(GetCurrentBlockHashRsp).Error
}

func GetBlockHashByHeight(height uint32) (common.Uint256, error) {
	future := defLedgerPid.RequestFuture(&GetBlockHashReq{height}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return common.Uint256{}, err
	}
	return result.(GetBlockHashRsp).BlockHash, result.(GetBlockHashRsp).Error
}

func GetHeaderByHeight(height uint32) (*types.Header, error) {
	future := defLedgerPid.RequestFuture(&GetHeaderByHeightReq{height}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return nil, err
	}
	return result.(GetHeaderByHeightRsp).Header, result.(GetHeaderByHeightRsp).Error
}

func GetBlockByHeight(height uint32) (*types.Block, error) {
	future := defLedgerPid.RequestFuture(&GetBlockByHeightReq{height}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return nil, err
	}
	return result.(GetBlockByHeightRsp).Block, result.(GetBlockByHeightRsp).Error
}

func GetHeaderByHash(hash common.Uint256) (*types.Header, error) {
	future := defLedgerPid.RequestFuture(&GetHeaderByHashReq{hash}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return nil, err
	}
	return result.(GetHeaderByHashRsp).Header, result.(GetHeaderByHashRsp).Error
}

func GetBlockByHash(hash common.Uint256) (*types.Block, error) {
	future := defLedgerPid.RequestFuture(&GetBlockByHashReq{hash}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return nil, err
	}
	return result.(GetBlockByHashRsp).Block, result.(GetBlockByHashRsp).Error
}

func GetCurrentHeaderHeight() (uint32, error) {
	future := defLedgerPid.RequestFuture(&GetCurrentHeaderHeightReq{}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return 0, err
	}
	return result.(GetCurrentHeaderHeightRsp).Height, result.(GetCurrentHeaderHeightRsp).Error
}

func GetCurrentBlockHeight() (uint32, error) {
	future := defLedgerPid.RequestFuture(&GetCurrentBlockHeightReq{}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return 0, err
	}
	return result.(GetCurrentBlockHeightRsp).Height, result.(GetCurrentBlockHeightRsp).Error
}

func IsContainBlock(hash common.Uint256) (bool, error) {
	future := defLedgerPid.RequestFuture(&IsContainBlockReq{hash}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return false, err
	}
	return result.(IsContainBlockRsp).IsContain, result.(IsContainBlockRsp).Error
}
