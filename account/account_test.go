package account

import (
	"github.com/ontio/ontology/core/types"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestNewAccount(t *testing.T) {
	defer func() {
		os.RemoveAll("Log/")
	}()

	names := []string{
		"",
		"SHA224withECDSA",
		"SHA256withECDSA",
		"SHA384withECDSA",
		"SHA512withECDSA",
		"SHA3-224withECDSA",
		"SHA3-256withECDSA",
		"SHA3-384withECDSA",
		"SHA3-512withECDSA",
		"RIPEMD160withECDSA",
		"SM3withSM2",
		"SHA512withEdDSA",
	}
	accounts := make([]*Account, len(names))
	for k, v := range names {
		accounts[k] = NewAccount(v)
		assert.NotNil(t, accounts[k])
		assert.NotNil(t, accounts[k].PrivateKey)
		assert.NotNil(t, accounts[k].PublicKey)
		assert.NotNil(t, accounts[k].Address)
		assert.NotNil(t, accounts[k].PrivKey())
		assert.NotNil(t, accounts[k].PubKey())
		assert.NotNil(t, accounts[k].Scheme())
		assert.Equal(t, accounts[k].Address, types.AddressFromPubKey(accounts[k].PublicKey))
	}
}
