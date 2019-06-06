package TestContracts

import (
	"fmt"
	"testing"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/ontio/ontology/tests"
	"github.com/ontio/ontology/tests/common"
)

func init() {
	TestConsts.TestRootDir = "../"
}

func Test_ShardMgmtInit(t *testing.T) {

	// 1. create root chain
	shardID := common.NewShardIDUnchecked(config.DEFAULT_SHARD_ID)
	TestCommon.CreateChain(t, "test", shardID, 0)

	// 2. build shard-mgmt init tx

	tx := TestCommon.CreateAdminTx(t, shardID, utils.ShardMgmtContractAddress, shardmgmt.INIT_NAME, nil)

	// 3. create new block
	blk := TestCommon.CreateBlock(t, ledger.GetShardLedger(shardID), []*types.Transaction{tx})

	// 4. add block
	TestCommon.ExecBlock(t, shardID, blk)
	TestCommon.SubmitBlock(t, shardID, blk)

	// 5. query db
	state := TestCommon.GetShardStateFromLedger(t, ledger.GetShardLedger(shardID), shardID)
	fmt.Printf("%v", state)
}
