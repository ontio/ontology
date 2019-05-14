/*
 * Copyright (C) 2019 The DAD Authors
 * This file is part of The DAD library.
 *
 * The DAD is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The DAD is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The DAD.  If not, see <http://www.gnu.org/licenses/>.
 */

// homomorphic encryption contract:
// User can use those API supported by this contract's to encrypt privacy data on chain
// to pretect privacy data
package homomorphicencryption

import (
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

const ()

const (
	//INIT_CONFIG function name
	INIT_CONFIG = "initConfig"

	//key prefix

	//global
)

//InitHomomorphicEncryption init contract address
func InitHomomorphicEncryption() {
	native.Contracts[utils.HomomorphicEncryptionContractAddress] = RegisterHomomorphicEncryptionContract
}

//RegisterHomomorphicEncryptionContract methods of HomomorphicEncryption contract
func RegisterHomomorphicEncryptionContract(native *native.NativeService) {
	native.Register(INIT_CONFIG, InitConfig)
}

//InitConfig HomomorphicEncryption contract, include vbft config, global param and ontid admin.
func InitConfig(native *native.NativeService) ([]byte, error) {
	return utils.BYTE_TRUE, nil
}
