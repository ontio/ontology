
package xshard

import (
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
	"github.com/ontio/ontology/smartcontract/service/native/shard_stake"
)

func GetShardTxsByParentHeight(start, end uint32) map[types.ShardID][]*types.Transaction {
	return nil
}

func GetShardView(lgr *ledger.Ledger, shardID types.ShardID) (*utils.ChangeView, error) {
	return nil, nil
}

func GetShardState(lgr *ledger.Ledger, shardID types.ShardID) (*shardstates.ShardState, error) {
	return nil, nil
}

func GetShardPeerStakeInfo(lgr *ledger.Ledger, shardID types.ShardID, shardView uint32) (map[string]*shard_stake.PeerViewInfo, error) {
	return nil, nil
}
