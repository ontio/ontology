package types

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"math"
	"math/rand"
	"testing"
)

func genShardId(t *testing.T, l1, l2, l3, l4 uint16) ShardID {
	id, err := ShardIDFromLevels(l1, l2, l3, l4)
	assert.Nil(t, err)
	return id
}

func assertWrongId(t *testing.T, l1, l2, l3, l4 uint16) {
	_, err := ShardIDFromLevels(l1, l2, l3, l4)
	assert.NotNil(t, err)
}

func TestShardID_Level(t *testing.T) {
	assert.Equal(t, genShardId(t, 0, 0, 0, 0).Level(), 0)
	assert.Equal(t, genShardId(t, 1, 0, 0, 0).Level(), 1)
	assert.Equal(t, genShardId(t, 1, 9, 0, 0).Level(), 2)
	assert.Equal(t, genShardId(t, 1, 5, 1, 0).Level(), 3)
	assert.Equal(t, genShardId(t, 1, 5, 9, 4).Level(), 4)
}

func TestNewShardID(t *testing.T) {
	assertWrongId(t, 0, 1, 0, 0)
	assertWrongId(t, 0, 1, 0, 0)
	assertWrongId(t, 1, 1, 0, 34)
	assertWrongId(t, 1, 0, 5, 0)
	assertWrongId(t, 6, 1, 0, 2)
	assertWrongId(t, 9, 0, 0, 8)
}

func TestShardID_GenSubShardID(t *testing.T) {
	_, err := genShardId(t, 1, 0, 0, 0).GenSubShardID(0)
	assert.NotNil(t, err)
	_, err = genShardId(t, 1, 0, 0, 0).GenSubShardID(math.MaxUint16)
	assert.Nil(t, err)
	id, _ := genShardId(t, 1, 0, 0, 0).GenSubShardID(234)
	assert.Equal(t, id, genShardId(t, 1, 234, 0, 0))

	id, _ = genShardId(t, 1, 5, 2, 0).GenSubShardID(234)
	assert.Equal(t, id, genShardId(t, 1, 5, 2, 234))
}

func TestShardID_ParentID(t *testing.T) {
	id := genShardId(t, 1, 5, 0, 0)
	for i := 0; i < 100; i++ {
		index := uint16(rand.Uint32())
		if index == 0 {
			index = 1
		}
		index = 1090
		subId, _ := id.GenSubShardID(index)
		assert.Equal(t, id, subId.ParentID())
	}
}

type shardIdJson struct {
	ShardID ShardID `json:"shard_id"`
}

type shardIdJsonPointer struct {
	ShardID *ShardID `json:"shard_id"`
}

func TestShardID_MarshalJSON(t *testing.T) {
	id := shardIdJson{ShardID: genShardId(t, 1, 2, 3, 4)}
	buf, _ := json.Marshal(id)
	assert.Equal(t, string(buf), fmt.Sprintf(`{"shard_id":%d}`, id.ShardID.id))
	buf, _ = json.Marshal(&id)
	assert.Equal(t, string(buf), fmt.Sprintf(`{"shard_id":%d}`, id.ShardID.id))
	id2 := shardIdJson{}
	_ = json.Unmarshal(buf, &id2)
	assert.Equal(t, id, id2)

	ptr := shardIdJsonPointer{ShardID: &id.ShardID}
	buf, _ = json.Marshal(ptr)
	assert.Equal(t, string(buf), fmt.Sprintf(`{"shard_id":%d}`, id.ShardID.id))
	buf, _ = json.Marshal(&ptr)
	assert.Equal(t, string(buf), fmt.Sprintf(`{"shard_id":%d}`, id.ShardID.id))

	ptr2 := shardIdJsonPointer{}
	_ = json.Unmarshal(buf, &ptr2)
	assert.Equal(t, ptr, ptr2)

	shardAB := ShardAB{genShardId(t, 1, 2, 3, 4), genShardId(t, 2, 3, 0, 0)}
	shardc := ShardC{shardAB}
	buf, _ = json.Marshal(shardc)
	assert.Equal(t, string(buf), fmt.Sprintf(`{"A":%d,"B":%d}`, shardAB.A.id, shardAB.B.id))
}

type ShardAB struct {
	A ShardID
	B ShardID
}

type ShardC struct {
	ShardAB
}
