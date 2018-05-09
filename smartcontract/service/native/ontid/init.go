package ontid

import (
	"github.com/ontio/ontology/core/genesis"
	"github.com/ontio/ontology/smartcontract/service/native"
)

var contractAddress = genesis.OntIDContractAddress[:]

func init() {
	native.Contracts[genesis.OntIDContractAddress] = RegisterIDContract
}

func RegisterIDContract(srvc *native.NativeService) {
	srvc.Register("regIDWithPublicKey", regIdWithPublicKey)
	srvc.Register("addKey", addKey)
	srvc.Register("removeKey", removeKey)
	srvc.Register("addRecovery", addRecovery)
	srvc.Register("changeRecovery", changeRecovery)
	srvc.Register("regIDWithAttributes", regIdWithAttributes)
	srvc.Register("addAttribute", addAttribute)
	srvc.Register("removeAttribute", removeAttribute)
	srvc.Register("verifySignature", verifySignature)
	return
}
