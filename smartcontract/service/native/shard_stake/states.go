package shard_stake

import (
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/utils"
	"io"
)

type PeerViewInfo struct {
	NodeType            shardstates.NodeType `json:"node_type"`             // consensus or candidate
	WholeFee            uint64               `json:"whole_fee"`             // each epoch handling fee
	FeeBalance          uint64               `json:"fee_balance"`           // each epoch handling fee not be withdrawn
	WholeStakeAmount    uint64               `json:"whole_stake_amount"`    // all user stake amount
	WholeUnfreezeAmount uint64               `json:"whole_unfreeze_amount"` // all user can withdraw amount
}

type ViewInfo struct {
	Peers map[keypair.PublicKey]*PeerViewInfo `json:"peers"`
}

func (this *ViewInfo) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, this)
}

func (this *ViewInfo) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, this)
}

type UserPeerStakeInfo struct {
	StakeAmount    uint64 `json:"stake_amount"`
	UnfreezeAmount uint64 `json:"unfreeze_amount"`
}

type UserStakeInfo struct {
	Peers map[keypair.PublicKey]*UserPeerStakeInfo `json:"peers"`
}

func (this *UserStakeInfo) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, this)
}

func (this *UserStakeInfo) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, this)
}
