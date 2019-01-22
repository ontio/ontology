package chainmgr

import (
	"github.com/ontio/ontology/core/chainmgr/message"
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

func (this *ChainManager) addShardEvent(evt shardstates.ShardMgmtEvent) error {
	this.lock.Lock()
	defer this.lock.Unlock()

	if err := this.blockPool.AddEvent(evt); err != nil {
		return err
	}
	return nil
}
