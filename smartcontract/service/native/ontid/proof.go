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
	"errors"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

func addProof(srvc *native.NativeService) ([]byte, error) {
	return utils.BYTE_FALSE, errors.New("property \"proof\" in ONT ID document is not supported yet")
}

func getProof(srvc *native.NativeService, encId []byte) (string, error) {
	key := append(encId, FIELD_PROOF)
	proofStore, err := utils.GetStorageItem(srvc, key)
	if err != nil {
		return "", errors.New("getProof error:" + err.Error())
	}
	if proofStore == nil {
		return "", nil
	}
	source := common.NewZeroCopySource(proofStore.Value)
	proof, err := utils.DecodeVarBytes(source)
	if err != nil {
		return "", errors.New("DecodeVarBytes error:" + err.Error())
	}
	return string(proof), nil
}
