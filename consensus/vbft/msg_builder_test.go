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

package vbft

import (
	"fmt"
	"testing"

	"github.com/ontio/ontology/account"
)

func constructMsg() *blockProposalMsg {
	acc := account.NewAccount("SHA256withECDSA")
	if acc == nil {
		fmt.Println("GetDefaultAccount error: acc is nil")
		return nil
	}
	msg, err := constructProposalMsgTest(acc)
	if err != nil {
		fmt.Printf("constructProposalMsg failed:%v", err)
		return nil
	}
	return msg
}
func TestSerializeVbftMsg(t *testing.T) {
	msg := constructMsg()
	_, err := SerializeVbftMsg(msg)
	if err != nil {
		t.Errorf("TestSerializeVbftMsg failed :%v", err)
		return
	}
	t.Logf("TestSerializeVbftMsg succ")
}

func TestDeserializeVbftMsg(t *testing.T) {
	msg := constructMsg()
	data, err := SerializeVbftMsg(msg)
	if err != nil {
		t.Errorf("TestSerializeVbftMsg failed :%v", err)
		return
	}
	_, err = DeserializeVbftMsg(data)
	if err != nil {
		t.Errorf("DeserializeVbftMsg failed :%v", err)
		return
	}
	t.Logf("TestDeserializeVbftMsg succ")
}
