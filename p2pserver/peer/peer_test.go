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
)

func initTestPeer() *Peer {
	p := NewPeer()
	p.base.version = 1
	p.base.services = 1
	p.base.port = 10338
	p.base.relay = true
	p.base.height = 123355
	p.base.id = 29357734007

	return p
}

func TestGetPeerComInfo(t *testing.T) {
	p := initTestPeer()
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

	if p.base.GetPort() != 10338 {
		t.Errorf("PeerCom GetPort error")
	} else {
		p.base.SetPort(20338)
		if p.base.GetPort() != 20338 {
			t.Errorf("PeerCom SetPort error")
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
}

func TestUpdatePeer(t *testing.T) {
	p := initTestPeer()
	p.UpdateInfo(time.Now(), 3, 3, 30334, 0x7533345, 0, 7322222, "1.5.2")
	p.SetState(3)
	p.SetHttpInfoState(true)
	p.Link.SetAddr("127.0.0.1:20338")
	p.DumpInfo()

}
