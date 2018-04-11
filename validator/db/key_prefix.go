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

package db

import (
	pool "github.com/valyala/bytebufferpool"

	"github.com/ontio/ontology/common"
)

// DataEntryPrefix
type KeyPrefix byte

const (
	//SYSTEM
	SYS_VERSION       KeyPrefix = 0
	SYS_GENESIS_BLOCK KeyPrefix = 1 // key: prefix, value: gensisBlock

	SYS_BEST_BLOCK        KeyPrefix = 2 // key : prefix, value: bestblock
	SYS_BEST_BLOCK_HEADER KeyPrefix = 3 // key: prefix, value: BlockHeader

	// DATA
	//DATA_Block KeyPrefix = iota
	//DATA_Header
	DATA_TRANSACTION KeyPrefix = 10 // key: prefix+txid, value: height + tx
)

func GenGenesisBlockKey() *pool.ByteBuffer {
	key := keyPool.Get()
	key.WriteByte(byte(SYS_GENESIS_BLOCK))
	return key
}

func GenBestBlockHeaderKey() *pool.ByteBuffer {
	key := keyPool.Get()
	key.WriteByte(byte(SYS_BEST_BLOCK_HEADER))
	return key
}

func GenDataTransactionKey(hash common.Uint256) *pool.ByteBuffer {
	key := keyPool.Get()
	key.WriteByte(byte(DATA_TRANSACTION))
	key.Write(hash.ToArray())
	return key
}
