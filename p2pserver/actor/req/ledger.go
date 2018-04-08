package req

import (
	"time"
	"bytes"

	"github.com/Ontology/common"
	"github.com/Ontology/common/log"
	"github.com/Ontology/core/types"
	"github.com/Ontology/errors"
	"github.com/Ontology/eventbus/actor"
	ledger "github.com/Ontology/core/ledger/actor"
	msg "github.com/Ontology/p2pserver/message"
	msgCommon "github.com/Ontology/p2pserver/common"
)

var DefLedgerPid *actor.PID

func SetLedgerPid(ledgePid *actor.PID) {
	DefLedgerPid = ledgePid
}

func AddHeader(header *types.Header) {
	DefLedgerPid.Tell(&ledger.AddHeaderReq{Header: header})
}

func AddHeaders(headers []*types.Header) {
	DefLedgerPid.Tell(&ledger.AddHeadersReq{Headers: headers})
}

func AddBlock(block *types.Block) {
	DefLedgerPid.Tell(&ledger.AddBlockReq{Block: block})
}

func GetTxnFromLedger(hash common.Uint256) (*types.Transaction, error) {
	future := DefLedgerPid.RequestFuture(&ledger.GetTransactionReq{TxHash: hash}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return nil, err
	}
	return result.(*ledger.GetTransactionRsp).Tx, result.(*ledger.GetTransactionRsp).Error
}

func GetCurrentBlockHash() (common.Uint256, error) {
	future := DefLedgerPid.RequestFuture(&ledger.GetCurrentBlockHashReq{}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return common.Uint256{}, err
	}
	return result.(*ledger.GetCurrentBlockHashRsp).BlockHash, result.(*ledger.GetCurrentBlockHashRsp).Error
}

func GetCurrentHeaderHash() (common.Uint256, error) {
	future := DefLedgerPid.RequestFuture(&ledger.GetCurrentHeaderHashReq{}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return common.Uint256{}, err
	}
	return result.(*ledger.GetCurrentHeaderHashRsp).BlockHash, result.(*ledger.GetCurrentHeaderHashRsp).Error
}

func GetBlockHashByHeight(height uint32) (common.Uint256, error) {
	future := DefLedgerPid.RequestFuture(&ledger.GetBlockHashReq{Height: height}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return common.Uint256{}, err
	}
	return result.(*ledger.GetBlockHashRsp).BlockHash, result.(*ledger.GetBlockHashRsp).Error
}

func GetHeaderByHeight(height uint32) (*types.Header, error) {
	future := DefLedgerPid.RequestFuture(&ledger.GetHeaderByHeightReq{Height: height}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return nil, err
	}
	return result.(*ledger.GetHeaderByHeightRsp).Header, result.(*ledger.GetHeaderByHeightRsp).Error
}

func GetBlockByHeight(height uint32) (*types.Block, error) {
	future := DefLedgerPid.RequestFuture(&ledger.GetBlockByHeightReq{Height: height}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return nil, err
	}
	return result.(*ledger.GetBlockByHeightRsp).Block, result.(*ledger.GetBlockByHeightRsp).Error
}

func GetHeaderByHash(hash common.Uint256) (*types.Header, error) {
	future := DefLedgerPid.RequestFuture(&ledger.GetHeaderByHashReq{BlockHash: hash}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return nil, err
	}
	return result.(*ledger.GetHeaderByHashRsp).Header, result.(*ledger.GetHeaderByHashRsp).Error
}

func GetBlockByHash(hash common.Uint256) (*types.Block, error) {
	future := DefLedgerPid.RequestFuture(&ledger.GetBlockByHashReq{BlockHash: hash}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return nil, err
	}
	return result.(*ledger.GetBlockByHashRsp).Block, result.(*ledger.GetBlockByHashRsp).Error
}

func GetCurrentHeaderHeight() (uint32, error) {
	future := DefLedgerPid.RequestFuture(&ledger.GetCurrentHeaderHeightReq{}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return 0, err
	}
	return result.(*ledger.GetCurrentHeaderHeightRsp).Height, result.(*ledger.GetCurrentHeaderHeightRsp).Error
}

func GetCurrentBlockHeight() (uint32, error) {
	future := DefLedgerPid.RequestFuture(&ledger.GetCurrentBlockHeightReq{}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return 0, err
	}
	return result.(*ledger.GetCurrentBlockHeightRsp).Height, result.(*ledger.GetCurrentBlockHeightRsp).Error
}

func IsContainBlock(hash common.Uint256) (bool, error) {
	future := DefLedgerPid.RequestFuture(&ledger.IsContainBlockReq{BlockHash: hash}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return false, err
	}
	return result.(*ledger.IsContainBlockRsp).IsContain, result.(*ledger.IsContainBlockRsp).Error
}

func GetHeadersFromHash(startHash common.Uint256, stopHash common.Uint256) ([]types.Header, uint32, error) {
	var count uint32 = 0
	var empty [msgCommon.HASH_LEN]byte
	headers := []types.Header{}
	var startHeight uint32
	var stopHeight uint32
	curHeight, _ := GetCurrentHeaderHeight()
	if startHash == empty {
		if stopHash == empty {
			if curHeight > msgCommon.MAX_BLK_HDR_CNT {
				count = msgCommon.MAX_BLK_HDR_CNT
			} else {
				count = curHeight
			}
		} else {
			bkStop, err := GetHeaderByHash(stopHash)
			if err != nil {
				return nil, 0, err
			}
			stopHeight = bkStop.Height
			count = curHeight - stopHeight
			if count > msgCommon.MAX_BLK_HDR_CNT {
				count = msgCommon.MAX_BLK_HDR_CNT
			}
		}
	} else {
		bkStart, err := GetHeaderByHash(startHash)
		if err != nil {
			return nil, 0, err
		}
		startHeight = bkStart.Height
		if stopHash != empty {
			bkStop, err := GetHeaderByHash(stopHash)
			if err != nil {
				return nil, 0, err
			}
			stopHeight = bkStop.Height

			// avoid unsigned integer underflow
			if startHeight < stopHeight {
				return nil, 0, errors.NewErr("do not have header to send")
			}
			count = startHeight - stopHeight

			if count >= msgCommon.MAX_BLK_HDR_CNT {
				count = msgCommon.MAX_BLK_HDR_CNT
				stopHeight = startHeight - msgCommon.MAX_BLK_HDR_CNT
			}
		} else {

			if startHeight > msgCommon.MAX_BLK_HDR_CNT {
				count = msgCommon.MAX_BLK_HDR_CNT
			} else {
				count = startHeight
			}
		}
	}

	var i uint32
	for i = 1; i <= count; i++ {
		hash, err := GetBlockHashByHeight(stopHeight + i)
		hd, err := GetHeaderByHash(hash)
		if err != nil {
			log.Errorf("GetBlockWithHeight failed with err=%s, hash=%x,height=%d\n", err.Error(), hash, stopHeight+i)
			return nil, 0, err
		}
		headers = append(headers, *hd)
	}

	return headers, count, nil
}

func GetInvFromBlockHash(startHash common.Uint256, stopHash common.Uint256) (*msg.InvPayload, error) {
	var count uint32 = 0
	var i uint32
	var empty common.Uint256
	var startHeight uint32
	var stopHeight uint32
	curHeight, _ := GetCurrentBlockHeight()
	if startHash == empty {
		if stopHash == empty {
			if curHeight > msgCommon.MAX_BLK_HDR_CNT {
				count = msgCommon.MAX_BLK_HDR_CNT
			} else {
				count = curHeight
			}
		} else {
			bkStop, err := GetHeaderByHash(stopHash)
			if err != nil {
				return nil, err
			}
			stopHeight = bkStop.Height
			count = curHeight - stopHeight
			if curHeight > msgCommon.MAX_INV_HDR_CNT {
				count = msgCommon.MAX_INV_HDR_CNT
			}
		}
	} else {
		bkStart, err := GetHeaderByHash(startHash)
		if err != nil {
			return nil, err
		}
		startHeight = bkStart.Height
		if stopHash != empty {
			bkStop, err := GetHeaderByHash(stopHash)
			if err != nil {
				return nil, err
			}
			stopHeight = bkStop.Height
			count = startHeight - stopHeight
			if count >= msgCommon.MAX_INV_HDR_CNT {
				count = msgCommon.MAX_INV_HDR_CNT
				stopHeight = startHeight + msgCommon.MAX_INV_HDR_CNT
			}
		} else {

			if startHeight > msgCommon.MAX_INV_HDR_CNT {
				count = msgCommon.MAX_INV_HDR_CNT
			} else {
				count = startHeight
			}
		}
	}
	tmpBuffer := bytes.NewBuffer([]byte{})
	for i = 1; i <= count; i++ {
		//FIXME need add error handle for GetBlockWithHash
		hash, _ := GetBlockHashByHeight(stopHeight + i)
		log.Debug("GetInvFromBlockHash i is ", i, " , hash is ", hash)
		hash.Serialize(tmpBuffer)
	}
	log.Debug("GetInvFromBlockHash hash is ", tmpBuffer.Bytes())

	return &msg.InvPayload{
		InvType: common.BLOCK,
		Cnt:     count,
		Blk:     tmpBuffer.Bytes(),
	}, nil
}

