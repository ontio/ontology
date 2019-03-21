package shardgas

import (
	"bytes"
	"encoding/hex"
	"github.com/magiconair/properties/assert"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/core/types"
	"testing"
)

func TestPeerWithdrawGasParam(t *testing.T) {
	acc := account.NewAccount("")
	sourceId := types.NewShardIDUnchecked(0)
	param := &PeerWithdrawGasParam{
		Signer:     acc.Address,
		PeerPubKey: hex.EncodeToString(keypair.SerializePublicKey(acc.PublicKey)),
		User:       acc.Address,
		ShardId:    sourceId,
		Amount:     10000,
		WithdrawId: 1,
	}
	bf := new(bytes.Buffer)
	err := param.Serialize(bf)
	if err != nil {
		t.Fatal(err)
	}
	newParam := &PeerWithdrawGasParam{}
	err = newParam.Deserialize(bf)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, param, newParam)
}
