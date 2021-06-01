/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package types

import (
	"crypto/ecdsa"
	"fmt"
	ethcomm "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ontio/ontology/common"
	"math"
	"math/big"
	"testing"

	"github.com/ontio/ontology/core/payload"
	"github.com/stretchr/testify/assert"
)

func TestTransaction_SigHashForChain(t *testing.T) {
	mutable := &MutableTransaction{
		TxType:  InvokeNeo,
		Payload: &payload.InvokeCode{},
	}

	tx, err := mutable.IntoImmutable()
	assert.Nil(t, err)

	assert.Equal(t, tx.Hash(), tx.SigHashForChain(0))
	assert.NotEqual(t, tx.Hash(), tx.SigHashForChain(1))
	assert.NotEqual(t, tx.Hash(), tx.SigHashForChain(math.MaxUint32))
}

func genTx(nonce uint64) *Transaction {
	privateKey, _ := crypto.HexToECDSA("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d5817ac83d38b6a19")
	//assert.Nil(t, err)

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	fmt.Printf("addr:%s\n", fromAddress.Hex())

	ontAddress, _ := common.AddressParseFromBytes(fromAddress[:])
	//assert.Nil(t, err)
	fmt.Printf("ont addr:%s\n", ontAddress.ToBase58())

	value := big.NewInt(1000000000)
	gaslimit := uint64(21000)
	gasPrice := big.NewInt(2500)

	toAddress := ethcomm.HexToAddress("0x4592d8f8d7b001e72cb26a73e4fa1806a51ac79d")

	var data []byte
	tx := types.NewTransaction(nonce, toAddress, value, gaslimit, gasPrice, data)

	chainId := big.NewInt(0)
	signedTx, _ := types.SignTx(tx, types.NewEIP155Signer(chainId), privateKey)
	//assert.Nil(t, err)

	otx, _ := TransactionFromEIP155(signedTx)
	//assert.Nil(t, err)
	return otx
}

func Test_EIP155Tx(t *testing.T) {
	otx := genTx(0)

	sink := common.ZeroCopySink{}
	otx.Serialization(&sink)

	tx, err := TransactionFromRawBytes(sink.Bytes())
	assert.Nil(t, err)

	assert.NotNil(t, tx)

	assert.Equal(t, otx.Nonce, tx.Nonce)
	assert.Equal(t, otx.Payer, tx.Payer)
	assert.Equal(t, otx.GasLimit, tx.GasLimit)
	assert.Equal(t, otx.GasPrice, tx.GasPrice)
	assert.Equal(t, otx.TxType, tx.TxType)
	assert.Equal(t, otx.Version, tx.Version)
	assert.Equal(t, otx.Raw, tx.Raw)
	//assert.Equal(t,otx.Payload.(*payload.EIP155Code).EIPTx.,tx.Raw)

}
