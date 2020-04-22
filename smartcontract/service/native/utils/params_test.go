package utils

import (
	"testing"

	"github.com/ontio/ontology/common"
	"github.com/stretchr/testify/assert"
)

func TestIsNativeContract(t *testing.T) {
	address := []common.Address{OntContractAddress, OngContractAddress, OntIDContractAddress,
		ParamContractAddress, AuthContractAddress, GovernanceContractAddress,
		HeaderSyncContractAddress, CrossChainContractAddress, LockProxyContractAddress}
	for _, addr := range address {
		assert.True(t, IsNativeContract(addr))
	}

	assert.False(t, IsNativeContract(common.ADDRESS_EMPTY))
}
