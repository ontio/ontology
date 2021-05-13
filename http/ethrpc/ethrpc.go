package ethrpc

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ontio/ontology/smartcontract/context"
	"math/big"
)

type EthereumAPI struct {
}

func (api *EthereumAPI) ChainId() hexutil.Uint64 {
	return hexutil.Uint64(5851)
}

func (api *EthereumAPI)BlockNumber() hexutil.Uint64 {
	return hexutil.Uint64(1000)
}

func (s *EthereumAPI) GetBalance(ctx context.Context, address common.Address, blockNrOrHash rpc.BlockNumberOrHash) (*hexutil.Big, error) {
	return (*hexutil.Big)(big.NewInt(10000000)), nil
}
