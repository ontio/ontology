package account

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/core/types"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"sort"
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

func TestWalletSave(t *testing.T) {
	walletFile := "w.data"
	defer func() {
		os.Remove(walletFile)
		os.RemoveAll("Log/")
	}()

	wallet := NewWalletData()
	size := 10
	for i := 0; i < size; i++ {
		acc, _ := genAccountData()
		wallet.AddAccount(acc)
		err := wallet.Save(walletFile)
		if err != nil {
			t.Errorf("Save error:%s", err)
			return
		}
	}

	wallet2 := NewWalletData()
	err := wallet2.Load(walletFile)
	if err != nil {
		t.Errorf("Load error:%s", err)
		return
	}

	assert.Equal(t, len(wallet2.Accounts), len(wallet.Accounts))
}

func TestWalletDel(t *testing.T){
	wallet := NewWalletData()
	size := 10
	accList := make([]string, 0,size)
	for i := 0; i < size; i++ {
		acc, _ := genAccountData()
		wallet.AddAccount(acc)
		accList = append(accList, acc.Address)
	}
	sort.Strings(accList)
	for _, address := range accList{
		wallet.DelAccount(address)
		_, index := wallet.GetAccountByAddress(address)
		if ! assert.Equal(t, -1, index){
			return
		}
	}
}