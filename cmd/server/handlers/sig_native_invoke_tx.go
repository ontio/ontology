package handlers

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ontio/ontology/cmd/abi"
	clisvrcom "github.com/ontio/ontology/cmd/server/common"
	cliutil "github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
)

type SigNativeInvokeTxReq struct {
	GasPrice uint64        `json:"gas_price"`
	GasLimit uint64        `json:"gas_limit"`
	Address  string        `json:"address"`
	Method   string        `json:"method"`
	Params   []interface{} `json:"params"`
	Version  byte          `json:"version"`
}

type SigNativeInvokeTxRsp struct {
	SignedTx string `json:"signed_tx"`
}

func SigNativeInvokeTx(req *clisvrcom.CliRpcRequest, resp *clisvrcom.CliRpcResponse) {
	rawReq := &SigNativeInvokeTxReq{}
	err := json.Unmarshal(req.Params, rawReq)
	if err != nil {
		log.Infof("Cli Qid:%s SigNativeInvokeTx json.Unmarshal SigNativeInvokeTxReq:%s error:%s", req.Qid, req.Params, err)
		resp.ErrorCode = clisvrcom.CLIERR_INVALID_PARAMS
		return
	}
	addrData, err := hex.DecodeString(rawReq.Address)
	if err != nil {
		log.Infof("Cli Qid:%s SigNativeInvokeTx hex.DecodeString address:%s error:%s", rawReq.Address, err)
		return
	}
	contractAddr, err := common.AddressParseFromBytes(addrData)
	if err != nil {
		log.Infof("Cli Qid:%s SigNativeInvokeTx AddressParseFromBytes:%s error:%s", req.Qid, rawReq.Address, err)
		resp.ErrorCode = clisvrcom.CLIERR_INVALID_PARAMS
		return
	}
	nativeAbi := abi.DefAbiMgr.GetNativeAbi(rawReq.Address)
	if nativeAbi == nil {
		resp.ErrorCode = clisvrcom.CLIERR_ABI_NOT_FOUND
		return
	}
	funcAbi := nativeAbi.GetFunc(rawReq.Method)
	if funcAbi == nil {
		resp.ErrorCode = clisvrcom.CLIERR_ABI_NOT_FOUND
		return
	}
	data, err := cliutil.ParseNativeParam(rawReq.Params, funcAbi.Parameters)
	if err != nil {
		resp.ErrorCode = clisvrcom.CLIERR_ABI_UNMATCH
		resp.ErrorInfo = fmt.Sprintf("Parse param error:%s", err)
		return
	}
	tx, err := cliutil.InvokeNativeContractTx(rawReq.GasPrice, rawReq.GasLimit, rawReq.Version, contractAddr, rawReq.Method, data)
	if err != nil {
		log.Infof("Cli Qid:%s SigNativeInvokeTx InvokeNativeContractTx error:%s", req.Qid, err)
		resp.ErrorCode = clisvrcom.CLIERR_INTERNAL_ERR
		return
	}
	signer := clisvrcom.DefAccount
	err = cliutil.SignTransaction(signer, tx)
	if err != nil {
		log.Infof("Cli Qid:%s SigNativeInvokeTx SignTransaction error:%s", req.Qid, err)
		resp.ErrorCode = clisvrcom.CLIERR_INTERNAL_ERR
		return
	}
	buf := bytes.NewBuffer(nil)
	err = tx.Serialize(buf)
	if err != nil {
		log.Infof("Cli Qid:%s SigNativeInvokeTx tx Serialize error:%s", req.Qid, err)
		resp.ErrorCode = clisvrcom.CLIERR_INTERNAL_ERR
		return
	}
	resp.Result = &SigNativeInvokeTxRsp{
		SignedTx: hex.EncodeToString(buf.Bytes()),
	}
}
