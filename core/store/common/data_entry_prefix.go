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
	DATA_Block DataEntryPrefix = iota
	DATA_Header = 0x01
	DATA_Transaction = 0x02

	// Transaction
	ST_Bookkeeper = 0x03
	ST_Contract = 0x04
	ST_Storage = 0x05
	ST_Validator = 0x07
	ST_Vote = 0x08

	IX_HeaderHashList = 0x09

	//SYSTEM
	SYS_CurrentBlock = 0x10
	SYS_Version = 0x11
	SYS_CurrentStateRoot = 0x12
	SYS_BlockMerkleTree = 0x13

	EVENT_Notify = 0x14
)
