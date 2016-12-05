package transaction

import (
	"GoOnchain/common"
)

type TxOutput struct {
	AssetID common.Uint256
	Value common.Fixed8
	ProgramHash common.Uint160
}