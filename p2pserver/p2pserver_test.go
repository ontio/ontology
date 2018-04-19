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

package p2pserver

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"
	"time"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/common"
)

var key keypair.PublicKey
var acct *account.Account

func init() {
	log.Init(log.Stdout)
	fmt.Println("Start test the netserver...")
	acct = account.NewAccount("SHA256withECDSA")
	key = acct.PubKey()

}
func TestNewP2PServer(t *testing.T) {
	log.Init(log.Stdout)
	fmt.Println("Start test new p2pserver...")

	p2p, err := NewServer(acct)
	if err != nil {
		t.Fatalf("TestP2PActorServer: p2pserver NewServer error %s", err)
	}
	//false because the ledger actor not running
	p2p.Start(false)
	defer p2p.Stop()

	if p2p.GetVersion() != common.PROTOCOL_VERSION {
		t.Error("TestNewP2PServer p2p version error", p2p.GetVersion())
	}

	var id uint64
	k := keypair.SerializePublicKey(key)
	err = binary.Read(bytes.NewBuffer(k[:8]), binary.LittleEndian, &(id))
	if err != nil {
		t.Error(err)
	}

	if p2p.GetID() != id {
		t.Error("TestNewP2PServer p2p id error")
	}
	if p2p.GetVersion() != common.PROTOCOL_VERSION {
		t.Error("TestNewP2PServer p2p version error")
	}
	sync, cons := p2p.GetPort()
	if sync != 20338 {
		t.Error("TestNewP2PServer sync port error")
	}

	if cons != 20339 {
		t.Error("TestNewP2PServer consensus port error")
	}
	go p2p.WaitForSyncBlkFinish()
	<-time.After(time.Second * common.KEEPALIVE_TIMEOUT)
}
