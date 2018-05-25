package handlers

import (
	"testing"
	"encoding/json"
	"github.com/ontio/ontology/cmd/abi"
	clisvrcom "github.com/ontio/ontology/cmd/server/common"
)

func TestSigNativeInvokeTx(t *testing.T) {
	invokeReq := &SigNativeInvokeTxReq{
		GasPrice: 0,
		GasLimit: 40000,
		Address:  "ff00000000000000000000000000000000000002",
		Method:   "transfer",
		Version:  0,
		Params: []interface{}{
			[]interface{}{
				[]interface{}{
					"TA587BCw7HFwuUuzY1wg2HXCN7cHBPaXSe",
					"TA5gYXCSiUq9ejGCa54M3yoj9kfMv3ir4j",
					"10000000000",
				},
			},
		},
	}
	data, err := json.Marshal(invokeReq)
	if err != nil {
		t.Errorf("json.Marshal SigNativeInvokeTxReq error:%s", err)
		return
	}
	req := &clisvrcom.CliRpcRequest{
		Qid:    "t",
		Method: "signativeinvoketx",
		Params: data,
	}
	rsp := &clisvrcom.CliRpcResponse{}
	abi.DefAbiMgr.Path = "../../abi"
	abi.DefAbiMgr.Init()
	SigNativeInvokeTx(req, rsp)
	if rsp.ErrorCode != 0 {
		t.Errorf("SigNativeInvokeTx failed. ErrorCode:%d ErrorInfo:%s", rsp.ErrorCode, rsp.ErrorInfo)
		return
	}
}
