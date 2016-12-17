package transaction

import (
	"GoOnchain/common"
	"io"
)


type BalanceTxInput struct {
	AssetID common.Uint256
	Value common.Fixed64
	ProgramHash common.Uint160
}

func (bi *BalanceTxInput) Serialize(w io.Writer)  {
	bi.AssetID.Serialize(w)
	bi.Value.Serialize(w)
	bi.ProgramHash.Serialize(w)
}