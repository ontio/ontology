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
	"io"
	"net"
	"time"

	"github.com/ontio/ontology/p2pserver/common"
	tsp "github.com/ontio/ontology/p2pserver/net/transport"
)

type connection struct {
	net.Conn
	io.Reader
}

type recvStream struct {
	io.Reader
}

func (this* recvStream) CanContinue() bool {

	return false
}

func (this * connection) GetRecvStream() (tsp.RecvStream, error) {

	return  &recvStream{this.Reader}, nil
}

func (this * connection) GetTransportType() byte {

	return 	common.T_TCP
}

func (this * connection) Write(cmdType string, b []byte) (n int, err error) {

	return this.Conn.Write(b)
}

func (this * connection) Close() error {

	return  this.Conn.Close()
}

func (this* connection) LocalAddr() net.Addr {

	return this.Conn.LocalAddr()
}

func (this* connection) RemoteAddr() net.Addr {

	return this.Conn.RemoteAddr()
}
func (this * connection) SetWriteDeadline(t time.Time) error {

	return  this.Conn.SetWriteDeadline(t)
}
