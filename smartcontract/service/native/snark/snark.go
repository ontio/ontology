package snark

import (
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

func InitSNARK() {
	native.Contracts[utils.SNARKContractAddress] = RegisterSNARKContract
}

func RegisterSNARKContract(native *native.NativeService) {
	native.Register("ecAdd", ECAdd)
	native.Register("twistAdd", TwistECAdd)
	native.Register("ecMul", ECMul)
	native.Register("twistMul", TwistECMul)
	native.Register("pairingCheck", PairingCheck)
	native.Register("phgr13", PHGR13Verify)
}
