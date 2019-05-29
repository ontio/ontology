
package TestCommon

import (
	"testing"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/core/utils"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology-crypto/signature"
	"github.com/ontio/ontology/core/chainmgr"
	"fmt"
	utils2 "github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/account"
)

func CreateAdminTx(t *testing.T, shard common.ShardID, addr common.Address, method string, args []byte) *types.Transaction {
	mutable := utils.BuildNativeTransaction(addr, method, args)
	mutable.GasPrice = 0
	mutable.GasLimit = 20000
	mutable.Nonce = 123456

	shardName := chainmgr.GetShardName(shard)
	pks := make([]keypair.PublicKey, 0)
	accounts := make([]*account.Account, 0)
	for i := 0; i < 7; i++ {
		acc := GetAccount(shardName + "_peerOwner" + fmt.Sprintf("%d",i))
		pks = append(pks, acc.PublicKey)
		accounts = append(accounts, acc)
	}

	for _, acc := range accounts {
		if err := utils2.MultiSigTransaction(mutable, 5, pks, acc); err != nil {
			t.Fatalf("multi sign tx: %s", err)
		}
	}

	tx, err := mutable.IntoImmutable()
	if err != nil {
		t.Fatalf("to immutable tx: %s", err)
	}

	return tx
}

func CreateNativeTx(t *testing.T, user string, addr common.Address, method string, args []byte) *types.Transaction {
	acc := GetAccount(user)
	if acc == nil {
		t.Fatalf("Invalid user: %s", user)
	}

	mutable := utils.BuildNativeTransaction(addr, method, args)
	mutable.GasPrice = 0
	mutable.GasLimit = 20000
	mutable.Payer = acc.Address
	mutable.Nonce = 123456

	txHash := mutable.Hash()
	sig, err := signature.Sign(acc.SigScheme, acc.PrivateKey, txHash.ToArray(), nil)
	if err != nil {
		t.Fatalf("sign tx: %s", err)
	}
	sigData, err := signature.Serialize(sig)
	if err != nil {
		t.Fatalf("serialize sig: %s", err)
	}
	mutable.Sigs = []types.Sig{
		{
			PubKeys: []keypair.PublicKey{acc.PubKey()},
			M:       1,
			SigData: [][]byte{sigData},
		},
	}
	tx, err := mutable.IntoImmutable()
	if err != nil {
		t.Fatalf("to immutable tx: %s", err)
	}

	return tx
}

func CreateNeoInvokeTx(t *testing.T, user string, addr common.Address, params []interface{}) *types.Transaction {
	return nil
}

func CreateNeoDeployTx(t *testing.T, user string, code []byte, contractName string) *types.Transaction {
	return nil
}
