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

func (bi *BalanceTxInput) Deserialize(r io.Reader) error  {
	err := bi.AssetID.Deserialize(r)
	if err != nil {return err}

	err = bi.Value.Deserialize(r)
	if err != nil {return err}

	err = bi.ProgramHash.Deserialize(r)
	if err != nil {return err}

	return nil
}