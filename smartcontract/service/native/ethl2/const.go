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

package ethl2

import (
	"github.com/ontio/ontology/common"
)

const (
	MethodPutName       = "put"
	MethodAppendAddress = "appendaddress"
	MethodGetAddress    = "getaddress"

	MethodSetEthGasLimit = "setgaslimit"
	MethodGetEthGasLimit = "getgaslimit"

	MethodSetMaxEthTxlenByte = "setmaxtxlen"
	MethodGetMaxEthTxlenByte = "getmaxtxlen"

	// key prefix for dict key used in this contract

	PutKeyPrefix         = "ethl2"
	AuthKeyPrefix        = "authaddressset"
	EthGaslimitKeyPrefix = "ethgaslimit"
	EthTxLenKeyPrefix    = "ethtxlenbyte"
)

const (
	EthEIP155Type        = byte(0x00)
	EthSignedMessageType = byte(0x02)
)

func GenPutKey(contract common.Address, input string) []byte {
	return append(contract[:], (PutKeyPrefix + input)...)
}

func GetAppendAutAddressKey(contract common.Address) []byte {
	return append(contract[:], AuthKeyPrefix...)
}

func SetEthTxLenKey(contract common.Address) []byte {
	return append(contract[:], EthTxLenKeyPrefix...)
}

func SetEthGasLimitKey(contract common.Address) []byte {
	return append(contract[:], EthGaslimitKeyPrefix...)
}
