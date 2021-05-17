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
package common

import (
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"gotest.tools/assert"
	"sort"
	"testing"
)

func Test_sortTxEntry(t *testing.T) {
	addr1 := utils.OntContractAddress
	addr2 := utils.OngContractAddress

	txlist := []*TXEntry{
		{Tx: &types.Transaction{TxType: types.InvokeWasm, Nonce: 0, Payer: addr1, GasPrice: 100, Raw: []byte("tx01")}},
		{Tx: &types.Transaction{TxType: types.InvokeNeo, Nonce: 0, Payer: addr2, GasPrice: 110, Raw: []byte("tx02")}},
		{Tx: &types.Transaction{TxType: types.EIP155, Nonce: 0, Payer: addr1, GasPrice: 130, Raw: []byte("tx03")}},
		{Tx: &types.Transaction{TxType: types.EIP155, Nonce: 1, Payer: addr1, GasPrice: 100, Raw: []byte("tx04")}},
		{Tx: &types.Transaction{TxType: types.EIP155, Nonce: 0, Payer: addr2, GasPrice: 150, Raw: []byte("tx05")}},
	}

	sort.Sort(OrderByNetWorkFee(txlist))

	//newList := make([]*TXEntry,len(txlist))
	//for i,t := range txlist {
	//	fmt.Printf("%d,type:%v,nonce:%d,payer:%s,gasprice:%d,raw:%s\n",i,t.Tx.TxType,t.Tx.Nonce,t.Tx.Payer,t.Tx.GasPrice,t.Tx.Raw)
	//	//newList[i] = t
	//}
	assert.Equal(t, string(txlist[0].Tx.Raw), "tx05")
	assert.Equal(t, string(txlist[1].Tx.Raw), "tx03")
	assert.Equal(t, string(txlist[2].Tx.Raw), "tx04")
	assert.Equal(t, string(txlist[3].Tx.Raw), "tx02")
	assert.Equal(t, string(txlist[4].Tx.Raw), "tx01")

}
