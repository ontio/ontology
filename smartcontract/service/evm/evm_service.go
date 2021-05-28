package evm

import (
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/context"
	"github.com/ontio/ontology/smartcontract/storage"
)

type EvmService struct {
	Code       []byte
	CacheDB    *storage.CacheDB
	Tx         *types.Transaction
	Time       uint32
	Height     uint32
	BlockHash  common.Uint256
	ContextRef context.ContextRef
}

func (this *EvmService) Invoke() (interface{}, error) {
	//config := params.MainnetChainConfig //todo use config based on network

	//_, receipt, err := ApplyTransaction(config, store, statedb, header, tx, &usedGas, utils.GovernanceContractAddress, evm.Config{})

	//ApplyTransaction(config,)

	return nil, nil
}
