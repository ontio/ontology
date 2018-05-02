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

package vconfig

import (
	"testing"

	"github.com/ontio/ontology/common/log"
)

const (
	DefaultConfigFileName = "../../../config-vbft.json"
)

func TestGenConsensusPayload(t *testing.T) {
	log.Init(log.PATH, log.Stdout)
	res, err := genConsensusPayload()
	if err != nil {
		t.Errorf("test failed: %v", err)
	} else {
		t.Logf("config content: %v\n", res)
	}
}

func TestGenesisChainConfig(t *testing.T) {
	log.Init(log.PATH, log.Stdout)
	config := &Configuration{
		View:                 1,
		N:                    7,
		C:                    1,
		K:                    4,
		L:                    64,
		InitTxid:             0,
		GenesisTimestamp:     0,
		BlockMsgDelay:        10000,
		HashMsgDelay:         10000,
		PeerHandshakeTimeout: 10,
	}
	var peersinfo []*PeerStakeInfo
	peer1 := &PeerStakeInfo{
		Index:  0,
		NodeID: "120202c924ed1a67fd1719020ce599d723d09d48362376836e04b0be72dfe825e24d810000",
		Stake:  1000,
	}
	peer2 := &PeerStakeInfo{
		Index:  1,
		NodeID: "120202935fb8d28b70706de6014a937402a30ae74a56987ed951abbe1ac9eeda56f0160000",
		Stake:  2000,
	}

	peer3 := &PeerStakeInfo{
		Index:  2,
		NodeID: "120202172f290c6d63b8014573c7722a72ccf778dd36272519fe0ff0b8b1281ec56b880000",
		Stake:  3000,
	}

	peer4 := &PeerStakeInfo{
		Index:  3,
		NodeID: "1202036db3da9deb8bea20b1024944946ef3e6fcb71367faa096c63ad5ae97fc7af7a10000",
		Stake:  4000,
	}
	peersinfo = append(peersinfo, peer1, peer2, peer3, peer4)
	chainconfig, err := GenesisChainConfig(config, peersinfo)
	if err != nil {
		t.Errorf("TestGenesisChainConfig failed:%s", err)
		return
	}
	t.Logf("TestGenesisChainConfig succ: %v", chainconfig.PosTable)
}
