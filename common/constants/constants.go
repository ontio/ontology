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
	"fmt"
	"time"
)

// genesis constants
var (
	//TODO: modify this when on mainnet
	GENESIS_BLOCK_TIMESTAMP = uint32(time.Date(2018, time.June, 12, 0, 0, 0, 0, time.UTC).Unix())
)

// ont constants
const (
	ONT_NAME         = "ONT Token"
	ONT_SYMBOL       = "ONT"
	ONT_DECIMALS     = 1
	ONT_TOTAL_SUPPLY = uint64(1000000000)
)

// ong constants
const (
	ONG_NAME         = "ONG Token"
	ONG_SYMBOL       = "ONG"
	ONG_DECIMALS     = 9
	ONG_TOTAL_SUPPLY = uint64(1000000000000000000)
)

// ont/ong unbound model constants
const UNBOUND_TIME_INTERVAL = uint32(31536000)

var UNBOUND_GENERATION_AMOUNT = [18]uint64{5, 4, 3, 3, 2, 2, 2, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}

// the end of unbound timestamp offset from genesis block's timestamp
var UNBOUND_DEADLINE = (func() uint32 {
	count := uint64(0)
	for _, m := range UNBOUND_GENERATION_AMOUNT {
		count += m
	}
	count *= uint64(UNBOUND_TIME_INTERVAL)

	numInterval := len(UNBOUND_GENERATION_AMOUNT)

	if UNBOUND_GENERATION_AMOUNT[numInterval-1] != 1 ||
		!(count-uint64(UNBOUND_TIME_INTERVAL) < ONT_TOTAL_SUPPLY && ONT_TOTAL_SUPPLY <= count) {
		panic("incompatible constants setting")
	}

	return UNBOUND_TIME_INTERVAL*uint32(numInterval) - uint32(count-uint64(ONT_TOTAL_SUPPLY))
})()

// multi-sig constants
const MULTI_SIG_MAX_PUBKEY_SIZE = 1024

const (
	NETWORK_ID_MAIN_NET      = 1
	NETWORK_ID_POLARIS_NET   = 2
	NETWORK_ID_SOLO_NET      = 3
	NETWORK_NAME_MAIN_NET    = "ontology"
	NETWORK_NAME_POLARIS_NET = "polaris"
	NETWORK_NAME_SOLO_NET    = "testmode"
)

var NETWORK_MAGIC = map[uint32]uint32{
	NETWORK_ID_MAIN_NET:    0x74746e41, //Network main
	NETWORK_ID_POLARIS_NET: 0x01015401, //Network polaris
	NETWORK_ID_SOLO_NET:    0x00000000, //Network solo
}

var NETWORK_NAME = map[uint32]string{
	NETWORK_ID_MAIN_NET:    NETWORK_NAME_MAIN_NET,
	NETWORK_ID_POLARIS_NET: NETWORK_NAME_POLARIS_NET,
	NETWORK_ID_SOLO_NET:    NETWORK_NAME_SOLO_NET,
}

func GetNetworkMagic(id uint32) uint32 {
	nid, ok := NETWORK_MAGIC[id]
	if ok {
		return nid
	}
	return id
}

func GetNetworkName(id uint32) string {
	name, ok := NETWORK_NAME[id]
	if ok {
		return name
	}
	return fmt.Sprintf("%d", id)
}
