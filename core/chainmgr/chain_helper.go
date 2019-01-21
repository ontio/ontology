package chainmgr

import (
	"github.com/ontio/ontology/core/chainmgr/message"
)

func (this *ChainManager) addShardBlockInfo(blkInfo *message.ShardBlockInfo) error {
	this.lock.Lock()
	defer this.lock.Unlock()

	if err := this.blockPool.AddBlock(blkInfo); err != nil {
		return err
	}

	return nil
}
