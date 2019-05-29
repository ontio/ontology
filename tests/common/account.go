package TestCommon

import (
	"os"
	"testing"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology-crypto/signature"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/cmd/common"
	"github.com/ontio/ontology/tests"
)

const PASSWORD = "1"

var allAccounts map[string]*account.Account

func init() {
	TestConsts.TestRootDir = "../"
	allAccounts = make(map[string]*account.Account)
}

func GetAccount(name string) *account.Account {
	return allAccounts[name]
}

func OpenAccount(t *testing.T, name string) {
	walletDir := TestConsts.TestRootDir + "./wallets/"
	walletFile := walletDir + name + ".dat"
	wallet, err := account.Open(walletFile)
	if err != nil {
		t.Fatalf("failed to open wallet file %s, %s", walletFile, err)
	}

	acc, err := common.GetAccountMulti(wallet, []byte(PASSWORD), name)
	if err != nil {
		t.Fatalf("failed to open account: %s", err)
	}

	allAccounts[name] = acc
}

func CreateAccount(t *testing.T, name string) {
	walletDir := TestConsts.TestRootDir + "./wallets/"
	if _, err := os.Stat(walletDir); os.IsNotExist(err) {
		os.Mkdir(walletDir, 0755)
	}

	walletFile := walletDir + name + ".dat"
	_, err := os.Stat(walletFile)
	if err == nil {
		OpenAccount(t, name)
		return
	} else if os.IsNotExist(err) {
		wallet, err := account.Open(walletFile)
		if err != nil {
			t.Fatalf(err.Error())
		}
		acc, err := wallet.NewAccount(name, keypair.PK_ECDSA, keypair.P256, signature.SHA256withECDSA, []byte(PASSWORD))
		if err != nil {
			t.Fatalf(err.Error())
		}
		allAccounts[name] = acc
		return
	}
	t.Fatalf("create account err: %s", err)
}
