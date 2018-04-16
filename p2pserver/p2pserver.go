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
	"errors"

	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/p2pserver/common"

	p2pnet "github.com/ontio/ontology/p2pserver/net/protocol"
	"github.com/ontio/ontology/p2pserver/peer"
)

//P2PServer control all network activities
type P2PServer struct {
	network p2pnet.P2P
	isSync  bool
}

//NewServer return a new p2pserver according to the pubkey
func NewServer(acc *account.Account) (*P2PServer, error) {

	return nil, nil
}

//GetConnectionCnt return the established connect count
func (this *P2PServer) GetConnectionCnt() uint32 {
	return this.network.GetConnectionCnt()
}

//Start create all services
func (this *P2PServer) Start(isSync bool) error {

	return nil
}

//Stop halt all service by send signal to channels
func (this *P2PServer) Stop() error {
	return nil
}

// GetNetWork returns the low level netserver
func (this *P2PServer) GetNetWork() p2pnet.P2P {
	return this.network
}

//IsSyncing return whether p2p is syncing
func (this *P2PServer) IsSyncing() bool {
	return this.isSync
}

//GetPort return two network port
func (this *P2PServer) GetPort() (uint16, uint16) {
	return 0, 0
}

//GetVersion return self version
func (this *P2PServer) GetVersion() uint32 {
	return 0
}

//GetNeighborAddrs return all nbr`s address
func (this *P2PServer) GetNeighborAddrs() ([]common.PeerAddr, uint64) {
	return nil, 0
}

//Xmit called by other module to broadcast msg
func (this *P2PServer) Xmit(message interface{}) error {

	return nil
}

//Send tranfer buffer to peer
func (this *P2PServer) Send(p *peer.Peer, buf []byte) error {

	return errors.New("send to a not ESTABLISH peer")
}

func (this *P2PServer) GetID() uint64 {
	return 0
}

// Todo: remove it if no use
func (this *P2PServer) GetConnectionState() uint32 {
	return 0
}

//GetTime return lastet contact time
func (this *P2PServer) GetTime() int64 {
	return 0
}
