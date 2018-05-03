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
	"fmt"
	"net"
	"sync"
)

type DHT struct {
	nodeID       uint64
	mu           sync.Mutex
	routingTable *routingTable
	conn         *net.UDPConn
	recvCh       chan *DHTMessage
	stopCh       chan struct{}
}

func (this *DHT) init() {

}

func (this *DHT) Start() {

}

func (this *DHT) Stop() {

}

func (this *DHT) Loop() {

}

func (this *DHT) lookup(targetID uint64) {

}

func (this *DHT) FindNode(remotePeer, targetID uint64) {

}

func (this *DHT) AddNode(remoteNode uint64) {

}

func (this *DHT) Ping(addr string) {

}

func (this *DHT) Pong(addr string) {

}

func (this *DHT) processPacket(packet []byte) {

}

func (this *DHT) ListenUDP() error {
	return nil
}

func (this *DHT) send(addr *net.UDPAddr, msg []byte) error {
	return nil
}
