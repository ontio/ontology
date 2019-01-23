package message

import (
	"encoding/json"
	"fmt"

	"bytes"
	"github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology/common/serialization"
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

func NewShardContractEventMsg(shardID uint64, eventType uint32, event serialization.SerializableData, sender *actor.PID) (*CrossShardMsg, error) {
	buf := new(bytes.Buffer)
	if err := event.Serialize(buf); err != nil {
		return nil, fmt.Errorf("serialize event %d: %s", eventType, err)
	}

	msg := &ShardContractEventMsg{
		FromShard: shardID,
		EventType: eventType,
		EventData: buf.Bytes(),
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("serialize shard contract-event: %s", err)
	}

	return &CrossShardMsg{
		Version: SHARD_PROTOCOL_VERSION,
		Type:    SHARD_CONTRACT_EVENT_MSG,
		Sender:  sender,
		Data:    payload,
	}, nil
}

func NewShardBlockInfo(shardID uint64, blk *types.Block) (*ShardBlockInfo, error) {
	if blk == nil {
		return nil, fmt.Errorf("newShardBlockInfo, nil block")
	}

	blockInfo := &ShardBlockInfo{
		ShardID: shardID,
		Height:  uint64(blk.Header.Height),
		State:   ShardBlockNew,
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
		ShardID: msg.ShardID,
		Height:  uint64(msg.BlockHeader.Header.Height),
		State:   ShardBlockReceived,
		Header: &ShardBlockHeader{
			Header: msg.BlockHeader.Header,
		},
	}

	// TODO: add event from msg to blockInfo

	return blockInfo, nil
}
