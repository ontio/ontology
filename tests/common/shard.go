package TestCommon

import (
	"testing"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/chainmgr/xshard"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
)

func GetShardStateFromLedger(t *testing.T, lgr *ledger.Ledger, shardID common.ShardID) *shardstates.ShardState {
	state, err := xshard.GetShardState(lgr, shardID)
	if err != nil {
		t.Fatalf(err.Error())
	}
	return state
}
