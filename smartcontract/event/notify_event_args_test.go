package event

import (
	"encoding/hex"
	"fmt"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestDeserialization(t *testing.T) {
	data, err := hex.DecodeString("000000000000000000000000000000000000000203000000ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef000000000000000000000000a34886547e00d8f15eaf5a98f99f4f76aaeb3bd500000000000000000000000000000000000000000000000000000000000000072000000000000000000000000000000000000000000000000000ffa4f70a6cd800")
	if err != nil {
		panic(err)
	}
	source := common.NewZeroCopySource(data)
	sl := &types.StorageLog{}
	sl.Deserialization(source)
	fmt.Println(sl.Address.String())
	a := big.NewInt(0).SetBytes(sl.Data)
	fmt.Println(a.String())
	info := NotifyEventInfoFromEvmLog(sl)
	sl2, err := NotifyEventInfoToEvmLog(info)
	assert.Nil(t, err)
	fmt.Println(sl2)
}
