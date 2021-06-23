package evm

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"

	oComm "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/http/ethrpc/utils"
)

func Test_event(t *testing.T) {
	txhash := "0xec56538d2cd67f585560a3769f0694e0b03354eb45258a4b2533cd2ac7cfbd74"
	fmt.Printf("%s", utils.EthToOntHash(common.HexToHash(txhash)).ToHexString())
}

func Test_deseiralizeLog(t *testing.T) {

	//transfer ong from : 0x96216849c49358b10257cb55b28ea603c874b05e to 0x4592d8f8d7b001e72cb26a73e4fa1806a51ac79d amount 10^9 (1 ong)
	states := "0x96216849c49358b10257cb55b28ea603c874b05e03000000ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef00000000000000000000000096216849c49358b10257cb55b28ea603c874b05e0000000000000000000000004592d8f8d7b001e72cb26a73e4fa1806a51ac79d043b9aca00"
	data, err := hexutil.Decode(states)
	assert.Nil(t, err)
	source := oComm.NewZeroCopySource(data)
	var storageLog types.StorageLog
	err = storageLog.Deserialization(source)
	assert.Nil(t, err)

	for _, t := range storageLog.Topics {
		fmt.Printf("%s\n", t.Hex())
	}
	d := big.NewInt(0).SetBytes(storageLog.Data)

	fmt.Printf("%d\n", d)

	assert.Equal(t, len(storageLog.Topics), 3)
	assert.Equal(t, storageLog.Topics[0].Hex(), "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")
	assert.Equal(t, storageLog.Topics[1].Hex(), "0x00000000000000000000000096216849c49358b10257cb55b28ea603c874b05e")
	assert.Equal(t, storageLog.Topics[2].Hex(), "0x0000000000000000000000004592d8f8d7b001e72cb26a73e4fa1806a51ac79d")

	assert.Equal(t, d.Int64(), int64(1000000000))
}
