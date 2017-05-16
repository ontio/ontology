package program

import (
	"fmt"
	"math/big"
	"testing"
)

func TestProgramBuilder_PushData(t *testing.T) {
	for i := int64(-20); i <= 26; i++ {
		pb := NewProgramBuilder()
		testx := big.NewInt(i)
		pb.PushNumber(testx)
		fmt.Printf("%d=%#v\n", i, pb.ToArray())
	}

}
