package evm

import (
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/store"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/context"
	"github.com/ontio/ontology/smartcontract/service/native/ong"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/ontio/ontology/smartcontract/storage"
	"github.com/ontio/ontology/vm/evm"
	"github.com/ontio/ontology/vm/evm/params"

	ethcom "github.com/ethereum/go-ethereum/common"
)

type EvmService struct {
	Store      store.LedgerStore
	Code       []byte
	CacheDB    *storage.CacheDB
	Tx         *types.Transaction
	Time       uint32
	Height     uint32
	BlockHash  common.Uint256
	ContextRef context.ContextRef
}

func (this *EvmService) Invoke() (interface{}, error) {
	config := params.MainnetChainConfig //todo use config based on network
	txIndex := 0
	txhash := this.Tx.Hash()
	thash := ethcom.BytesToHash(txhash[:])
	statedb := storage.NewStateDB(this.CacheDB, thash, ethcom.BytesToHash(this.BlockHash[:]), int(txIndex), ong.OngBalanceHandle{})
	usedGas := uint64(0)

	//_, receipt, err := ApplyTransaction(config, store, statedb, header, tx, &usedGas, utils.GovernanceContractAddress, evm.Config{})
	block, err := this.Store.GetBlockByHash(this.BlockHash)
	if err != nil {
		return nil, err
	}
	eiptx, err := this.Tx.GetEIP155Tx()
	if err != nil {
		return nil, err
	}
	res, _, err := ApplyTransaction(config, this.Store, statedb, block.Header, eiptx, &usedGas, utils.GovernanceContractAddress, evm.Config{})
	this.ContextRef.CheckUseGas(usedGas)

	return res, err
}
