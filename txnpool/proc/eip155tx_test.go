package proc

import (
	"crypto/ecdsa"
	"fmt"
	ethcomm "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ontio/ontology/common"
	txtypes "github.com/ontio/ontology/core/types"
	tc "github.com/ontio/ontology/txnpool/common"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
	"time"
)

func Test_GenEIP155tx(t *testing.T) {
	privateKey, err := crypto.HexToECDSA("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d5817ac83d38b6a19")
	assert.Nil(t, err)

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		assert.True(t, ok)
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	fmt.Printf("addr:%s\n", fromAddress.Hex())

	ontAddress, err := common.AddressParseFromBytes(fromAddress[:])
	assert.Nil(t, err)
	fmt.Printf("ont addr:%s\n", ontAddress.ToBase58())

	value := big.NewInt(1000000000)
	gaslimit := uint64(21000)
	gasPrice := big.NewInt(2500)
	nonce := uint64(0)

	toAddress := ethcomm.HexToAddress("0x4592d8f8d7b001e72cb26a73e4fa1806a51ac79d")

	var data []byte
	tx := types.NewTransaction(nonce, toAddress, value, gaslimit, gasPrice, data)

	chainId := big.NewInt(0)
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainId), privateKey)
	assert.Nil(t, err)

	otx, err := txtypes.TransactionFromEIP155(signedTx)
	assert.Nil(t, err)

	fmt.Printf("1. otx.payer:%s\n",otx.Payer.ToBase58())

	assert.True(t, otx.TxType == txtypes.EIP155)

	t.Log("Starting test tx")
	var s *TXPoolServer
	s = NewTxPoolServer(tc.MAX_WORKER_NUM, true, false)
	if s == nil {
		t.Error("Test case: new tx pool server failed")
		return
	}
	defer s.Stop()

	f := s.assignTxToWorker(otx, sender, nil)
	assert.True(t, f)

	time.Sleep(10 * time.Second)
	txEntry := &tc.TXEntry{
		Tx:    txn,
		Attrs: []*tc.TXAttr{},
	}
	fmt.Printf("before %s nonce is :%d\n",ontAddress.ToBase58(),s.pendingNonces.get(ontAddress))
	f = s.addTxList(txEntry)
	assert.True(t, f)
	fmt.Printf("after %s nonce is :%d\n",ontAddress.ToBase58(),s.pendingNonces.get(ontAddress))

	ret := s.checkTx(txn.Hash())
	if ret == false {
		t.Error("Failed to check the tx")
		return
	}

	entry := s.getTransaction(txn.Hash())
	if entry == nil {
		t.Error("Failed to get the transaction")
		return
	}

	pendingNonce := s.pendingNonces.get(ontAddress)
	assert.Equal(t, pendingNonce,1)

	t.Log("Ending test tx")

}
