package neovm

import (
	"bytes"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/vm/neovm/types"
	"github.com/stretchr/testify/assert"
	"math"
	"math/big"
	"testing"
)

func TestBuildParamToNative(t *testing.T) {
	inte := types.NewInteger(new(big.Int).SetUint64(math.MaxUint64))
	boo := types.NewBoolean(false)
	bs := types.NewByteArray([]byte("hello"))
	s := make([]types.StackItems, 0)
	s = append(s, inte)
	s = append(s, boo)
	s = append(s, bs)
	stru := types.NewStruct(s)
	arr := types.NewArray(nil)
	arr.Add(stru)

	buff := new(bytes.Buffer)
	err := BuildParamToNative(buff, arr)
	assert.Nil(t, err)
	assert.Equal(t, "010109ffffffffffffffff00000568656c6c6f", common.ToHexString(buff.Bytes()))
}
