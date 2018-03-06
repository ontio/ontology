package actor

import (
	"github.com/ONTID/eventbus/actor"
	"github.com/Ontology/ledger"
	"github.com/Ontology/transaction"
	"github.com/Ontology/common"
	"github.com/Ontology/common/log"
	"github.com/Ontology/errors"
	"time"
)

var DefLedgerPid actor.PID

type AddHeaderReq struct{
	Header *[]ledger.Header
}
type AddHeaderRsp struct{
	BlockHash *common.Uint256
	Error error
}

type AddBlockReq struct{
	Block *ledger.Block
}
type AddBlockRsp struct{
	BlockHash *common.Uint256
	Error error
}

type GetTransactionReq struct{
	TxHash common.Uint256
}
type GetTransactionRsp struct{
	Tx *transaction.Transaction
	Error error
}

type GetBlockByHashReq struct{
	BlockHash common.Uint256
}
type GetBlockByHashRsp struct{
	Block *ledger.Block
	Error error
}

type GetBlockByHeightReq struct{
	Height uint32
}
type GetBlockByHeightRsp struct{
	Block *ledger.Block
	Error error
}

type GetHeaderByHashReq struct{
	BlockHash common.Uint256
}
type GetHeaderByHashRsp struct{
	Header *ledger.Header
	Error error
}

type GetHeaderByHeightReq struct{
	Height uint32
}
type GetHeaderByHeightRsp struct{
	Header *ledger.Header
	Error error
}

type GetCurrentBlockHashReq struct{}
type GetCurrentBlockHashRsp struct{
	BlockHash common.Uint256
	Error error
}

type GetCurrentBlockHeightReq struct{}
type GetCurrentBlockHeightRsp struct{
	Height uint32
	Error error
}

type GetCurrentHeaderHeightReq struct{}
type GetCurrentHeaderHeightRsp struct{
	Height uint32
	Error error
}

type GetBlockHashReq struct{
	Height uint32
}
type GetBlockHashRsp struct{
	BlockHash common.Uint256
	Error error
}

type IsContainBlockReq struct{
	BlockHash common.Uint256
}
type IsContainBlockRsp struct{
	IsContain bool
	Error error
}

//------------------------------------------------------------------------------------
func AddHeader(header *[]ledger.Header){
	DefLedgerPid.Tell(&AddHeaderReq{Header: header})
}

func AddBlock(block *ledger.Block){
	DefLedgerPid.Tell(&AddBlockReq{Block: block})
}

func GetTxnFromLedger(hash common.Uint256)(*transaction.Transaction, error){
	future := DefLedgerPid.RequestFuture(&GetTransactionReq{hash}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
	}
	return result.(GetTransactionRsp).Tx, result.(GetTransactionRsp).Error
}

func GetCurrentBlockHash() (common.Uint256, error) {
	future := DefLedgerPid.RequestFuture(&GetCurrentBlockHashReq{}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
	}
	return result.(GetCurrentBlockHashRsp).BlockHash, result.(GetCurrentBlockHashRsp).Error
}

func GetBlockHashByHeight(height uint32) (common.Uint256, error) {
	future := DefLedgerPid.RequestFuture(&GetBlockHashReq{height}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
	}
	return result.(GetBlockHashRsp).BlockHash, result.(GetBlockHashRsp).Error
}

func GetHeaderByHeight(height uint32) (*ledger.Header, error){
	future := DefLedgerPid.RequestFuture(&GetHeaderByHeightReq{height}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
	}
	return result.(GetHeaderByHeightRsp).Header, result.(GetHeaderByHeightRsp).Error
}

func GetBlockByHeight(height uint32)(*ledger.Block, error){
	future := DefLedgerPid.RequestFuture(&GetBlockByHeightReq{height}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
	}
	return result.(GetBlockByHeightRsp).Block, result.(GetBlockByHeightRsp).Error
}

func GetHeaderByHash(hash common.Uint256)(*ledger.Header, error){
	future := DefLedgerPid.RequestFuture(&GetHeaderByHashReq{hash}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
	}
	return result.(GetHeaderByHashRsp).Header, result.(GetHeaderByHashRsp).Error
}

func GetBlockByHash(hash common.Uint256)(*ledger.Block, error){
	future := DefLedgerPid.RequestFuture(&GetBlockByHashReq{hash}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
	}
	return result.(GetBlockByHashRsp).Block, result.(GetBlockByHashRsp).Error
}

func GetCurrentHeaderHeight() (uint32, error) {
	future := DefLedgerPid.RequestFuture(&GetCurrentHeaderHeightReq{}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
	}
	return result.(GetCurrentHeaderHeightRsp).Height, result.(GetCurrentHeaderHeightRsp).Error
}

func GetCurrentBlockHeight() (uint32, error) {
	future := DefLedgerPid.RequestFuture(&GetCurrentBlockHeightReq{}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
	}
	return result.(GetCurrentBlockHeightRsp).Height, result.(GetCurrentBlockHeightRsp).Error
}

func IsContainBlock(hash common.Uint256) (bool, error) {
	future := DefLedgerPid.RequestFuture(&IsContainBlockReq{hash}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
	}
	return result.(IsContainBlockRsp).IsContain, result.(IsContainBlockRsp).Error
}
