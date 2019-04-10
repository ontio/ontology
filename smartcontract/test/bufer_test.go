package test

import (
	"testing"
	"fmt"
	"github.com/ontio/ontology/common"
)

func TestBufer(t *testing.T) {
	bf := common.NewZeroCopySink(nil)

	bf.WriteUint64(uint64(22))
	fmt.Println("buf:", bf.Bytes())
}
