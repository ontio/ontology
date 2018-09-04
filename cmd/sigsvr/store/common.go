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
package store

import "encoding/binary"

const (
	DEFAULT_WALLET_NAME = "MyWallet"
	WALLET_VERSION      = "1.1"
	WALLET_INIT_DATA    = "walletStore"
)

const (
	WALLET_INIT_PREFIX               = 0x00
	WALLET_NAME_PREFIX               = 0x01
	WALLET_VERSION_PREFIX            = 0x02
	WALLET_SCRYPT_PREFIX             = 0x03
	WALLET_NEXT_ACCOUNT_INDEX_PREFIX = 0x04
	WALLET_ACCOUNT_INDEX_PREFIX      = 0x05
	WALLET_ACCOUNT_PREFIX            = 0x06
	WALLET_EXTRA_PREFIX              = 0x07
	WALLET_ACCOUNT_NUMBER            = 0x08
)

func GetWalletInitKey() []byte {
	return []byte{WALLET_INIT_PREFIX}
}

func GetWalletNameKey() []byte {
	return []byte{WALLET_NAME_PREFIX}
}

func GetWalletVersionKey() []byte {
	return []byte{WALLET_VERSION_PREFIX}
}

func GetWalletScryptKey() []byte {
	return []byte{WALLET_SCRYPT_PREFIX}
}

func GetAccountIndexKey(index uint32) []byte {
	data := make([]byte, 4, 4)
	binary.LittleEndian.PutUint32(data, index)
	return append([]byte{WALLET_ACCOUNT_INDEX_PREFIX}, data...)
}

func GetNextAccountIndexKey() []byte {
	return []byte{WALLET_NEXT_ACCOUNT_INDEX_PREFIX}
}

func GetAccountKey(address string) []byte {
	return append([]byte{WALLET_ACCOUNT_PREFIX}, []byte(address)...)
}

func GetWalletExtraKey() []byte {
	return []byte{WALLET_EXTRA_PREFIX}
}

func GetWalletAccountNumberKey() []byte {
	return []byte{WALLET_ACCOUNT_NUMBER}
}
