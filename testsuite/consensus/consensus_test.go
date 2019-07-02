/*
 * Copyright (C) 2019 The ontology Authors
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

package TestConsensus

import (
	"fmt"
	"testing"
	"time"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/consensus"
	"github.com/ontio/ontology/consensus/vbft"
	"github.com/ontio/ontology/core/chainmgr"
	"github.com/ontio/ontology/core/chainmgr/xshard"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/events"
	"github.com/ontio/ontology/testsuite"
	"github.com/ontio/ontology/testsuite/common"
)

func init() {
	TestConsts.TestRootDir = "../"
}

func StartMockerConsensus(t *testing.T, shardID common.ShardID, name string, srcLgr *ledger.Ledger) consensus.ConsensusService {
	shardName := chainmgr.GetShardName(shardID)

	acc := TestCommon.GetAccount(shardName + "_" + name)
	if acc == nil {
		t.Fatalf("failed to get user account of shard %s", shardName)
	}

	lgr := TestCommon.CloneChain(t, name, srcLgr)
	ledger.RemoveLedger(shardID)

	txPool := TestCommon.NewTxnPool(t, name, shardID)
	peer := TestCommon.NewPeer(lgr)
	peer.Register()
	p2pActor := TestCommon.NewP2PActor(t, name, peer)

	txPool.Start(t)
	p2pActor.Start(t)
	peer.Start()

	service, err := consensus.NewConsensusService(consensus.CONSENSUS_VBFT, shardID, acc, txPool.GetPID(t), lgr, p2pActor.GetPID(t))
	if err != nil {
		t.Fatalf("start consensus: %s", err)
	}
	peer.SetConsensusPid(t, service.GetPID())
	return service
}

func Test_NewConsensusService_7nodes(t *testing.T) {
	shardID := common.NewShardIDUnchecked(config.DEFAULT_SHARD_ID)
	events.Init()
	xshard.InitCrossShardPool(shardID, 100)

	// . create template chain
	TestCommon.CreateChain(t, "src", shardID, 0)
	lgr := ledger.GetShardLedger(shardID)
	ledger.RemoveLedger(shardID)

	services := make([]consensus.ConsensusService, 0)
	for i := 0; i < 7; i++ {
		name := fmt.Sprintf("peerOwner%d", i)
		s := StartMockerConsensus(t, shardID, name, lgr)
		services = append(services, s)
	}

	for _, s := range services {
		s.Start()
	}

	time.Sleep(60 * time.Second)

	for i, s := range services {
		v := s.(*vbft.Server)
		if v == nil {
			t.Fatalf("failed cast consensus service to vbft")
		}
		if v.GetCurrentBlockNo() == 0 {
			t.Fatalf("vbft %d, current block height %d", i, v.GetCurrentBlockNo())
		}
		log.Infof("vbft %d, height %d", i, v.GetCurrentBlockNo())
	}
}

func Test_MakeProposal(t *testing.T) {

}
