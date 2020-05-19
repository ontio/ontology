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

func TestService(t *testing.T) {
	testcase(t, CaseService)
}

func CaseService(t *testing.T, n *native.NativeService) {
	id, err := account.GenerateID()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(id)
	acc := account.NewAccount("")
	if regID(n, id, acc) != nil {
		t.Fatal("register id error")
	}
	service := &ServiceParam{
		OntId:          []byte(id),
		ServiceId:      []byte("someService"),
		Type:           []byte("sss"),
		ServiceEndpint: []byte("http;;s;s;s;;s"),
		Index:          1,
	}

	sink := common.NewZeroCopySink(nil)
	service.Serialization(sink)
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{acc.Address}
	_, err = addService(n)
	fmt.Println(common.ToHexString(sink.Bytes()))
	if err != nil {
		t.Fatal()
	}
	encId, _ := encodeID([]byte(id))
	res, err := getServicesJson(n, encId)
	if err != nil {
		t.Fatal()
	}
	for i := 0; i < len(res); i++ {
		fmt.Println(res[i])
	}
	service = &ServiceParam{
		OntId:          []byte(id),
		ServiceId:      []byte("someService"),
		Type:           []byte("sss"),
		ServiceEndpint: []byte("http;;s;s;s;;ssssss"),
		Index:          1,
	}
	sink = common.NewZeroCopySink(nil)
	service.Serialization(sink)
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{acc.Address}
	_, err = updateService(n)
	fmt.Println(common.ToHexString(sink.Bytes()))
	if err != nil {
		t.Fatal()
	}
	encId, _ = encodeID([]byte(id))
	res, err = getServicesJson(n, encId)
	if err != nil {
		t.Fatal()
	}
	for i := 0; i < len(res); i++ {
		fmt.Println(res[i])
	}

	serviceRemove := &ServiceRemoveParam{
		OntId:     []byte(id),
		ServiceId: []byte("someService"),
		Index:     1,
	}
	sink = common.NewZeroCopySink(nil)
	serviceRemove.Serialization(sink)
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{acc.Address}
	_, err = removeService(n)
	fmt.Println(common.ToHexString(sink.Bytes()))
	if err != nil {
		t.Fatal()
	}
	encId, _ = encodeID([]byte(id))
	res, err = getServicesJson(n, encId)
	if err != nil {
		t.Fatal()
	}
	for i := 0; i < len(res); i++ {
		fmt.Println(res[i])
	}

}
