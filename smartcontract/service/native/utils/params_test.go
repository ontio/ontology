package utils

import "testing"

func TestContractToBase58(t *testing.T) {
	t.Logf("OntContractAddress is %s", OntContractAddress.ToBase58())
	t.Logf("OngContractAddress is %s", OngContractAddress.ToBase58())
	t.Logf("OntIDContractAddress is %s", OntIDContractAddress.ToBase58())
	t.Logf("ParamContractAddress is %s", ParamContractAddress.ToBase58())
	t.Logf("AuthContractAddress is %s", AuthContractAddress.ToBase58())
	t.Logf("GovernanceContractAddress is %s", GovernanceContractAddress.ToBase58())
	t.Logf("ShardMgmtContractAddress is %s", ShardMgmtContractAddress.ToBase58())
	t.Logf("ShardGasMgmtContractAddress is %s", ShardGasMgmtContractAddress.ToBase58())
	t.Logf("ShardSysMsgContractAddress is %s", ShardSysMsgContractAddress.ToBase58())
	t.Logf("ShardCCMCAddress is %s", ShardCCMCAddress.ToBase58())
	t.Logf("ShardPingAddress is %s", ShardPingAddress.ToBase58())
	t.Logf("ShardHotelAddress is %s", ShardHotelAddress.ToBase58())
	t.Logf("ShardStakeAddress is %s", ShardStakeAddress.ToBase58())
	t.Logf("ShardAssetAddress is %s", ShardAssetAddress.ToBase58())
}
