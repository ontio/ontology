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

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
)

func constructConfig() (*config.VBFTConfig, error) {
	conf := &config.VBFTConfig{
		N:                    7,
		C:                    2,
		K:                    7,
		L:                    112,
		BlockMsgDelay:        10000,
		HashMsgDelay:         10000,
		PeerHandshakeTimeout: 10,
		MaxBlockChangeView:   1000,
	}
	var peersinfo []*config.VBFTPeerStakeInfo
	peer1 := &config.VBFTPeerStakeInfo{
		Index:      1,
		PeerPubkey: "0253ccfd439b29eca0fe90ca7c6eaa1f98572a054aa2d1d56e72ad96c466107a85",
		InitPos:    0,
	}
	peer2 := &config.VBFTPeerStakeInfo{
		Index:      2,
		PeerPubkey: "035eb654bad6c6409894b9b42289a43614874c7984bde6b03aaf6fc1d0486d9d45",
		InitPos:    0,
	}

	peer3 := &config.VBFTPeerStakeInfo{
		Index:      3,
		PeerPubkey: "0281d198c0dd3737a9c39191bc2d1af7d65a44261a8a64d6ef74d63f27cfb5ed92",
		InitPos:    0,
	}

	peer4 := &config.VBFTPeerStakeInfo{
		Index:      4,
		PeerPubkey: "023967bba3060bf8ade06d9bad45d02853f6c623e4d4f52d767eb56df4d364a99f",
		InitPos:    0,
	}
	peer5 := &config.VBFTPeerStakeInfo{
		Index:      5,
		PeerPubkey: "038bfc50b0e3f0e5df6d451069065cbfa7ab5d382a5839cce82e0c963edb026e94",
		InitPos:    0,
	}
	peer6 := &config.VBFTPeerStakeInfo{
		Index:      6,
		PeerPubkey: "03f1095289e7fddb882f1cb3e158acc1c30d9de606af21c97ba851821e8b6ea535",
		InitPos:    0,
	}
	peer7 := &config.VBFTPeerStakeInfo{
		Index:      8,
		PeerPubkey: "0215865baab70607f4a2413a7a9ba95ab2c3c0202d5b7731c6824eef48e899fc90",
		InitPos:    5000,
	}
	peersinfo = append(peersinfo, peer1, peer2, peer3, peer4, peer5, peer6, peer7)
	conf.Peers = peersinfo
	return conf, nil
}

func TestGenConsensusPayload(t *testing.T) {
	log.InitLog(log.InfoLog, log.Stdout)
	config, err := constructConfig()
	if err != nil {
		t.Errorf("constructConfig failed:%s", err)
		return
	}
	res, err := genConsensusPayload(config, common.Uint256{}, 3)
	if err != nil {
		t.Errorf("test failed: %v", err)
	} else {
		t.Logf("config content: %v\n", res)
	}
}

func TestGenesisChainConfig(t *testing.T) {
	log.InitLog(log.InfoLog, log.Stdout)
	config, err := constructConfig()
	if err != nil {
		t.Errorf("constructConfig failed:%s", err)
		return
	}
	chainconfig, err := GenesisChainConfig(config, config.Peers, common.Uint256{}, 1)
	if err != nil {
		t.Errorf("TestGenesisChainConfig failed:%s", err)
		return
	}
	t.Logf("TestGenesisChainConfig succ: %v", chainconfig.PosTable)
}
