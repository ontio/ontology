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
	"strings"
	"time"

	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/net/netserver"
	p2pnet "github.com/ontio/ontology/p2pserver/net/protocol"
	"github.com/ontio/ontology/p2pserver/protocols"
)

//P2PServer control all network activities
type P2PServer struct {
	network *netserver.NetServer
}

//NewServer return a new p2pserver according to the pubkey
func NewServer() (*P2PServer, error) {
	ld := ledger.DefLedger

	protocol := protocols.NewMsgHandler(ld)
	n, err := netserver.NewNetServer(protocol, config.DefConfig.P2PNode)
	if err != nil {
		return nil, err
	}

	p := &P2PServer{
		network: n,
	}

	return p, nil
}

//Start create all services
func (this *P2PServer) Start() error {
	return this.network.Start()
}

//Stop halt all service by send signal to channels
func (this *P2PServer) Stop() {
	this.network.Stop()
}

// GetNetwork returns the low level netserver
func (this *P2PServer) GetNetwork() p2pnet.P2P {
	return this.network
}

//WaitForPeersStart check whether enough peer linked in loop
func (this *P2PServer) WaitForPeersStart() {
	periodTime := config.DEFAULT_GEN_BLOCK_TIME / common.UPDATE_RATE_PER_BLOCK
	for {
		log.Info("[p2p]Wait for minimum connection...")
		if this.reachMinConnection() {
			break
		}

		<-time.After(time.Second * (time.Duration(periodTime)))
	}
}

//reachMinConnection return whether net layer have enough link under different config
func (this *P2PServer) reachMinConnection() bool {
	if !config.DefConfig.Consensus.EnableConsensus {
		//just sync
		return true
	}
	consensusType := strings.ToLower(config.DefConfig.Genesis.ConsensusType)
	if consensusType == "" {
		consensusType = "dbft"
	}
	minCount := config.DBFT_MIN_NODE_NUM
	switch consensusType {
	case "dbft":
	case "solo":
		minCount = config.SOLO_MIN_NODE_NUM
	case "vbft":
		minCount = config.VBFT_MIN_NODE_NUM

	}
	return int(this.network.GetConnectionCnt())+1 >= minCount
}
