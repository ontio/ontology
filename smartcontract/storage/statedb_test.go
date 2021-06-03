// Copyright (C) 2021 The Ontology Authors

package storage

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	comm "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/store/leveldbstore"
	"github.com/ontio/ontology/core/store/overlaydb"
	"github.com/stretchr/testify/require"
)

type dummy struct{}

func (d dummy) SubBalance(cache *CacheDB, addr comm.Address, val *big.Int) error {
	return nil
}
func (d dummy) AddBalance(cache *CacheDB, addr comm.Address, val *big.Int) error {
	return nil
}
func (d dummy) SetBalance(cache *CacheDB, addr comm.Address, val *big.Int) error {
	return nil
}
func (d dummy) GetBalance(cache *CacheDB, addr comm.Address) (*big.Int, error) {
	return big.NewInt(0), nil
}

var _ OngBalanceHandle = dummy{}

func TestEtherAccount(t *testing.T) {
	a := require.New(t)

	memback := leveldbstore.NewMemLevelDBStore()
	overlay := overlaydb.NewOverlayDB(memback)

	cache := NewCacheDB(overlay)
	// don't consider ong yet
	sd := NewStateDB(cache, common.Hash{}, common.Hash{}, dummy{})
	a.NotNil(sd, "fail")

	h := crypto.Keccak256Hash([]byte("hello"))

	ea := &EthAcount{
		Nonce:    1023,
		CodeHash: h,
	}
	a.False(ea.IsEmpty(), "expect not empty")

	sink := comm.NewZeroCopySink(nil)
	ea.Serialization(sink)

	clone := &EthAcount{}
	source := comm.NewZeroCopySource(sink.Bytes())
	err := clone.Deserialization(source)
	a.Nil(err, "fail")
	a.Equal(clone, ea, "fail")

	pri, err := crypto.GenerateKey()
	a.Nil(err, "fail")
	ethAddr := crypto.PubkeyToAddress(pri.PublicKey)

	sd.cacheDB.PutEthAccount(ethAddr, *ea)

	getea := sd.getEthAccount(ethAddr)
	a.Equal(getea, *clone, "fail")

	a.Equal(sd.GetNonce(ethAddr), ea.Nonce, "fail")
	sd.SetNonce(ethAddr, 1024)
	a.Equal(sd.GetNonce(ethAddr), ea.Nonce+1, "fail")
	// don't effect code hash
	a.Equal(sd.getEthAccount(ethAddr).CodeHash, ea.CodeHash, "fail")

	sd.SetCode(ethAddr, []byte("hello again"))
	a.Equal(sd.GetCodeHash(ethAddr), crypto.Keccak256Hash([]byte("hello again")), "fail")
	a.Equal(sd.GetCode(ethAddr), []byte("hello again"), "fail")

	a.False(sd.HasSuicided(ethAddr), "fail")
	ret := sd.Suicide(ethAddr)
	a.True(ret, "fail")
	a.True(sd.HasSuicided(ethAddr), "fail")

	// nonexist account get ==> default value
	pri2, _ := crypto.GenerateKey()
	anotherAddr := crypto.PubkeyToAddress(pri2.PublicKey)

	nonce := sd.GetNonce(anotherAddr)
	a.Equal(nonce, uint64(0), "fail")
	hash := sd.GetCodeHash(anotherAddr)
	a.Equal(hash, common.Hash{}, "fail")

	sd.SetNonce(anotherAddr, 1)
	a.Equal(sd.GetNonce(anotherAddr), uint64(1), "fail")
}
