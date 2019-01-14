package shardmgmt

import (
	"github.com/ontio/ontology/common"
	"io"
	"github.com/ontio/ontology/common/serialization"
	"fmt"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/utils"
)

type CommonParam struct {
	Input []byte
}

func (this *CommonParam) Serialize(w io.Writer) error {
	if err := serialization.WriteVarBytes(w, this.Input); err != nil {
		return fmt.Errorf("CommonParam serialize write failed: %s", err)
	}
	return nil
}

func (this *CommonParam) Deserialize(r io.Reader) error {
	buf, err := serialization.ReadVarBytes(r)
	if err != nil {
		return fmt.Errorf("CommonParam deserialize read failed: %s", err)
	}
	this.Input = buf
	return nil
}

type CreateShardParam struct {
	ParentShardID uint64         `json:"parent_shard_id"`
	Creator       common.Address `json:"creator"`
}

func (this *CreateShardParam) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, this)
}

func (this *CreateShardParam) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, this)
}

type ConfigShardParam struct {
	ShardID        uint64 `json:"shard_id"`
	ConfigTestData []byte `json:"config_test_data"`
}

func (this *ConfigShardParam) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, this)
}

func (this *ConfigShardParam) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, this)
}

type JoinShardParam struct {
	ShardID     uint64         `json:"shard_id"`
	PeerOwner   common.Address `json:"peer_owner"`
	PeerPubKey  string         `json:"peer_pub_key"`
	StakeAmount uint64         `json:"stake_amount"`
}

func (this *JoinShardParam) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, this)
}

func (this *JoinShardParam) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, this)
}
