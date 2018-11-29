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

package tcp

import (
	"bufio"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/common"
	"net"

	tsp "github.com/ontio/ontology/p2pserver/net/transport"

)

type listener struct {
	net.Listener
}

func newListen(port uint16) (tsp.Listener, error) {

	nl, err := createListener(port)
	if err != nil {
		log.Errorf("[p2p]Listen Error on %d of network TCP, err:%s", port,  err)
		return nil, err
	}

	return &listener{nl, }, nil
}

func (this *listener) Accept() (tsp.Connection, error){
	conn, err := this.Listener.Accept()
	if err != nil {
		log.Errorf("")
		return nil, err
	}

	reader := bufio.NewReaderSize(conn, common.MAX_BUF_LEN)

	return  &connection{conn, reader}, nil
}

func (this *listener) Close() error {
	return this.Listener.Close()
}

func (this *listener) Addr() net.Addr {
	return this.Listener.Addr()
}
