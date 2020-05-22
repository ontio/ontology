/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */
package ontid

import (
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

func Init() {
	native.Contracts[utils.OntIDContractAddress] = RegisterIDContract
}

func RegisterIDContract(srvc *native.NativeService) {
	srvc.Register("regIDWithPublicKey", regIdWithPublicKey)
	srvc.Register("regIDWithController", regIdWithController)
	srvc.Register("revokeID", revokeID)
	srvc.Register("revokeIDByController", revokeIDByController)
	srvc.Register("removeController", removeController)
	srvc.Register("addRecovery", addRecovery)
	srvc.Register("changeRecovery", changeRecovery)
	srvc.Register("setRecovery", setRecovery)
	srvc.Register("updateRecovery", updateRecovery)
	srvc.Register("addKey", addKey)
	srvc.Register("removeKey", removeKey)
	srvc.Register("addKeyByController", addKeyByController)
	srvc.Register("removeKeyByController", removeKeyByController)
	srvc.Register("addKeyByRecovery", addKeyByRecovery)
	srvc.Register("removeKeyByRecovery", removeKeyByRecovery)
	srvc.Register("regIDWithAttributes", regIdWithAttributes)
	srvc.Register("addAttributes", addAttributes)
	srvc.Register("removeAttribute", removeAttribute)
	srvc.Register("addAttributesByController", addAttributesByController)
	srvc.Register("removeAttributeByController", removeAttributeByController)
	srvc.Register("verifySignature", verifySignature)
	srvc.Register("verifyController", verifyController)
	srvc.Register("getPublicKeys", GetPublicKeys)
	srvc.Register("getKeyState", GetKeyState)
	srvc.Register("getAttributes", GetAttributes)
	srvc.Register("getDDO", GetDDO)
	if srvc.Height < config.GetNewOntIdHeight() {
		return
	}
	srvc.Register("removeRecovery", removeRecovery)
	srvc.Register("addKeyByIndex", addKeyByIndex)
	srvc.Register("removeKeyByIndex", removeKeyByIndex)
	srvc.Register("addAttributesByIndex", addAttributesByIndex)
	srvc.Register("removeAttributeByIndex", removeAttributeByIndex)
	srvc.Register("addNewAuthKey", addNewAuthKey)
	srvc.Register("addNewAuthKeyByRecovery", addNewAuthKeyByRecovery)
	srvc.Register("addNewAuthKeyByController", addNewAuthKeyByController)
	srvc.Register("setAuthKey", setAuthKey)
	srvc.Register("setAuthKeyByRecovery", setAuthKeyByRecovery)
	srvc.Register("setAuthKeyByController", setAuthKeyByController)
	srvc.Register("removeAuthKey", removeAuthKey)
	srvc.Register("removeAuthKeyByRecovery", removeAuthKeyByRecovery)
	srvc.Register("removeAuthKeyByController", removeAuthKeyByController)
	srvc.Register("addService", addService)
	srvc.Register("updateService", updateService)
	srvc.Register("removeService", removeService)
	srvc.Register("addContext", addContext)
	srvc.Register("removeContext", removeContext)
	srvc.Register("addProof", addProof)
	srvc.Register("getPublicKeysJson", GetPublicKeysJson)
	srvc.Register("getAttributesJson", GetAttributesJson)
	srvc.Register("getAttributeByKey", GetAttributeByKey)
	srvc.Register("getServiceJson", GetServiceJson)
	srvc.Register("getControllerJson", GetControllerJson)
	srvc.Register("getDocumentJson", GetDocumentJson)
	return
}
