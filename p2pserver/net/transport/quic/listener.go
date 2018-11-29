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

package quic

import  (
	"crypto/tls"
	"net"
	"strconv"

	quic "github.com/lucas-clemente/quic-go"
	"github.com/ontio/ontology/common/log"
	tsp "github.com/ontio/ontology/p2pserver/net/transport"
)

type listener struct {
	quicListener quic.Listener
}

func newListener(port uint16, tlsConf *tls.Config) (tsp.Listener, error) {
	lnPortStr := ":"+strconv.Itoa(int(port))
	l, err := quic.ListenAddr(lnPortStr, tlsConf, nil)
	if err != nil {
		log.Errorf("[p2p]Can't listen on port %d, err:%s",port, err)
		return nil, err
	}

	return &listener{quicListener: l}, nil
}

func (this *listener) Accept() (tsp.Connection, error) {
	for {
		sess, err := this.quicListener.Accept()
		if err != nil {
			return nil, err
		}
		conn, err := this.setupConn(sess)
		if err != nil {
			sess.CloseWithError(0, err)
			continue
		}
		return conn, nil
	}
}

func (this *listener) setupConn(sess quic.Session) (tsp.Connection, error) {

	return &connection{sess: sess}, nil
}

func (this *listener) Close() error {
	return this.quicListener.Close()
}

func (this *listener) Addr() net.Addr {
	return this.quicListener.Addr()
}

