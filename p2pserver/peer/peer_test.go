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

package peer

import (
	"testing"
	"time"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common/log"
)

var p *Peer
var key keypair.PublicKey

func init() {
	log.Init(log.Stdout)
	p = NewPeer()
	p.base.version = 1
	p.base.services = 1
	p.base.syncPort = 10338
	p.base.consPort = 10339
	p.base.relay = true
	p.base.height = 123355
	p.base.id = 29357734007
	acct := account.NewAccount("SHA256withECDSA")
	key = acct.PubKey()
	p.SetBookKeeperAddr(key)
}
func TestGetPeerComInfo(t *testing.T) {
	p.DumpInfo()
	if p.base.GetVersion() != 1 {
		t.Errorf("PeerCom GetVersion error")
	} else {
		p.base.SetVersion(2)
		if p.base.GetVersion() != 2 {
			t.Errorf("PeerCom SetVersion error")
		}
	}

	if p.base.GetServices() != 1 {
		t.Errorf("PeerCom GetServices error")
	} else {
		p.base.SetServices(2)
		if p.base.GetServices() != 2 {
			t.Errorf("PeerCom SetServices error")
		}
	}

	if p.base.GetSyncPort() != 10338 {
		t.Errorf("PeerCom GetSyncPort error")
	} else {
		p.base.SetSyncPort(20338)
		if p.base.GetSyncPort() != 20338 {
			t.Errorf("PeerCom SetSyncPort error")
		}
	}

	if p.base.GetConsPort() != 10339 {
		t.Errorf("PeerCom GetConsPort error")
	} else {
		p.base.SetConsPort(20339)
		if p.base.GetConsPort() != 20339 {
			t.Errorf("PeerCom SetConsPort error")
		}
	}

	if p.base.GetRelay() != true {
		t.Errorf("PeerCom GetRelay error")
	} else {
		p.base.SetRelay(false)
		if p.base.GetRelay() != false {
			t.Errorf("PeerCom SetRelay error")
		}
	}

	if p.base.GetHeight() != 123355 {
		t.Errorf("PeerCom GetHeight error")
	} else {
		p.base.SetHeight(234343497)
		if p.base.GetHeight() != 234343497 {
			t.Errorf("PeerCom SetHeight error")
		}
	}

	if p.base.GetID() != 29357734007 {
		t.Errorf("PeerCom GetID error")
	} else {
		p.base.SetID(1224322422)
		if p.base.GetID() != 1224322422 {
			t.Errorf("PeerCom SetID error")
		}
	}
	if p.base.GetPubKey() != key {
		t.Errorf("PeerCom GetPubKey error")
	}
}

func TestUpdatePeer(t *testing.T) {
	p.UpdateInfo(time.Now(), 3, 3, 30334, 30335, 0x7533345, 0, 7322222)
	p.SetConsState(2)
	p.SetSyncState(3)
	p.SetHttpInfoState(true)
	p.SyncLink.SetAddr("127.0.0.1:20338")
	p.DumpInfo()

}
