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
package proc

import (
	"math/big"
	"testing"

	ethcomm "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	sysconfig "github.com/ontio/ontology/common/config"
	txtypes "github.com/ontio/ontology/core/types"
	"github.com/stretchr/testify/assert"
)

func initCfg() {
	sysconfig.DefConfig.P2PNode.EVMChainId = 12345
}

func genTxWithNonceAndPrice(nonce uint64, gp int64) *txtypes.Transaction {
	privateKey, _ := crypto.HexToECDSA("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d5817ac83d38b6a19")
	//assert.Nil(t, err)
	//
	//publicKey := privateKey.Public()
	//publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	//if !ok {
	//	return nil
	//}
	//
	//fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	//fmt.Printf("addr:%s\n", fromAddress.Hex())
	//
	//ontAddress, _ := common.AddressParseFromBytes(fromAddress[:])
	//fmt.Printf("ont addr:%s\n", ontAddress.ToBase58())

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
		//fmt.Printf("err:%s\n", err.Error())
		return nil
	}

	//bt, _ := rlp.EncodeToBytes(signedTx)
	//fmt.Printf("rlptx:%s", hex.EncodeToString(bt))

	otx, err := txtypes.TransactionFromEIP155(signedTx)
	if err != nil {
		//fmt.Printf("err:%s\n", err.Error())
		return nil
	}
	return otx
}

func Test_ethtxRLP(t *testing.T) {
	initCfg()
	genTxWithNonceAndPrice(1, 2500)

}

func Test_From(t *testing.T) {
	initCfg()
	otx1 := genTxWithNonceAndPrice(0, 2500)
	otx2 := genTxWithNonceAndPrice(0, 2500)
	otx3 := genTxWithNonceAndPrice(0, 3000)
	otx4 := genTxWithNonceAndPrice(1, 2500)
	assert.Equal(t, otx1.Payer, otx2.Payer)
	assert.Equal(t, otx2.Payer, otx3.Payer)
	assert.Equal(t, otx3.Payer, otx4.Payer)
	assert.Equal(t, otx1.Payer, otx4.Payer)
}
