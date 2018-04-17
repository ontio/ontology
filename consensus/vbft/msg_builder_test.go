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
	"os"
	"testing"

	"github.com/ontio/ontology/account"
)

func constructMsg() *blockProposalMsg {
	passwd := string("passwordtest")
	acct := account.Open(account.WALLET_FILENAME, []byte(passwd))
	acc := acct.GetDefaultAccount()
	if acc == nil {
		fmt.Println("GetDefaultAccount error: account is nil")
		os.Exit(1)
	}
	msg, err := constructProposalMsg(acc)
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
