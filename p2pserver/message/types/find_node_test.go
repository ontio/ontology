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
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package types

import (
	"testing"

	"github.com/ontio/ontology/p2pserver/common"
)

func TestFindNodeRequest(t *testing.T) {
	var req FindNodeReq
	req.TargetID = common.PeerId{}

	MessageTest(t, &req)
}

func TestFindNodeResponse(t *testing.T) {
	var resp FindNodeResp
	resp.TargetID = common.PeerId{}
	resp.Address = "127.0.0.1:1222"
	id := common.PseudoPeerIdFromUint64(uint64(0x456))
	resp.CloserPeers = []common.PeerIDAddressPair{
		common.PeerIDAddressPair{
			ID:      id,
			Address: "127.0.0.1:4222",
		},
	}
	resp.Success = true

	MessageTest(t, &resp)
}
