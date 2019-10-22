package testsuite

import (
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native"
	_ "github.com/ontio/ontology/smartcontract/service/native/init"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/ontio/ontology/smartcontract/storage"
	"github.com/stretchr/testify/assert"

	"testing"
)

func setOntBalance(db *storage.CacheDB, addr common.Address, value uint64) {
	balanceKey := ont.GenBalanceKey(utils.OntContractAddress, addr)
	item := utils.GenUInt64StorageItem(value)
	db.Put(balanceKey, item.ToArray())
}

func ontBalanceOf(native *native.NativeService, addr common.Address) int {
	sink := common.NewZeroCopySink(nil)
	utils.EncodeAddress(sink, addr)
	native.Input = sink.Bytes()
	buf, _ := ont.OntBalanceOf(native)
	val := common.BigIntFromNeoBytes(buf)
	return int(val.Uint64())
}

func ontTransfer(native *native.NativeService, from, to common.Address, value uint64) error {
	native.Tx.SignedAddr = append(native.Tx.SignedAddr, from)

	state := ont.State{from, to, value}
	native.Input = common.SerializeToBytes(&ont.Transfers{States: []ont.State{state}})

	_, err := ont.OntTransfer(native)
	return err
}

func TestTransfer(t *testing.T) {
	InvokeNativeContract(t, utils.OntContractAddress, func(native *native.NativeService) ([]byte, error) {
		a := RandomAddress()
		b := RandomAddress()
		c := RandomAddress()
		setOntBalance(native.CacheDB, a, 10000)

		assert.Equal(t, ontBalanceOf(native, a), 10000)
		assert.Equal(t, ontBalanceOf(native, b), 0)
		assert.Equal(t, ontBalanceOf(native, c), 0)

		assert.Nil(t, ontTransfer(native, a, b, 10))
		assert.Equal(t, ontBalanceOf(native, a), 9990)
		assert.Equal(t, ontBalanceOf(native, b), 10)

		assert.Nil(t, ontTransfer(native, b, c, 10))
		assert.Equal(t, ontBalanceOf(native, b), 0)
		assert.Equal(t, ontBalanceOf(native, c), 10)

		return nil, nil
	})
}
