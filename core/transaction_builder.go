package core

import (
	"github.com/Ontology/core/code"
	"github.com/Ontology/core/payload"
	"github.com/Ontology/core/types"
	"github.com/Ontology/crypto"
)

//initial a new transaction with asset registration payload
func NewBookKeeperTransaction(pubKey *crypto.PubKey, isAdd bool, cert []byte, issuer *crypto.PubKey) (*types.Transaction, error) {
	bookKeeperPayload := &payload.BookKeeper{
		PubKey: pubKey,
		Action: payload.BookKeeperAction_SUB,
		Cert:   cert,
		Issuer: issuer,
	}

	if isAdd {
		bookKeeperPayload.Action = payload.BookKeeperAction_ADD
	}

	return &types.Transaction{
		TxType:     types.BookKeeper,
		Payload:    bookKeeperPayload,
		Attributes: nil,
	}, nil
}

func NewDeployTransaction(fc *code.FunctionCode, name, codeversion, author, email, desp string, vmType types.VmType, needStorage bool) *types.Transaction {
	//TODO: check arguments
	DeployCodePayload := &payload.DeployCode{
		Code:        fc,
		NeedStorage: needStorage,
		Name:        name,
		CodeVersion: codeversion,
		Author:      author,
		Email:       email,
		Description: desp,
	}

	return &types.Transaction{
		TxType:     types.Deploy,
		Payload:    DeployCodePayload,
		Attributes: nil,
	}
}

func NewInvokeTransaction(vmcode types.VmCode, param []byte) (*types.Transaction, error) {
	//TODO: check arguments
	invokeCodePayload := &payload.InvokeCode{
		Code:   vmcode,
		Params: param,
	}

	return &types.Transaction{
		TxType:     types.Invoke,
		Payload:    invokeCodePayload,
		Attributes: nil,
	}, nil
}
