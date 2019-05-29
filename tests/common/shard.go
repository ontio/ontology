
package TestCommon

import (
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
	"testing"
	"github.com/ontio/ontology/core/chainmgr/xshard"
)

func GetShardStateFromLedger(t *testing.T, lgr *ledger.Ledger, shardID common.ShardID) *shardstates.ShardState {
	state, err := xshard.GetShardState(lgr, shardID)
	if err != nil {
		t.Fatalf(err.Error())
	}
	return state
}
