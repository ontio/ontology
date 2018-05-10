package account

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/core/types"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func genAccountData() (*AccountData, *keypair.ProtectedKey) {
	var acc = new(AccountData)
	prvkey, pubkey, _ := keypair.GenerateKeyPair(keypair.PK_ECDSA, keypair.P256)
	ta := types.AddressFromPubKey(pubkey)
	address := ta.ToBase58()
	password := []byte("123456")
	prvSectet, _ := keypair.EncryptPrivateKey(prvkey, address, password)
	h := sha256.Sum256(password)
	acc.SetKeyPair(prvSectet)
	acc.SigSch = "SHA256withECDSA"
	acc.PubKey = hex.EncodeToString(keypair.SerializePublicKey(pubkey))
	acc.PassHash = hex.EncodeToString(h[:])
	return acc, prvSectet
}

func TestAccountData(t *testing.T) {
	acc, prvSectet := genAccountData()
	assert.NotNil(t, acc)
	assert.Equal(t, acc.Address, acc.ProtectedKey.Address)
	assert.Equal(t, prvSectet, acc.GetKeyPair())
	assert.True(t, acc.VerifyPassword([]byte("123456")))
}

func TestWalletStorage(t *testing.T) {
	defer func() {
		os.Remove(WALLET_FILENAME)
		os.RemoveAll("Log/")
	}()

	wallet := new(WalletData)
	wallet.Inititalize()
	wallet.Save(WALLET_FILENAME)
	walletReadFromFile := new(WalletData)
	walletReadFromFile.Load(WALLET_FILENAME)
	assert.Equal(t, walletReadFromFile, wallet)
	accout, _ := genAccountData()
	wallet.AddAccount(accout)
	wallet.AddAccount(accout)
	wallet.Save(WALLET_FILENAME)
	walletReadFromFile.Load(WALLET_FILENAME)
	assert.Equal(t, walletReadFromFile, wallet)
	wallet.DelAccount(2)
	assert.Equal(t, 1, len(wallet.Accounts))
	assert.Panics(t, func() { wallet.DelAccount(2) })
	wallet.Save(WALLET_FILENAME)
	walletReadFromFile.Load(WALLET_FILENAME)
	assert.Equal(t, walletReadFromFile, wallet)
	defaultAccount := wallet.GetDefaultAccount()
	assert.Equal(t, defaultAccount, accout)
}
