package handlers

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	clisvrcom "github.com/ontio/ontology/cmd/server/common"
	cliutil "github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
)

type SigNeoVMInvokeTxReq struct {
	GasPrice uint64        `json:"gas_price"`
	GasLimit uint64        `json:"gas_limit"`
	Address  string        `json:"address"`
	Params   []interface{} `json:"params"`
	Version  byte          `json:"version"`
}

type SigNeoVMInvokeTxRsp struct {
	SignedTx string `json:"signed_tx"`
}

func SigNeoVMInvokeTx(req *clisvrcom.CliRpcRequest, resp *clisvrcom.CliRpcResponse) {
	rawReq := &SigNeoVMInvokeTxReq{}
	err := json.Unmarshal(req.Params, rawReq)
	if err != nil {
		log.Infof("SigNeoVMInvokeTx json.Unmarshal SigNeoVMInvokeTxReq:%s error:%s", req.Params, err)
		resp.ErrorCode = clisvrcom.CLIERR_INVALID_PARAMS
		return
	}
	params, err := cliutil.ParseNeoVMInvokeParams(rawReq.Params)
	if err != nil {
		resp.ErrorCode = clisvrcom.CLIERR_INVALID_PARAMS
		resp.ErrorInfo = fmt.Sprintf("ParseNeoVMInvokeParams error:%s", err)
		return
	}
	contAddr, err := common.AddressFromBase58(rawReq.Address)
	if err != nil {
		log.Infof("SigNeoVMInvokeTx contract AddressFromBase58:%s error:%s", rawReq.Address, err)
		resp.ErrorCode = clisvrcom.CLIERR_INVALID_PARAMS
		return
	}
	tx, err := cliutil.InvokeNeoVMContractTx(rawReq.GasPrice, rawReq.GasLimit, rawReq.Version, contAddr, params)
	if err != nil {
		log.Infof("SigNeoVMInvokeTx InvokeNeoVMContractTx error:%s", err)
		resp.ErrorCode = clisvrcom.CLIERR_INVALID_PARAMS
		return
	}
	signer := clisvrcom.DefAccount
	err = cliutil.SignTransaction(signer, tx)
	if err != nil {
		log.Infof("Cli Qid:%s SigNeoVMInvokeTx SignTransaction error:%s", req.Qid, err)
		resp.ErrorCode = clisvrcom.CLIERR_INTERNAL_ERR
		return
	}
	buf := bytes.NewBuffer(nil)
	err = tx.Serialize(buf)
	if err != nil {
		log.Infof("Cli Qid:%s SigNeoVMInvokeTx tx Serialize error:%s", req.Qid, err)
		resp.ErrorCode = clisvrcom.CLIERR_INTERNAL_ERR
		return
	}
	resp.Result = &SigNeoVMInvokeTxRsp{
		SignedTx: hex.EncodeToString(buf.Bytes()),
	}
}
