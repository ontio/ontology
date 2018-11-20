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

package common

// DataEntryPrefix
type DataEntryPrefix byte

const (
	// DATA
	DATA_BLOCK             DataEntryPrefix = 0x00 //Block height => block hash key prefix
	DATA_HEADER                            = 0x01 //Block hash => block hash key prefix
	DATA_TRANSACTION                       = 0x02 //Transction hash = > transaction key prefix
	DATA_STATE_MERKLE_ROOT                 = 0x21 // block height => write set hash + state merkle root

	// Transaction
	ST_BOOKKEEPER DataEntryPrefix = 0x03 //BookKeeper state key prefix
	ST_CONTRACT   DataEntryPrefix = 0x04 //Smart contract state key prefix
	ST_STORAGE    DataEntryPrefix = 0x05 //Smart contract storage key prefix
	ST_VALIDATOR  DataEntryPrefix = 0x07 //no use
	ST_VOTE       DataEntryPrefix = 0x08 //Vote state key prefix

	IX_HEADER_HASH_LIST DataEntryPrefix = 0x09 //Block height => block hash key prefix

	//SYSTEM
	SYS_CURRENT_BLOCK      DataEntryPrefix = 0x10 //Current block key prefix
	SYS_VERSION            DataEntryPrefix = 0x11 //Store version key prefix
	SYS_CURRENT_STATE_ROOT DataEntryPrefix = 0x12 //no use
	SYS_BLOCK_MERKLE_TREE  DataEntryPrefix = 0x13 // Block merkle tree root key prefix
	SYS_STATE_MERKLE_TREE  DataEntryPrefix = 0x20 // state merkle tree root key prefix

	EVENT_NOTIFY DataEntryPrefix = 0x14 //Event notify key prefix
)
