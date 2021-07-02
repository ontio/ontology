package tests

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	types2 "github.com/ethereum/go-ethereum/core/types"
	oComm "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestNewERC20(t *testing.T) {
	states := "0x96216849c49358b10257cb55b28ea603c874b05e03000000ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef00000000000000000000000096216849c49358b10257cb55b28ea603c874b05e0000000000000000000000004592d8f8d7b001e72cb26a73e4fa1806a51ac79d043b9aca00"
	data, err := hexutil.Decode(states)
	assert.Nil(t, err)
	source := oComm.NewZeroCopySource(data)
	var storageLog types.StorageLog
	err = storageLog.Deserialization(source)
	assert.Nil(t, err)

	parsed, _ := abi.JSON(strings.NewReader(ERC20ABI))
	nbc := bind.NewBoundContract(common.Address{}, parsed, nil, nil, nil)

	tf := new(ERC20Transfer)
	l := types2.Log{
		Address:     storageLog.Address,
		Topics:      storageLog.Topics,
		Data:        storageLog.Data,
	}
	err = nbc.UnpackLog(tf, "Transfer", l)
	assert.Nil(t, err)
}
