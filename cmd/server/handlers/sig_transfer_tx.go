package handlers

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	clisvrcom "github.com/ontio/ontology/cmd/server/common"
	cliutil "github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/common/log"
)

type SigTransferTransactionReq struct {
	GasPrice uint64 `json:"gas_price"`
	GasLimit uint64 `json:"gas_limit"`
	Asset    string `json:"asset"`
	From     string `json:"from"`
	To       string `json:"to"`
	Amount   uint64 `json:"amount"`
}

type SinTransferTransactionRsp struct {
	SignedTx string `json:"signed_tx"`
}

func SigTransferTransaction(req *clisvrcom.CliRpcRequest, resp *clisvrcom.CliRpcResponse) {
	rawReq := &SigTransferTransactionReq{}
	err := json.Unmarshal(req.Params, rawReq)
	if err != nil {
		resp.ErrorCode = clisvrcom.CLIERR_INVALID_PARAMS
		return
	}
	transferTx, err := cliutil.TransferTx(rawReq.GasPrice, rawReq.GasLimit, rawReq.Asset, rawReq.From, rawReq.To, rawReq.Amount)
	if err != nil {
		resp.ErrorCode = clisvrcom.CLIERR_INVALID_PARAMS
		return
	}
	signer := clisvrcom.DefAccount
	if signer == nil {
		resp.ErrorCode = clisvrcom.CLIERR_ACCOUNT_UNLOCK
		return
	}
	err = cliutil.SignTransaction(signer, transferTx)
	if err != nil {
		log.Infof("Cli Qid:%s SigTransferTransaction SignTransaction error:%s", req.Qid, err)
		resp.ErrorCode = clisvrcom.CLIERR_INTERNAL_ERR
		return
	}
	buf := bytes.NewBuffer(nil)
	err = transferTx.Serialize(buf)
	if err != nil {
		log.Infof("Cli Qid:%s SigTransferTransaction tx Serialize error:%s", req.Qid, err)
		resp.ErrorCode = clisvrcom.CLIERR_INTERNAL_ERR
		return
	}
	resp.Result = &SinTransferTransactionRsp{
		SignedTx: hex.EncodeToString(buf.Bytes()),
	}
}
