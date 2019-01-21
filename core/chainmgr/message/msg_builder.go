package message

import (
	"encoding/json"
	"fmt"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology/core/types"
)

func NewShardHelloMsg(localShard, targetShard uint64, sender *actor.PID) (*CrossShardMsg, error) {
	hello := &ShardHelloMsg{
		TargetShardID: targetShard,
		SourceShardID: localShard,
	}
	payload, err := json.Marshal(hello)
	if err != nil {
		return nil, fmt.Errorf("marshal hello msg: %s", err)
	}

	return &CrossShardMsg{
		Version: SHARD_PROTOCOL_VERSION,
		Type:    HELLO_MSG,
		Sender:  sender,
		Data:    payload,
	}, nil
}

func NewShardConfigMsg(accPayload []byte, configPayload []byte, sender *actor.PID) (*CrossShardMsg, error) {
	ack := &ShardConfigMsg{
		Account: accPayload,
		Config:  configPayload,
	}
	payload, err := json.Marshal(ack)
	if err != nil {
		return nil, fmt.Errorf("marshal hello ack msg: %s", err)
	}

	return &CrossShardMsg{
		Version: SHARD_PROTOCOL_VERSION,
		Type:    CONFIG_MSG,
		Sender:  sender,
		Data:    payload,
	}, nil
}

func NewShardBlockRspMsg(shardID uint64, blockNum uint64, blockHdr *types.Header, sender *actor.PID) (*CrossShardMsg, error) {
	blkRsp := &ShardBlockRspMsg{
		ShardID: shardID,
		Height:  blockNum,
		BlockHeader: &ShardBlockHeader{
			Header: blockHdr,
		},
	}

	// TODO: add events to blockRspMsg

	payload, err := json.Marshal(blkRsp)
	if err != nil {
		return nil, fmt.Errorf("marshal shard block rsp msg: %s", err)
	}

	return &CrossShardMsg{
		Version: SHARD_PROTOCOL_VERSION,
		Type:    BLOCK_RSP_MSG,
		Sender:  sender,
		Data:    payload,
	}, nil
}

func NewShardBlockInfo(shardID uint64, blk *types.Block) (*ShardBlockInfo, error) {
	if blk == nil {
		return nil, fmt.Errorf("newShardBlockInfo, nil block")
	}

	blockInfo := &ShardBlockInfo{
		ShardID:     shardID,
		BlockHeight: uint64(blk.Header.Height),
		State:       ShardBlockNew,
		Header: &ShardBlockHeader{
			Header: blk.Header,
		},
	}

	// TODO: add event from block to blockInfo

	return blockInfo, nil
}

func NewShardBlockInfoFromRemote(msg *ShardBlockRspMsg) (*ShardBlockInfo, error) {
	if msg == nil {
		return nil, fmt.Errorf("newShardBlockInfo, nil msg")
	}

	blockInfo := &ShardBlockInfo{
		ShardID:     msg.ShardID,
		BlockHeight: uint64(msg.BlockHeader.Header.Height),
		State:       ShardBlockReceived,
		Header: &ShardBlockHeader{
			Header: msg.BlockHeader.Header,
		},
	}

	// TODO: add event from msg to blockInfo

	return blockInfo, nil
}
