package shardstates

import (
	"encoding/hex"
	"github.com/magiconair/properties/assert"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
	"testing"
)

func TestEncodeByte(t *testing.T) {
	acc := account.NewAccount("")
	peerPK := hex.EncodeToString(keypair.SerializePublicKey(acc.PublicKey))
	t.Log(peerPK)
	assert.Equal(t, acc.PublicKey, acc.PublicKey)

	data, _:= hex.DecodeString(peerPK)
	genPubKey, _:= keypair.DeserializePublicKey(data)
	assert.Equal(t, genPubKey, acc.PublicKey)
}
