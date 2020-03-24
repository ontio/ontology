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

package mock

import (
	"crypto/rand"
	"encoding/binary"
	"testing"

	"net"

	"github.com/ontio/ontology/p2pserver/common"
	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

func genPeerID() common.PeerId {
	b := make([]byte, 8)
	rand.Read(b)
	id := binary.BigEndian.Uint64(b)
	return common.PseudoPeerIdFromUint64(id)
}

func TestNetwork(t *testing.T) {
	a := require.New(t)
	dp := genPeerID()
	lp := genPeerID()

	n := NewNetwork()
	d := n.NewDialer(dp)
	laddr, l := n.NewListener(lp)

	// before allow
	dconn, err := d.Dial(laddr)
	a.Nil(dconn, "connection should be nil")
	a.NotNil(err, "err shuld not be nil")
	a.Contains(err.Error(), "can not be reached", "can not dial to remote address")

	// after allow
	n.AllowConnect(dp, lp)
	dconn, err = d.Dial(laddr)
	a.Nil(err, "should be nil")
	a.Equal(dconn.RemoteAddr().String(), l.Addr().String(), "dialer remote should be the listeners address")

	lconn, err := l.Accept()
	a.Nil(err, "accept should get one conn")
	a.NotNil(lconn, "should be a real conn")
	a.Equal(lconn.RemoteAddr().String(), dconn.LocalAddr().String(), "accept conn remote should be the dialer address")
	a.Equal(lconn.LocalAddr().String(), dconn.RemoteAddr().String(), "remote should match")

	// dial again
	dconn2, err := d.Dial(laddr)
	a.Nil(err, "dial again should be OK")
	a.Equal(dconn.LocalAddr().String(), dconn2.LocalAddr().String(), "dialer source ip should be same")

	lconn2, err := l.Accept()
	a.Nil(err, "accept should get one conn")
	a.NotNil(lconn2, "should be a real conn")
	a.Equal(lconn2.RemoteAddr().String(), dconn2.LocalAddr().String(), "remote should match")
	a.Equal(dconn2.RemoteAddr().String(), lconn2.LocalAddr().String(), "remote should match")

	// Accept after close
	l.Close()
	lconn2, err = l.Accept()
	a.NotNil(err, "closed 'stream' should accept none")
	a.Nil(lconn2, "should be nil")
}

func TestNetIP(t *testing.T) {
	a := require.New(t)
	ip := net.ParseIP("0.0.0.1")
	a.Equal(ip.String(), "0.0.0.1")
	_, err := net.LookupHost("1.0.0.1")
	assert.Nil(t, err)
}
