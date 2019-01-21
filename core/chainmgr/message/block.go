package message

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
	"io"
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
	ShardID     uint64                         `json:"shard_id"`
	BlockHeight uint64                         `json:"block_height"`
	State       uint                           `json:"state"`
	Header      *ShardBlockHeader              `json:"header"`
	Events      []*shardstates.ShardEventState `json:"events"`
}

type ShardBlockMap map[uint64]*ShardBlockInfo // indexed by BlockHeight

type shardBlkHdrHelper struct {
	Payload []byte `json:"payload"`
}
func (this *ShardBlockHeader) MarshalJSON() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := this.Header.Serialize(buf); err != nil {
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

	buf := bytes.NewBuffer(helper.Payload)
	hdr := &types.Header{}
	if err := hdr.Deserialize(buf); err != nil {
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

//////
//
//  json helpers
//
//////

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
