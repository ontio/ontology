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
package creator

import (
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/common"
	"testing"
)

func init() {
	log.Init(log.Stdout)
}

func TestFactory(t *testing.T) {
	tcpTransport1, _ := GetTransportFactory().GetTransport(10)
	if tcpTransport1 != nil {
		t.Error("tcpTransport1 should be nil")
	}

	tcpTransport2, _ := GetTransportFactory().GetTransport(common.T_TCP)
	if tcpTransport2 == nil {
		t.Error("tcpTransport2 shouldnot be nil")
	}
}
