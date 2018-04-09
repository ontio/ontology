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

package server

import (
	types "github.com/Ontology/p2pserver/common"
)

type StartServerReq struct {
	StartSync bool
}
type StartServerRsp struct {
	Error error
}

type StopServerReq struct {
}
type StopServerRsp struct {
	Error error
}

type GetVersionReq struct {
}
type GetVersionRsp struct {
	Version uint32
}

type GetConnectionCntReq struct {
}
type GetConnectionCntRsp struct {
	Cnt uint32
}

type GetIdReq struct {
}
type GetIdRsp struct {
	Id uint64
}

type GetSyncPortReq struct {
}
type GetSyncPortRsp struct {
	SyncPort uint16
}

type GetConsPortReq struct {
}
type GetConsPortRsp struct {
	ConsPort uint16
}

type GetPortReq struct {
}
type GetPortRsp struct {
	SyncPort uint16
	ConsPort uint16
}

type GetConnectionStateReq struct {
}
type GetConnectionStateRsp struct {
	State uint32
}

type GetTimeReq struct {
}
type GetTimeRsp struct {
	Time int64
}

type GetNodeTypeReq struct {
}
type GetNodeTypeRsp struct {
	NodeType uint64
}

type GetRelayStateReq struct {
}
type GetRelayStateRsp struct {
	Relay bool
}

type GetNeighborAddrsReq struct {
}
type GetNeighborAddrsRsp struct {
	Addrs []types.PeerAddr
	Count uint64
}

type IsSyncingReq struct {
}
type IsSyncingRsp struct {
	IsSyncing bool
}
