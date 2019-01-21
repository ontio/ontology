package chainmgr

import (
	"github.com/ontio/ontology/core/chainmgr/message"
	"fmt"
)

func (this *ChainManager) addShardBlockInfo(blkInfo *message.ShardBlockInfo) error {
	this.Lock.Lock()
	defer this.Lock.Unlock()

	if _, present := this.ShardBlocks[blkInfo.ShardID]; !present {
		this.ShardBlocks[blkInfo.ShardID] = make(message.ShardBlockMap)
	}

	m := this.ShardBlocks[blkInfo.ShardID]
	if m == nil {
		return fmt.Errorf("add shard block, nil map")
	}
	if _, present := m[blkInfo.BlockHeight]; present {
		return fmt.Errorf("add shard block, dup blk")
	}

	m[blkInfo.BlockHeight] = blkInfo

	// TODO: if too much block cached in map, drop old blocks

	return nil
}


