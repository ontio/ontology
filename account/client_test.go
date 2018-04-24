package account

import (
	"os"
	"testing"
	"github.com/ontio/ontology-crypto/keypair"
)

var (
	testPath = "./wallet_test.dat"
	testEncrypt = "SHA256withECDSA"
	testPwd = []byte("test")
	testClient  *ClientImpl
)

func TestMain(m *testing.M) {
	testClient = NewClient(testPath, testPwd, true)
	m.Run()
	os.RemoveAll(testPath)
	os.RemoveAll("ActorLog")
}

func TestAccount(t *testing.T) {
	accNum := testClient.GetAccountNum()

	count := 100
	accs := make([]*Account, 0)
	for i := 0; i < count; i++ {
		acc, err := testClient.CreateAccount(testEncrypt)
		if err != nil {
			t.Errorf("TestCreateAccount CreateAccount error:%s", err)
			return
		}
		accs = append(accs, acc)
	}
	if testClient.GetAccountNum() != accNum + count {
		t.Errorf("TestCreateAccount AccountNum:%d != %d", testClient.GetAccountNum(), accNum + count)
		return
	}

	acc0 := testClient.GetAccountByIndex(0)
	defAcc := testClient.GetDefaultAccount()
	if acc0.Address.ToHexString() != defAcc.Address.ToHexString() {
		t.Errorf("GetDefaultAccount error default: %x != index_0 %x", defAcc.Address, acc0.Address)
		return
	}

	accAddr := testClient.GetAccountByAddress(defAcc.Address)
	if accAddr.Address.ToHexString() != defAcc.Address.ToHexString() {
		t.Errorf("GetDefaultAccount error accout: %x != default %x", defAcc.Address, acc0.Address)
		return
	}
}

func TestSaveAccount(t *testing.T) {
	accIndex := testClient.GetAccountNum()
	acc := NewAccount(testEncrypt)
	err := testClient.SaveAccount(acc)
	if err != nil {
		t.Errorf("SaveAccount error:%s", err)
		return
	}

	client1 := NewClient(testPath, testPwd, false)
	client1.LoadAccount()
	acc1 := client1.GetAccountByIndex(accIndex)
	if acc1 == nil {
		t.Errorf("TestSaveAccount GetAccountByIndex:%d failed", accIndex)
		return
	}

	if acc.Address.ToHexString() != acc1.Address.ToHexString() {
		t.Errorf("TestSaveAccount acount address:%x != %x", acc1.Address, acc.Address)
		return
	}
}

func TestCreateAccountByPrivateKey(t *testing.T){
	ps, pk, err := keypair.GenerateKeyPair(keypair.PK_ECDSA,  keypair.P256)
	if err != nil {
		t.Errorf("TestCreateAccountByPrivateKey GenerateKeyPair error:%s", err)
		return
	}
	psData := keypair.SerializePrivateKey(ps)
	acc, err := testClient.CreateAccountByPrivateKey(psData)
	if err != nil {
		 t.Errorf("CreateAccountByPrivateKey error:%s", err)
		 return
	}

	pkData := keypair.SerializePublicKey(pk)

	pkData1 := keypair.SerializePublicKey(acc.PublicKey)
	if string(pkData) != string(pkData1) {
		t.Errorf("TestCreateAccountByPrivateKey pk:%x != %x", pkData, pkData1)
		return
	}

	acc1 := testClient.GetAccountByAddress(acc.Address)
	if acc1 == nil {
		t.Errorf("TestCreateAccountByPrivateKey cannot GetAccountByAddress:%x", acc.Address)
		return
	}

	if acc.Address.ToHexString() != acc1.Address.ToHexString() {
		t.Errorf("TestCreateAccountByPrivateKey Address:%x != %x", acc1.Address, acc.Address)
		return
	}
}

func TestChangePassword(t *testing.T){
	newPwd := append(testPwd, byte(1))
	res := testClient.ChangePassword(testPwd, newPwd)
	if !res {
		t.Errorf("TestChangePassword ChangePassword failed")
		return
	}

	testPwd = newPwd
	client1 := NewClient(testPath, testPwd, false)
	if client1 == nil {
		t.Errorf("NewClient failed")
		return
	}
}