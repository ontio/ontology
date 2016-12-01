package transaction

import (
	"GoOnchain/common"
	"GoOnchain/core/contract"
)


type BalanceTxInput struct {
	AssetID common.Uint256
	Value common.Fixed8
	Address contract.Address
}