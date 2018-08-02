package handlers

import (
	"github.com/ontio/ontology-crypto/keypair"
	s "github.com/ontio/ontology-crypto/signature"
	"github.com/ontio/ontology/account"
	clisvrcom "github.com/ontio/ontology/cmd/sigsvr/common"
	"github.com/ontio/ontology/common/log"
)

type CreateAccountReq struct {
}

type CreateAccountRsp struct {
	Account string `json:"account"`
}

func CreateAccount(req *clisvrcom.CliRpcRequest, resp *clisvrcom.CliRpcResponse) {
	pwd := req.Pwd
	if pwd == "" {
		resp.ErrorCode = clisvrcom.CLIERR_INVALID_PARAMS
		resp.ErrorInfo = "pwd cannot empty"
		return
	}
	walletPath := req.Wallet
	var err error
	var wallet account.Client
	if walletPath != "" {
		wallet, err = account.Open(walletPath)
		if err != nil {
			resp.ErrorCode = clisvrcom.CLIERR_INVALID_PARAMS
			resp.ErrorInfo = "create wallet failed"
			log.Errorf("CreateAccount Qid:%s create wallet:%s error:%s", req.Qid, walletPath, err)
			return
		}
	} else {
		wallet = clisvrcom.DefWallet
	}
	if wallet == nil {
		resp.ErrorCode = clisvrcom.CLIERR_INVALID_PARAMS
		resp.ErrorInfo = "no wallet to create account"
		return
	}

	acc, err := wallet.NewAccount("", keypair.PK_ECDSA, keypair.P256, s.SHA256withECDSA, []byte(pwd))
	if err != nil {
		resp.ErrorCode = clisvrcom.CLIERR_INTERNAL_ERR
		resp.ErrorInfo = "create wallet failed"
		log.Errorf("CreateAccount Qid:%s create account  error:%s", req.Qid, err)
		return
	}

	resp.Result = &CreateAccountRsp{
		Account: acc.Address.ToBase58(),
	}
}
