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
	"fmt"
	"testing"

	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/common"
)

func init() {
	log.InitLog(log.InfoLog)
	fmt.Println("Start test the netserver...")

}
func TestNewP2PServer(t *testing.T) {
	log.Init(log.Stdout)
	fmt.Println("Start test new p2pserver...")

	p2p := NewServer()

	if p2p.GetVersion() != common.PROTOCOL_VERSION {
		t.Error("TestNewP2PServer p2p version error", p2p.GetVersion())
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
}
