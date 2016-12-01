package transaction

import (
	"GoOnchain/common"
	"GoOnchain/core/contract"
)

type TxOutput struct {
	AssetID common.Uint256
	Value common.Fixed8
	Address contract.Address
}