package TestCommon

import (
	"testing"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/chainmgr"
)

func TestCreateAccount(t *testing.T) {
	shardID := common.NewShardIDUnchecked(10)
	shardName := chainmgr.GetShardName(shardID)
	user := shardName + "_adminOntID"

	CreateAccount(t, user)

	acc := GetAccount(user)
	if acc == nil {
		t.Fatalf("failed to get created account")
	}
}
