package oep4

import (
	"encoding/json"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestOep4(t *testing.T) {
	state := &Oep4{
		Name:        "TestToken",
		Symbol:      "TT",
		Decimals:    12,
		TotalSupply: big.NewInt(1000000000),
	}
	sink := common.NewZeroCopySink(0)
	state.Serialization(sink)
	source := common.NewZeroCopySource(sink.Bytes())
	newState := &Oep4{}
	err := newState.Deserialization(source)
	assert.Nil(t, err)
}

func TestXShardTransferState(t *testing.T) {
	acc := account.NewAccount("")
	state := &XShardTransferState{
		ToShard:   types.NewShardIDUnchecked(39),
		ToAccount: acc.Address,
		Amount:    big.NewInt(384747),
		Status:    XSHARD_TRANSFER_COMPLETE,
	}
	sink := common.NewZeroCopySink(0)
	state.Serialization(sink)
	source := common.NewZeroCopySource(sink.Bytes())
	newState := &XShardTransferState{}
	err := newState.Deserialization(source)
	assert.Nil(t, err)
	data, err := json.Marshal(state)
	assert.Nil(t, err)
	t.Logf("marshal state is %s", string(data))
}
