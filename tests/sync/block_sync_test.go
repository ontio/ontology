
package TestSync

import (
	"github.com/ontio/ontology/tests"
	"testing"
	"github.com/ontio/ontology/tests/common"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/p2pserver"
)

func init() {
	TestConsts.TestRootDir = "../"
}

func Test_RootChainBlockSync(t *testing.T) {

	// . create blockchains for peer1 and peer2 (with same genesis block)
	shardID1 := common.NewShardIDUnchecked(config.DEFAULT_SHARD_ID)
	TestCommon.CreateChain(t, "src", shardID1, 0)
	lgr1 := ledger.GetShardLedger(shardID1)
	ledger.RemoveLedger(shardID1)

	shardID2 := common.NewShardIDUnchecked(config.DEFAULT_SHARD_ID)
	TestCommon.CreateChain(t, "dst", shardID2, 0)
	lgr2 := ledger.GetShardLedger(shardID2)

	// . create peer1, peer2
	peer1 := TestCommon.NewPeer(lgr1)
	peer2 := TestCommon.NewPeer(lgr2)

	peer1.Register()
	peer2.Register()

	// . add 10 blocks to peer1
	for i := 0; i < 10; i++ {
		blk := TestCommon.CreateBlock(t, shardID1, []*types.Transaction{})
		result, err := lgr1.ExecuteBlock(blk)
		if err != nil {
			t.Fatalf("execute blk: %s", err)
		}
		if err := lgr1.SubmitBlock(blk, result); err != nil {
			t.Fatalf("submit block: %s", err)
		}
	}

	// . start block-syncer of peer2
	peer1.Start()
	peer2.Start()

	syncer := p2pserver.NewBlockSyncMgr(shardID2, peer2, lgr2)
	peer2.AddBlockSyncer(shardID2, syncer)
	go syncer.Start()

	// . check block height, block hash, ledger of peer2

}
