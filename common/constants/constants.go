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

package constants

import (
	"time"

	"github.com/laizy/bigint"
)

// genesis constants
var (
	//TODO: modify this when on mainnet
	GENESIS_BLOCK_TIMESTAMP = uint32(time.Date(2018, time.June, 30, 0, 0, 0, 0, time.UTC).Unix())

	CHANGE_UNBOUND_TIMESTAMP_MAINNET = uint32(time.Date(2020, time.July, 7, 0, 0, 0, 0, time.UTC).Unix())
	CHANGE_UNBOUND_TIMESTAMP_POLARIS = uint32(time.Date(2020, time.June, 28, 0, 0, 0, 0, time.UTC).Unix())
)

// ont constants
const GWei = 1000000000

const (
	ONT_NAME            = "ONT Token"
	ONT_SYMBOL          = "ONT"
	ONT_DECIMALS        = 0
	ONT_DECIMALS_V2     = 9
	ONT_TOTAL_SUPPLY    = 1000000000
	ONT_TOTAL_SUPPLY_V2 = 1000000000000000000
)

// ong constants
const (
	ONG_NAME         = "ONG Token"
	ONG_SYMBOL       = "ONG"
	ONG_DECIMALS     = 9
	ONG_DECIMALS_V2  = 18
	ONG_TOTAL_SUPPLY = 1000000000000000000
)

var (
	ONG_TOTAL_SUPPLY_V2 = bigint.New(10).ExpUint8(27)
)

// ont/ong unbound model constants
const UNBOUND_TIME_INTERVAL = uint32(31536000)

var UNBOUND_GENERATION_AMOUNT = [18]uint64{5, 4, 3, 3, 2, 2, 2, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
var NEW_UNBOUND_GENERATION_AMOUNT = [18]uint64{5, 4, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 2, 2, 2, 3, 3}

// multi-sig constants
const MULTI_SIG_MAX_PUBKEY_SIZE = 16

// transaction constants
const TX_MAX_SIG_SIZE = 16

// network magic number
const (
	NETWORK_MAGIC_MAINNET = 0x8c77ab60
	NETWORK_MAGIC_POLARIS = 0xdf29882d
)

const (
	EIP155_CHAINID_MAINNET = 58
	EIP155_CHAINID_POLARIS = 5851
)

// ledger state hash check height
const STATE_HASH_HEIGHT_MAINNET = 3000000
const STATE_HASH_HEIGHT_POLARIS = 0

// neovm opcode update check height
const OPCODE_HEIGHT_UPDATE_FIRST_MAINNET = 6300000
const OPCODE_HEIGHT_UPDATE_FIRST_POLARIS = 0

// gas round tune operation height
const GAS_ROUND_TUNE_HEIGHT_MAINNET = 8500000
const GAS_ROUND_TUNE_HEIGHT_POLARIS = 0

const CONTRACT_DEPRECATE_API_HEIGHT_MAINNET = 8600000
const CONTRACT_DEPRECATE_API_HEIGHT_POLARIS = 0

// self gov register height
const BLOCKHEIGHT_SELFGOV_REGISTER_MAINNET = 8600000
const BLOCKHEIGHT_SELFGOV_REGISTER_POLARIS = 0

const BLOCKHEIGHT_NEW_ONTID_MAINNET = 9000000
const BLOCKHEIGHT_NEW_ONTID_POLARIS = 0

const BLOCKHEIGHT_ONTFS_MAINNET = 8550000
const BLOCKHEIGHT_ONTFS_POLARIS = 0

const BLOCKHEIGHT_CC_POLARIS = 0

// new node cost height
const BLOCKHEIGHT_NEW_PEER_COST_MAINNET = 9400000
const BLOCKHEIGHT_NEW_PEER_COST_POLARIS = 0

const BLOCKHEIGHT_TRACK_DESTROYED_CONTRACT_MAINNET = 11700000
const BLOCKHEIGHT_TRACK_DESTROYED_CONTRACT_POLARIS = 0

const USER_FEE_SPLIT_OVERFLOW_MAINNET = 16490000
const USER_FEE_SPLIT_OVERFLOW_POLARIS = 0

const UINT64_WRAPPING_MAINNET = 17370000

var (
	BLOCKHEIGHT_ADD_DECIMALS_MAINNET = uint32(13920000)
	BLOCKHEIGHT_ADD_DECIMALS_POLARIS = uint32(0)
)
