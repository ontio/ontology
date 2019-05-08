package shard_stake

import (
	"bytes"
	"encoding/hex"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInitShardParam(t *testing.T) {
	acc := account.NewAccount("")
	param := &InitShardParam{
		ShardId:        common.NewShardIDUnchecked(9),
		StakeAssetAddr: acc.Address,
		MinStake:       73256817645609,
	}
	bf := new(bytes.Buffer)
	err := param.Serialize(bf)
	assert.Nil(t, err)
	newParam := &InitShardParam{}
	err = newParam.Deserialize(bf)
	assert.Nil(t, err)
	assert.Equal(t, param, newParam)
}

func TestPeerAmount(t *testing.T) {
	acc := account.NewAccount("")
	param := &PeerAmount{
		PeerPubKey: hex.EncodeToString(keypair.SerializePublicKey(acc.PublicKey)),
		Amount:     157681654,
	}
	bf := new(bytes.Buffer)
	err := param.Serialize(bf)
	assert.Nil(t, err)
	newParam := &PeerAmount{}
	err = newParam.Deserialize(bf)
	assert.Nil(t, err)
	assert.Equal(t, param, newParam)
}

func TestPeerStakeParam(t *testing.T) {
	acc := account.NewAccount("")
	param := &PeerStakeParam{
		ShardId:   common.NewShardIDUnchecked(8),
		PeerOwner: acc.Address,
		Value: &PeerAmount{
			PeerPubKey: hex.EncodeToString(keypair.SerializePublicKey(acc.PublicKey)),
			Amount:     157681654,
		},
	}
	bf := new(bytes.Buffer)
	err := param.Serialize(bf)
	assert.Nil(t, err)
	newParam := &PeerStakeParam{}
	err = newParam.Deserialize(bf)
	assert.Nil(t, err)
	assert.Equal(t, param, newParam)
}

func TestUnfreezeFromShardParam(t *testing.T) {
	acc := account.NewAccount("")
	peer := &PeerAmount{
		PeerPubKey: hex.EncodeToString(keypair.SerializePublicKey(acc.PublicKey)),
		Amount:     157681654,
	}
	param := &UnfreezeFromShardParam{
		ShardId: common.NewShardIDUnchecked(8),
		User:    acc.Address,
		Value:   []*PeerAmount{peer, peer},
	}
	bf := new(bytes.Buffer)
	err := param.Serialize(bf)
	assert.Nil(t, err)
	newParam := &UnfreezeFromShardParam{}
	err = newParam.Deserialize(bf)
	assert.Nil(t, err)
	assert.Equal(t, param, newParam)
}

func TestWithdrawStakeAssetParam(t *testing.T) {
	acc := account.NewAccount("")
	param := &WithdrawStakeAssetParam{
		ShardId: common.NewShardIDUnchecked(8),
		User:    acc.Address,
	}
	bf := new(bytes.Buffer)
	err := param.Serialize(bf)
	assert.Nil(t, err)
	newParam := &WithdrawStakeAssetParam{}
	err = newParam.Deserialize(bf)
	assert.Nil(t, err)
	assert.Equal(t, param, newParam)
}

func TestWithdrawFeeParam(t *testing.T) {
	acc := account.NewAccount("")
	param := &WithdrawFeeParam{
		ShardId: common.NewShardIDUnchecked(8),
		User:    acc.Address,
	}
	bf := new(bytes.Buffer)
	err := param.Serialize(bf)
	assert.Nil(t, err)
	newParam := &WithdrawFeeParam{}
	err = newParam.Deserialize(bf)
	assert.Nil(t, err)
	assert.Equal(t, param, newParam)
}

func TestCommitDPosParam(t *testing.T) {
	shardId := common.NewShardIDUnchecked(10)
	hash, err := common.Uint256FromHexString("5b622cfbde2948ae61242fd5d7ee1c84983459e142339316bdb6ab09faee2e02")
	if err != nil {
		t.Fatal(err)
	}
	param := &CommitDposParam{
		ShardId:   shardId,
		Height:    33180098,
		Hash:      hash,
		FeeAmount: 33636363,
	}
	sink := common.NewZeroCopySink(0)
	param.Serialization(sink)
	source := common.NewZeroCopySource(sink.Bytes())
	newParam := &CommitDposParam{}
	err = newParam.Deserialization(source)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, param.ShardId.ToUint64(), newParam.ShardId.ToUint64())
	assert.Equal(t, param.Height, newParam.Height)
	assert.Equal(t, hash.ToHexString(), newParam.Hash.ToHexString())
	assert.Equal(t, param.FeeAmount, newParam.FeeAmount)
}

func TestSetMinStakeParam(t *testing.T) {
	param := &SetMinStakeParam{
		ShardId: common.NewShardIDUnchecked(8),
		Amount:  2672847629,
	}
	bf := new(bytes.Buffer)
	err := param.Serialize(bf)
	assert.Nil(t, err)
	newParam := &SetMinStakeParam{}
	err = newParam.Deserialize(bf)
	assert.Nil(t, err)
	assert.Equal(t, param, newParam)
}

func TestUserStakeParam(t *testing.T) {
	acc := account.NewAccount("")
	peer := &PeerAmount{
		PeerPubKey: hex.EncodeToString(keypair.SerializePublicKey(acc.PublicKey)),
		Amount:     157681654,
	}
	param := &UserStakeParam{
		ShardId: common.NewShardIDUnchecked(8),
		User:    acc.Address,
		Value:   []*PeerAmount{peer, peer},
	}
	bf := new(bytes.Buffer)
	err := param.Serialize(bf)
	assert.Nil(t, err)
	newParam := &UserStakeParam{}
	err = newParam.Deserialize(bf)
	assert.Nil(t, err)
	assert.Equal(t, param, newParam)
}

func TestChangeMaxAuthorizationParam(t *testing.T) {
	acc := account.NewAccount("")
	param := &ChangeMaxAuthorizationParam{
		ShardId: common.NewShardIDUnchecked(8),
		User:    acc.Address,
		Value: &PeerAmount{
			PeerPubKey: hex.EncodeToString(keypair.SerializePublicKey(acc.PublicKey)),
			Amount:     157681654,
		},
	}
	bf := new(bytes.Buffer)
	err := param.Serialize(bf)
	assert.Nil(t, err)
	newParam := &ChangeMaxAuthorizationParam{}
	err = newParam.Deserialize(bf)
	assert.Nil(t, err)
	assert.Equal(t, param, newParam)
}

func TestChangeProportionParam(t *testing.T) {
	acc := account.NewAccount("")
	param := &ChangeProportionParam{
		ShardId: common.NewShardIDUnchecked(8),
		User:    acc.Address,
		Value: &PeerAmount{
			PeerPubKey: hex.EncodeToString(keypair.SerializePublicKey(acc.PublicKey)),
			Amount:     157681654,
		},
	}
	bf := new(bytes.Buffer)
	err := param.Serialize(bf)
	assert.Nil(t, err)
	newParam := &ChangeProportionParam{}
	err = newParam.Deserialize(bf)
	assert.Nil(t, err)
	assert.Equal(t, param, newParam)
}

func TestDeletePeerParam(t *testing.T) {
	acc := account.NewAccount("")
	peer := hex.EncodeToString(keypair.SerializePublicKey(acc.PubKey()))
	param := &DeletePeerParam{
		ShardId: common.NewShardIDUnchecked(8),
		Peers:   []string{peer, peer},
	}
	bf := new(bytes.Buffer)
	err := param.Serialize(bf)
	assert.Nil(t, err)
	newParam := &DeletePeerParam{}
	err = newParam.Deserialize(bf)
	assert.Nil(t, err)
	assert.Equal(t, param, newParam)
}

func TestPeerExitParam(t *testing.T) {
	acc := account.NewAccount("")
	peer := hex.EncodeToString(keypair.SerializePublicKey(acc.PubKey()))
	param := &PeerExitParam{
		ShardId: common.NewShardIDUnchecked(8),
		Peer:    peer,
	}
	bf := new(bytes.Buffer)
	err := param.Serialize(bf)
	assert.Nil(t, err)
	newParam := &PeerExitParam{}
	err = newParam.Deserialize(bf)
	assert.Nil(t, err)
	assert.Equal(t, param, newParam)
}

func TestGetPeerInfoParam(t *testing.T) {
	param := &GetPeerInfoParam{
		ShardId: common.NewShardIDUnchecked(8),
		View:    47156875,
	}
	bf := new(bytes.Buffer)
	err := param.Serialize(bf)
	assert.Nil(t, err)
	newParam := &GetPeerInfoParam{}
	err = newParam.Deserialize(bf)
	assert.Nil(t, err)
	assert.Equal(t, param, newParam)
}

func TestGetUserStakeInfoParam(t *testing.T) {
	acc := account.NewAccount("")
	param := &GetUserStakeInfoParam{
		View:    47156875,
		ShardId: common.NewShardIDUnchecked(8),
		User:    acc.Address,
	}
	bf := new(bytes.Buffer)
	err := param.Serialize(bf)
	assert.Nil(t, err)
	newParam := &GetUserStakeInfoParam{}
	err = newParam.Deserialize(bf)
	assert.Nil(t, err)
	assert.Equal(t, param, newParam)
}
