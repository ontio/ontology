package common

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	ethcomm "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	sysconfig "github.com/ontio/ontology/common/config"
	txtypes "github.com/ontio/ontology/core/types"
	"github.com/stretchr/testify/assert"
)

func genTxWithNonceAndPrice(nonce uint64, gp int64) *txtypes.Transaction {
	privateKey, _ := crypto.HexToECDSA("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d5817ac83d38b6a19")

	value := big.NewInt(1000000000)
	gaslimit := uint64(21000)
	gasPrice := big.NewInt(gp)

	toAddress := ethcomm.HexToAddress("0x4592d8f8d7b001e72cb26a73e4fa1806a51ac79d")
	//toontAddress, _ := common.AddressParseFromBytes(toAddress[:])
	//fmt.Printf("to ont addr:%s\n", toontAddress.ToBase58())

	var data []byte
	tx := types.NewTransaction(nonce, toAddress, value, gaslimit, gasPrice, data)

	chainId := big.NewInt(int64(sysconfig.DefConfig.P2PNode.EVMChainId))
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainId), privateKey)
	if err != nil {
		fmt.Printf("err:%s\n", err.Error())
		return nil
	}

	//bt, _ := rlp.EncodeToBytes(signedTx)
	//fmt.Printf("rlptx:%s", hex.EncodeToString(bt))

	otx, err := txtypes.TransactionFromEIP155(signedTx)
	if err != nil {
		fmt.Printf("err:%s\n", err.Error())
		return nil
	}
	return otx
}

func TestTxSortedTimeMap(t *testing.T) {

	tsm := newTxSortedTimeMap()
	tx := genTxWithNonceAndPrice(0, 2500)
	//fmt.Printf("n0:%s\n",tx.Hash().ToHexString())
	tsm.Put(tx)
	assert.Equal(t, tsm.Len(), 1)
	tsm.Put(tx)
	time.Sleep(2 * time.Second)

	assert.Equal(t, tsm.Len(), 1)

	tx2 := genTxWithNonceAndPrice(1, 2500)
	//fmt.Printf("n1:%s\n",tx2.Hash().ToHexString())

	tsm.Put(tx2)
	time.Sleep(2 * time.Second)

	assert.Equal(t, tsm.Len(), 2)

	tx3 := genTxWithNonceAndPrice(2, 2500)
	//fmt.Printf("n2:%s\n",tx3.Hash().ToHexString())

	tsm.Put(tx3)
	time.Sleep(2 * time.Second)

	tx4 := genTxWithNonceAndPrice(3, 2500)
	tsm.Put(tx4)
	time.Sleep(2 * time.Second)

	//fmt.Printf("n3:%s\n",tx4.Hash().ToHexString())
	assert.Equal(t, tsm.Len(), 4)

	tsm.Remove(tx2.Hash())

	//for k,v := range tsm.items {
	//	fmt.Printf("key:%s,val:%d\n",k.ToHexString() ,v.Time)
	//}

	assert.Equal(t, tsm.Len(), 3)

	assert.NotNil(t, tsm.Get(tx.Hash()))
	assert.Nil(t, tsm.Get(tx2.Hash()))
	assert.NotNil(t, tsm.Get(tx3.Hash()))
	assert.NotNil(t, tsm.Get(tx4.Hash()))

	tmp := tsm.Get(tx3.Hash()).Time

	exp := tsm.ExpiredTxByTime(tmp)
	//fmt.Printf("t3.time:%d\n",tmp)
	assert.Equal(t, len(exp), 2)
	//fmt.Println("before delete")
	//for k,v := range tsm.items {
	//	fmt.Printf("key:%s,val:%d\n",k.ToHexString() ,v.Time)
	//}

	for _, e := range exp {
		tsm.Remove(e.Tx.Hash())
	}
	//fmt.Println("after delete")
	//for k,v := range tsm.items {
	//	fmt.Printf("key:%s,val:%d\n",k.ToHexString() ,v.Time)
	//}
	assert.Equal(t, len(tsm.items), 1)
	assert.NotNil(t, tsm.Get(tx4.Hash()))

}

func TestTxSortedTimeMap_LastElement(t *testing.T) {
	tsm := newTxSortedTimeMap()
	tx := genTxWithNonceAndPrice(0, 2500)
	tsm.Put(tx)
	time.Sleep(2 * time.Second)

	tx1 := genTxWithNonceAndPrice(1, 2500)
	tsm.Put(tx1)
	time.Sleep(2 * time.Second)

	tx2 := genTxWithNonceAndPrice(2, 2500)
	tsm.Put(tx2)
	time.Sleep(2 * time.Second)

	tx3 := genTxWithNonceAndPrice(3, 2500)
	tsm.Put(tx3)
	time.Sleep(2 * time.Second)

	assert.Equal(t, tsm.Len(), 4)
	tmp := tsm.LastElement()
	assert.Equal(t, tmp.Tx.Hash(), tx3.Hash())

	tx4 := genTxWithNonceAndPrice(4, 2500)
	tsm.Put(tx4)

	assert.Equal(t, tsm.Len(), 5)
	tmp = tsm.LastElement()
	assert.Equal(t, tmp.Tx.Hash(), tx4.Hash())
}
