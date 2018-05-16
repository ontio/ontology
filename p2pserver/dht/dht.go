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
	"errors"
	"net"
	"sort"
	"sync"

	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/dht/types"
	"github.com/ontio/ontology/p2pserver/message/msg_pack"
	mt "github.com/ontio/ontology/p2pserver/message/types"
)

type DHT struct {
	version      uint16
	nodeID       types.NodeID
	mu           sync.Mutex
	routingTable *routingTable
	addr         string
	port         uint16
	conn         *net.UDPConn
	recvCh       chan *types.DHTMessage
	stopCh       chan struct{}
}

func NewDHT() *DHT {
	dht := &DHT{}
	dht.init()
	return dht
}

func (this *DHT) init() {
	this.recvCh = make(chan *types.DHTMessage, types.MSG_CACHE)
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
				this.processPacket(pk.From, pk.Payload)
			}
		case <-this.stopCh:
			return
		}
	}

}

func (this *DHT) lookup(targetID types.NodeID) []*types.Node {
	bucket, _ := this.routingTable.locateBucket(targetID)
	node, ret := this.routingTable.isNodeInBucket(targetID, bucket)
	if ret == true {
		log.Infof("targetID %s is in the bucket %d", targetID.String(), bucket)
		return []*types.Node{node}
	}

	visited := make(map[types.NodeID]bool)
	knownNode := make(map[types.NodeID]bool)
	responseCh := make(chan []*types.Node, types.FACTOR)
	pendingQueries := 0

	visited[this.nodeID] = true

	closestNodes := this.routingTable.GetClosestNodes(types.BUCKET_SIZE, targetID)

	if len(closestNodes) == 0 {
		return nil
	}

	for {
		for i := 0; i < len(closestNodes) && pendingQueries < types.FACTOR; i++ {
			node := closestNodes[i]
			if visited[node.ID] == true {
				continue
			}
			visited[node.ID] = true
			pendingQueries++
			go func() {
				ret, _ := this.FindNode(node, targetID)
				responseCh <- ret
			}()
		}

		if pendingQueries == 0 {
			break
		}

		select {
		case entries, ok := <-responseCh:
			if ok {
				for _, n := range entries {
					log.Info("receive new node", n)
					// Todo:
					if knownNode[n.ID] == true {
						continue
					}
					knownNode[n.ID] = true
					idx := sort.Search(len(closestNodes), func(i int) bool {
						for j := range targetID {
							da := closestNodes[i].ID[j] ^ targetID[j]
							db := n.ID[j] ^ targetID[j]
							if da > db {
								return true
							} else if da < db {
								return false
							}
						}
						return false
					})
					if len(closestNodes) < types.BUCKET_SIZE {
						closestNodes = append(closestNodes, n)
					}
					if idx < len(closestNodes) {
						copy(closestNodes[idx+1:], closestNodes[idx:])
						closestNodes[idx] = n
					}
				}
			}
		}

		pendingQueries--
	}
	return closestNodes
}

func (this *DHT) FindNode(remotePeer *types.Node, targetID types.NodeID) ([]*types.Node, error) {
	return nil, nil

}

func (this *DHT) AddNode(remoteNode uint64) {

}

func (this *DHT) Ping(addr *net.UDPAddr) error {
	pingPayload := mt.DHTPingPayload{
		Version:  this.version,
		SrcPort:  this.port,
		DestPort: uint16(addr.Port),
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
		SrcPort:  this.port,
		DestPort: uint16(addr.Port),
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
		pk := &types.DHTMessage{
			From:    from,
			Payload: buf[:nbytes],
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
