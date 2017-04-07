package vm

import (
	"math/big"
	"testing"
)

func TestCommon(t *testing.T) {
	i := ToBigInt(big.NewInt(1))
	t.Log("i", i)
}
