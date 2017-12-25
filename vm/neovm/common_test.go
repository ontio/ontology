package neovm

import (
	"math/big"
	"testing"
	"fmt"
)

func TestCommon(t *testing.T) {
	i := ToBigInt(big.NewInt(1))
	t.Log("i", i)

	fmt.Println(ToArrayReverse([]byte{1, 2, 3}))
}

func ToArrayReverse(arr []byte) []byte {
	l := len(arr)
	x := make([]byte, 0)
	for i := l - 1; i >= 0; i-- {
		x = append(x, arr[i])
	}
	return x
}
