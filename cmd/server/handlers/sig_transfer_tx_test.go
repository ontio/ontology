package handlers

import (
	"encoding/json"
	"github.com/ontio/ontology/account"
	clisvrcom "github.com/ontio/ontology/cmd/server/common"
	"testing"
)

func TestSigTransferTransaction(t *testing.T) {
	acc := account.NewAccount("")
	defAcc := clisvrcom.DefAccount
	sigReq := &SigTransferTransactionReq{
		GasLimit: 0,
		GasPrice: 0,
		Asset:    "ont",
		From:     defAcc.Address.ToBase58(),
		To:       acc.Address.ToBase58(),
		Amount:   10,
	}
	data, err := json.Marshal(sigReq)
	if err != nil {
		t.Errorf("json.Marshal SigTransferTransactionReq error:%s", err)
	}
	req := &clisvrcom.CliRpcRequest{
		Qid:    "t",
		Method: "sigtransfertx",
		Params: data,
	}
	rsp := &clisvrcom.CliRpcResponse{}
	SigTransferTransaction(req, rsp)
	if rsp.ErrorCode != 0 {
		t.Errorf("SigTransferTransaction failed. ErrorCode:%d", rsp.ErrorCode)
		return
	}
}
