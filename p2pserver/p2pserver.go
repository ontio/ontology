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

	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/connect_controller"
	"github.com/ontio/ontology/p2pserver/net/netserver"
	p2p "github.com/ontio/ontology/p2pserver/net/protocol"
	"github.com/ontio/ontology/p2pserver/protocols"
	"github.com/ontio/ontology/p2pserver/protocols/utils"
)

//P2PServer control all network activities
type P2PServer struct {
	network *netserver.NetServer
	db      *ledger.Ledger
}

//NewServer return a new p2pserver according to the pubkey
func NewServer(acct *account.Account) (*P2PServer, error) {
	db := ledger.DefLedger
	var rsv []string
	var recRsv []string
	conf := config.DefConfig.P2PNode
	if conf.ReservedPeersOnly && conf.ReservedCfg != nil {
		rsv = conf.ReservedCfg.ReservedPeers
	}
	if conf.ReservedCfg != nil {
		recRsv = conf.ReservedCfg.ReservedPeers
	}

	staticFilter := connect_controller.NewStaticReserveFilter(rsv)
	protocol := protocols.NewMsgHandler(acct, connect_controller.NewStaticReserveFilter(recRsv), db, common.NewGlobalLoggerWrapper())
	reserved := protocol.GetReservedAddrFilter(len(rsv) != 0)
	reservedPeers := p2p.CombineAddrFilter(staticFilter, reserved)
	n, err := netserver.NewNetServer(protocol, conf, reservedPeers)
	if err != nil {
		return nil, err
	}

	p := &P2PServer{
		db:      db,
		network: n,
	}

	return p, nil
}

//Start create all services
func (self *P2PServer) Start() error {
	return self.network.Start()
}

//Stop halt all service by send signal to channels
func (self *P2PServer) Stop() {
	self.network.Stop()
}

// GetNetwork returns the low level netserver
func (self *P2PServer) GetNetwork() p2p.P2P {
	return self.network
}

//WaitForPeersStart check whether enough peer linked in loop
func (self *P2PServer) WaitForPeersStart() {
	periodTime := config.DEFAULT_GEN_BLOCK_TIME / common.UPDATE_RATE_PER_BLOCK
	for {
		log.Info("[p2p]Wait for minimum connection...")
		if self.reachMinConnection() {
			break
		}

		<-time.After(time.Second * (time.Duration(periodTime)))
	}
}

//reachMinConnection return whether net layer have enough link under different config
func (self *P2PServer) reachMinConnection() bool {
	if !config.DefConfig.Consensus.EnableConsensus {
		//just sync
		return true
	}
	consensusType := strings.ToLower(config.DefConfig.Genesis.ConsensusType)
	if consensusType == "" {
		consensusType = "dbft"
	}
	var minCount uint32 = config.DBFT_MIN_NODE_NUM
	switch consensusType {
	case "dbft":
	case "solo":
		minCount = config.SOLO_MIN_NODE_NUM
	case "vbft":
		minCount = self.getVbftGovNodeCount()
	}
	return self.network.GetConnectionCnt()+1 >= minCount
}

func (self *P2PServer) getVbftGovNodeCount() uint32 {
	view, err := utils.GetGovernanceView(self.db)
	if err != nil {
		return config.VBFT_MIN_NODE_NUM
	}
	_, count, err := utils.GetPeersConfig(self.db, view.View)
	if err != nil {
		return config.VBFT_MIN_NODE_NUM
	}

	return count - count/3
}
