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

package dht

import (
	//"fmt"
	"net"
	"sync"

	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/dht/types"
	"github.com/ontio/ontology/p2pserver/message/msg_pack"
	mt "github.com/ontio/ontology/p2pserver/message/types"
)

type DHT struct {
	version      uint32
	nodeID       types.NodeID
	mu           sync.Mutex
	routingTable *routingTable
	addr         string
	port         uint16
	conn         *net.UDPConn
	recvCh       chan *DHTMessage
	stopCh       chan struct{}
}

func (this *DHT) init() {
	this.recvCh = make(chan *DHTMessage, MSG_CACHE)
	this.stopCh = make(chan struct{})
	this.routingTable.init(this.nodeID)
}

func (this *DHT) Start() {

}

func (this *DHT) Stop() {

}

func (this *DHT) Loop() {
	for {
		select {
		case pk, ok := <-this.recvCh:
			if ok {
				this.processPacket(pk.from, pk.payload)
			}
		case <-this.stopCh:
			return
		}
	}

}

func (this *DHT) lookup(targetID uint64) {

}

func (this *DHT) FindNode(remotePeer, targetID uint64) {

}

func (this *DHT) AddNode(remoteNode uint64) {

}

func (this *DHT) Ping(addr *net.UDPAddr) error {
	pingPayload := mt.DHTPingPayload{
		Version:  this.version,
		srcPort:  this.port,
		DestPort: addr.Port,
	}

	ip := net.ParseIP(this.addr).To16()
	if ip == nil {
		log.Error("Parse IP address error\n", this.addr)
		return errors.New("Parse IP address error")
	}
	copy(pingPayload.SrcAddr[:], ip[:16])

	ip = addr.IP.To4()
	if ip == nil {
		ip = addr.IP.To16()
	}
	copy(pingPayload.DestAddr[:], ip[:16])

	copy(pingPayload.FromID[:], this.nodeID[:])

	pingPacket, err := msgpack.NewDHTPing(pingPayload)
	if err != nil {
		log.Error("failed to new dht ping packet", err)
		return err
	}
	this.send(addr, pingPacket)
	return nil
}

func (this *DHT) Pong(addr *net.UDPAddr) error {
	PongPayload := mt.DHTPongPayload{
		Version:  this.version,
		srcPort:  this.port,
		DestPort: addr.Port,
	}

	ip := net.ParseIP(this.addr).To16()
	if ip == nil {
		log.Error("Parse IP address error\n", this.addr)
		return errors.New("Parse IP address error")
	}
	copy(PongPayload.SrcAddr[:], ip[:16])

	ip = addr.IP.To4()
	if ip == nil {
		ip = addr.IP.To16()
	}
	copy(PongPayload.DestAddr[:], ip[:16])

	copy(PongPayload.FromID[:], this.nodeID[:])

	pongPacket, err := msgpack.NewDHTPong(PongPayload)
	if err != nil {
		log.Error("failed to new dht pong packet", err)
		return err
	}
	this.send(addr, pongPacket)
	return nil
}

func (this *DHT) processPacket(from *net.UDPAddr, packet []byte) {
	// Todo: add processPacket implementation
	msgType, err := mt.MsgType(packet)
	if err != nil {
		log.Info("failed to get msg type")
		return
	}

	log.Trace("Recv UDP msg", msgType)
}

func (this *DHT) recvUDPMsg() {
	defer this.conn.Close()
	buf := make([]byte, common.MAX_BUF_LEN)
	for {
		nbytes, from, err := this.conn.ReadFromUDP(buf)
		if err != nil {
			log.Error("ReadFromUDP error:", err)
			return
		}
		// Todo:
		pk := &DHTMessage{
			from:    from,
			payload: buf[:nbytes],
		}
		this.recvCh <- pk
	}
}

func (this *DHT) ListenUDP(laddr string) error {
	addr, err := net.ResolveUDPAddr("udp", laddr)
	if err != nil {
		log.Error("failed to resolve udp address", laddr, "error: ", err)
		return err
	}
	this.conn, err = net.ListenUDP("udp", addr)
	if err != nil {
		log.Error("failed to listen udp on", addr, "error: ", err)
		return err
	}

	go this.recvUDPMsg()
	return nil
}

func (this *DHT) send(addr *net.UDPAddr, msg []byte) error {
	_, err := this.conn.WriteToUDP(msg, addr)
	if err != nil {
		log.Error("failed to send msg", err)
		return err
	}
	return nil
}
