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
package ontid

import (
	"fmt"
	"testing"

	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native"
)

func TestAuthentication(t *testing.T) {
	testcase(t, CaseAuthentication)
}

func CaseAuthentication(t *testing.T, n *native.NativeService) {
	id, err := account.GenerateID()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(id)
	acc := account.NewAccount("")
	if regID(n, id, acc) != nil {
		t.Fatal("register id error")
	}

	authKeyParam := &SetAuthKeyParam{
		OntId:     []byte(id),
		Index:     1,
		SignIndex: 1,
	}
	// 2a6469643a6f6e743a5458625237696f58725a67456571536e696b3843444a3955666d7757505856584a360b736f6d6553657276696365037373730e687474703b3b733b733b733b3b73010e687474703b3b733b733b733b3b73

	sink := common.NewZeroCopySink(nil)
	authKeyParam.Serialization(sink)
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{acc.Address}
	_, err = setAuthKey(n)
	fmt.Println(common.ToHexString(sink.Bytes()))
	if err != nil {
		t.Fatal()
	}
	encId, _ := encodeID([]byte(id))
	res, err := getAuthentication(n, encId)
	if err != nil {
		t.Fatal()
	}
	fmt.Println(res)

	//OntId     []byte
	//Index     uint32
	//SignIndex uint32

	removeAuthKeyParam := &RemoveAuthKeyParam{
		OntId:     []byte(id),
		Index:     1,
		SignIndex: 1,
	}
	sink = common.NewZeroCopySink(nil)
	removeAuthKeyParam.Serialization(sink)
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{acc.Address}
	_, err = removeAuthKey(n)
	if err != nil {
		t.Fatal()
	}

	res, err = getAuthentication(n, encId)
	if err != nil {
		t.Fatal()
	}
	fmt.Println(res)

}
