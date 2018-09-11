package actor

import (
	"bytes"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

func updateNativeSCAddr(hash common.Address) common.Address {
	if hash == utils.OntContractAddress || bytes.Equal(common.ToArrayReverse(hash[:]), utils.OntContractAddress[:]) {
		hash = types.AddressFromVmCode(utils.OntContractAddress[:])
	} else if hash == utils.OngContractAddress || bytes.Equal(common.ToArrayReverse(hash[:]), utils.OngContractAddress[:]) {
		hash = types.AddressFromVmCode(utils.OngContractAddress[:])
	} else if hash == utils.OntIDContractAddress || bytes.Equal(common.ToArrayReverse(hash[:]), utils.OntIDContractAddress[:]) {
		hash = types.AddressFromVmCode(utils.OntIDContractAddress[:])
	} else if hash == utils.ParamContractAddress || bytes.Equal(common.ToArrayReverse(hash[:]), utils.ParamContractAddress[:]) {
		hash = types.AddressFromVmCode(utils.ParamContractAddress[:])
	} else if hash == utils.AuthContractAddress || bytes.Equal(common.ToArrayReverse(hash[:]), utils.AuthContractAddress[:]) {
		hash = types.AddressFromVmCode(utils.AuthContractAddress[:])
	} else if hash == utils.GovernanceContractAddress || bytes.Equal(common.ToArrayReverse(hash[:]), utils.GovernanceContractAddress[:]) {
		hash = types.AddressFromVmCode(utils.GovernanceContractAddress[:])
	}
	return hash
}
