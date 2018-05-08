package handlers

import (
	"encoding/json"
	clisvrcom "github.com/ontio/ontology/cmd/server/common"
	"github.com/ontio/ontology/cmd/utils"
	"testing"
)

func TestSigNeoVMInvokeTx(t *testing.T) {
	invokeReq := &SigNeoVMInvokeTxReq{
		GasPrice: 0,
		GasLimit: 0,
		Address:  "TA6PCtD9qEvN6Rk7i4EGY6u1cWQBWZTr5A",
		Version:  0,
		Params: []interface{}{
			&utils.NeoVMInvokeParam{
				Type:  "string",
				Value: "foo",
			},
			&utils.NeoVMInvokeParam{
				Type: "array",
				Value: []interface{}{
					&utils.NeoVMInvokeParam{
						Type:  "int",
						Value: "0",
					},
					&utils.NeoVMInvokeParam{
						Type:  "bool",
						Value: "true",
					},
				},
			},
		},
	}
	data, err := json.Marshal(invokeReq)
	if err != nil {
		t.Errorf("json.Marshal SigNeoVMInvokeTxReq error:%s", err)
		return
	}
	req := &clisvrcom.CliRpcRequest{
		Qid:    "t",
		Method: "siginvoketx",
		Params: data,
	}
	rsp := &clisvrcom.CliRpcResponse{}
	SigNeoVMInvokeTx(req, rsp)
	if rsp.ErrorCode != 0 {
		t.Errorf("SigNeoVMInvokeTx failed. ErrorCode:%d ErrorInfo:%s", rsp.ErrorCode, rsp.ErrorInfo)
		return
	}
}
