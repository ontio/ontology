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
	"net"
	"strconv"

	"github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/connect_controller"
	"github.com/ontio/ontology/p2pserver/net/netserver"
	p2p "github.com/ontio/ontology/p2pserver/net/protocol"
	"github.com/ontio/ontology/p2pserver/peer"
)

type Network interface {
	// NewListener will gen random ip to listen
	NewListener(id common.PeerId) (string, net.Listener)
	// addr: ip:port
	NewListenerWithAddr(id common.PeerId, addr string) net.Listener

	// NewDialer will gen random source IP
	NewDialer(id common.PeerId) connect_controller.Dialer
	NewDialerWithHost(id common.PeerId, host string) connect_controller.Dialer
	AllowConnect(id1, id2 common.PeerId)
	DeliverRate(percent uint)
}

func NewNode(keyId *common.PeerKeyId, listenAddr string, localInfo *peer.PeerInfo, proto p2p.Protocol, nw Network,
	reservedPeers []string, reserveAddrFilter p2p.AddressFilter, logger common.Logger) *netserver.NetServer {
	var listener net.Listener
	if listenAddr != "" {
		listener = nw.NewListenerWithAddr(keyId.Id, listenAddr)
	} else {
		listenAddr, listener = nw.NewListener(keyId.Id)
	}
	host, port, _ := net.SplitHostPort(listenAddr)
	dialer := nw.NewDialerWithHost(keyId.Id, host)
	localInfo.Addr = listenAddr
	iport, _ := strconv.Atoi(port)
	localInfo.Port = uint16(iport)
	opt := connect_controller.NewConnCtrlOption().MaxInBoundPerIp(10).
		MaxInBound(20).MaxOutBound(20).WithDialer(dialer).ReservedOnly(reservedPeers)
	opt.ReservedPeers = p2p.CombineAddrFilter(opt.ReservedPeers, reserveAddrFilter)
	return netserver.NewCustomNetServer(keyId, localInfo, proto, listener, opt, logger)
}
