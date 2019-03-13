package shard_stake

import (
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/utils"
	"io"
)

type UnfreezeFromShardParam struct {
	ShardId    types.ShardID  `json:"shard_id"`
	Address    common.Address `json:"address"`
	PeerPubKey []string       `json:"peer_pub_key"`
	Amount     []uint64       `json:"amount"`
}

func (this *UnfreezeFromShardParam) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, this)
}

func (this *UnfreezeFromShardParam) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, this)
}

type WithdrawStakeAssetParam struct {
	ShardId types.ShardID  `json:"shard_id"`
	User    common.Address `json:"user"`
}

func (this *WithdrawStakeAssetParam) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, this)
}

func (this *WithdrawStakeAssetParam) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, this)
}

type WithdrawFeeParam struct {
	ShardId types.ShardID  `json:"shard_id"`
	User    common.Address `json:"user"`
}

func (this *WithdrawFeeParam) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, this)
}

func (this *WithdrawFeeParam) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, this)
}

type CommitDposParam struct {
	ShardId    types.ShardID `json:"shard_id"`
	Amount     []uint64      `json:"amount"`
	PeerPubKey []string      `json:"peer_pub_key"`
}

func (this *CommitDposParam) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, this)
}

func (this *CommitDposParam) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, this)
}
