package core

import (
	"bytes"

	"github.com/Ontology/core/payload"
	"github.com/Ontology/core/types"
	vmtypes "github.com/Ontology/vm/types"
)

type TransactionExector interface {
	Execute(tx *types.Transaction)
}

type transactionExecutor struct {
	db map[types.Address]int64
}

func (self *transactionExecutor) Execute(tx *types.Transaction) {
	switch pld := tx.Payload.(type) {
	case *payload.InvokeCode:
		vmcode := pld.Code
		if vmcode.CodeType == vmtypes.NativeVM {
			if bytes.Equal(vmcode.Code, []byte("ont")) {
				if bytes.Equal(pld.Params[:8], []byte("transfer")) {
					var from, to types.Address
					copy(from[:], pld.Params[8:28])
					copy(to[:], pld.Params[28:48])
					value := int64(10)
					if self.db[from] >= value {
						self.db[from] -= value
						self.db[to] += value
					}
				}
			}
		}
	}
}

func NewONTTransferTransaction(from, to types.Address) *types.Transaction {
	code := []byte("ont")
	params := append([]byte("transfer"), from[:]...)
	params = append(params, to[:]...)
	vmcode := vmtypes.VmCode{
		CodeType: vmtypes.NativeVM,
		Code:     code,
	}

	tx, _ := NewInvokeTransaction(vmcode, params)
	return tx
}
