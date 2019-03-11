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
	for i := 0; i < 100; i++ {
		testPK := hex.EncodeToString(keypair.SerializePublicKey(acc.PublicKey))
		t.Log(testPK)
		assert.Equal(t, peerPK, testPK)
	}
}
