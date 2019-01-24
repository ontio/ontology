package chainmgr

import (
	"github.com/ontio/ontology/core/chainmgr/message"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
)

func (this *ChainManager) addShardBlockInfo(blkInfo *message.ShardBlockInfo) error {
	this.lock.Lock()
	defer this.lock.Unlock()

	if err := this.blockPool.AddBlock(blkInfo); err != nil {
		return err
	}

	return nil
}

func (this *ChainManager) getShardBlockInfo(shardID uint64, height uint64) *message.ShardBlockInfo {
	this.lock.RLock()
	defer this.lock.RUnlock()

	return this.blockPool.GetBlock(shardID, height)
}

func (this *ChainManager) addShardEvent(evt *shardstates.ShardEventState) error {
	this.lock.Lock()
	defer this.lock.Unlock()

	if err := this.blockPool.AddEvent(this.shardID, evt); err != nil {
		return err
	}
	return nil
}

func (this *ChainManager) updateShardBlockInfo(shardID uint64, height uint64, blk *types.Block, shardTxs map[uint64]*message.ShardBlockTx) {
	this.lock.Lock()
	defer this.lock.Unlock()

	blkInfo := this.blockPool.GetBlock(shardID, height)
	if blkInfo == nil {
		return
	}

	blkInfo.Header = &message.ShardBlockHeader{Header: blk.Header}
	blkInfo.ShardTxs = shardTxs
}
