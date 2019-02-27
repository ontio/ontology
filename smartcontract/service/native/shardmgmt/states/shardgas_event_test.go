package shardstates

import (
	"bytes"
	"testing"
)

func TestUserGasInfoSerialzie(t *testing.T) {
	userGasInfo := &UserGasInfo{
		Balance:         10000,
		PendingWithdraw: make([]*GasWithdrawInfo, 0),
	}
	serBuffer := new(bytes.Buffer)
	err := userGasInfo.Serialize(serBuffer)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("serialize result is %x", serBuffer.Bytes())
	deseBuffer := bytes.NewBuffer(serBuffer.Bytes())
	deseInfo := &UserGasInfo{}
	err = deseInfo.Deserialize(deseBuffer)
	if err != nil {
		t.Fatal(err)
	}
}
