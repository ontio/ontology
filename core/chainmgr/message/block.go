package message

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
)

const (
	ShardBlockNew = iota
	ShardBlockReceived
	ShardBlockProcessed
)

type ShardBlockHeader struct {
	Header *types.Header
}

type ShardBlockInfo struct {
	ShardID     uint64                       `json:"shard_id"`
	BlockHeight uint64                       `json:"block_height"`
	State       uint                         `json:"state"`
	Header      *ShardBlockHeader            `json:"header"`
	Events      []shardstates.ShardMgmtEvent `json:"events"`
}

type shardBlkHdrHelper struct {
	Payload []byte `json:"payload"`
}

func (this *ShardBlockHeader) MarshalJSON() ([]byte, error) {
	buf := common.NewZeroCopySink(nil)
	if err := this.Header.Serialization(buf); err != nil {
		return nil, fmt.Errorf("shard block hdr marshal: %s", err)
	}

	return json.Marshal(&shardBlkHdrHelper{
		Payload: buf.Bytes(),
	})
}

func (this *ShardBlockHeader) UnmarshalJSON(data []byte) error {
	helper := &shardBlkHdrHelper{}
	if err := json.Unmarshal(data, helper); err != nil {
		return fmt.Errorf("shard block hdr helper: %s", err)
	}

	buf := common.NewZeroCopySource(helper.Payload)
	hdr := &types.Header{}
	if err := hdr.Deserialization(buf); err != nil {
		return fmt.Errorf("shard block hdr unmarshal: %s", err)
	}
	this.Header = hdr
	return nil
}

func (this *ShardBlockInfo) Serialize(w io.Writer) error {
	return SerJson(w, this)
}

func (this *ShardBlockInfo) Deserialize(r io.Reader) error {
	return DesJson(r, this)
}

////////////////////////////////////
//
//  shard block pool
//
////////////////////////////////////

type ShardBlockMap map[uint64]*ShardBlockInfo // indexed by BlockHeight

type ShardBlockPool struct {
	Shards      map[uint64]ShardBlockMap // indexed by shardID
	MaxBlockCap uint32
}

func NewShardBlockPool(historyCap uint32) *ShardBlockPool {
	return &ShardBlockPool{
		Shards:      make(map[uint64]ShardBlockMap),
		MaxBlockCap: historyCap,
	}
}

func (pool *ShardBlockPool) AddBlock(blkInfo *ShardBlockInfo) error {
	if _, present := pool.Shards[blkInfo.ShardID]; !present {
		pool.Shards[blkInfo.ShardID] = make(ShardBlockMap)
	}

	m := pool.Shards[blkInfo.ShardID]
	if m == nil {
		return fmt.Errorf("add shard block, nil map")
	}
	if blk, present := m[blkInfo.BlockHeight]; present {
		if blk.State != ShardBlockNew {
			return fmt.Errorf("add shard block, new block on block state %d", blk.State)
		}
		if blk.Header != nil &&
			bytes.Compare(blk.Header.Header.BlockRoot[:], blkInfo.Header.Header.BlockRoot[:]) == 0 {
			return fmt.Errorf("add shard block, dup blk")
		}

		// replace events
		blkInfo.Events = blk.Events
		m[blkInfo.BlockHeight] = blkInfo
	}

	m[blkInfo.BlockHeight] = blkInfo

	// if too much block cached in map, drop old blocks
	if uint32(len(m)) < pool.MaxBlockCap {
		return nil
	}
	h := blkInfo.BlockHeight
	for _, blk := range m {
		if blk.BlockHeight > h {
			h = blk.BlockHeight
		}
	}

	toDrop := make([]uint64, 0)
	for _, blk := range m {
		if blk.BlockHeight < h-uint64(pool.MaxBlockCap) {
			toDrop = append(toDrop, blk.BlockHeight)
		}
	}
	for _, blkHeight := range toDrop {
		delete(m, blkHeight)
	}

	return nil
}

func (pool *ShardBlockPool) AddEvent(evt shardstates.ShardMgmtEvent) error {
	shardID := evt.GetSourceShardID()

	if _, present := pool.Shards[shardID]; !present {
		pool.Shards[shardID] = make(ShardBlockMap)
	}

	m := pool.Shards[shardID]
	if m == nil {
		return fmt.Errorf("add shard event, nil map")
	}
	if _, present := m[evt.GetHeight()]; !present {
		m[evt.GetHeight()] = &ShardBlockInfo{
			ShardID:     shardID,
			BlockHeight: evt.GetHeight(),
			State:       ShardBlockNew,
			Events:      []shardstates.ShardMgmtEvent{evt},
		}
		return nil
	}

	m[evt.GetHeight()].Events = append(m[evt.GetHeight()].Events, evt)
	return nil
}

////////////////////////////////////
//
//  json helpers
//
////////////////////////////////////

func SerJson(w io.Writer, v interface{}) error {
	buf, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("json marshal failed: %s", err)
	}

	if err := serialization.WriteVarBytes(w, buf); err != nil {
		return fmt.Errorf("json serialize write failed: %s", err)
	}
	return nil
}

func DesJson(r io.Reader, v interface{}) error {
	buf, err := serialization.ReadVarBytes(r)
	if err != nil {
		return fmt.Errorf("json deserialize read failed: %s", err)
	}
	if err := json.Unmarshal(buf, v); err != nil {
		return fmt.Errorf("json unmarshal failed: %s", err)
	}
	return nil
}
