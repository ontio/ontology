package TestSync

import (
	"testing"
	"time"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/p2pserver"
	"github.com/ontio/ontology/tests"
	"github.com/ontio/ontology/tests/common"
)

func init() {
	TestConsts.TestRootDir = "../"
}

func Test_RootChainBlockSync(t *testing.T) {

	// . create blockchains for peer1 and peer2 (with same genesis block)
	shardID := common.NewShardIDUnchecked(config.DEFAULT_SHARD_ID)
	TestCommon.CreateChain(t, "src", shardID, 0)
	lgr1 := ledger.GetShardLedger(shardID)
	ledger.RemoveLedger(shardID)

	lgr2 := TestCommon.CloneChain(t, "dst", lgr1)

	// . add 10 blocks to peer1
	for i := 0; i < 10; i++ {
		blk := TestCommon.CreateBlock(t, lgr1, []*types.Transaction{})
		result, err := lgr1.ExecuteBlock(blk)
		if err != nil {
			t.Fatalf("execute blk: %s", err)
		}
		if err := lgr1.SubmitBlock(blk, result); err != nil {
			t.Fatalf("submit block: %s", err)
		}
		log.Infof("src lgr height: %d", lgr1.GetCurrentBlockHeight())
	}

	// . create peer1, peer2
	peer1 := TestCommon.NewPeer(lgr1)
	peer2 := TestCommon.NewPeer(lgr2)

	syncer := p2pserver.NewBlockSyncMgr(shardID, peer2, lgr2)
	peer2.AddBlockSyncer(shardID, syncer)
	go syncer.Start()

	peer2.Register()
	peer1.Register()

	// . start block-syncer of peer2
	peer1.Start()
	peer2.Start()

	// . check block height, block hash, ledger of peer2
	for i := 0; i < 5; i++ {
		time.Sleep(time.Second * 5)
		log.Infof("syncer ledger height: %d vs %d \n", lgr1.GetCurrentBlockHeight(), lgr2.GetCurrentBlockHeight())
	}

	if lgr1.GetCurrentBlockHeight() != lgr2.GetCurrentBlockHeight() {
		t.Fatalf("failed to sync %d vs %d", lgr1.GetCurrentBlockHeight(), lgr2.GetCurrentBlockHeight())
	}
}
