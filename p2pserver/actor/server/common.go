package server

import (
	msg "github.com/Ontology/p2pserver/message"
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
	Addrs []msg.PeerAddr
	Count uint64
}

type IsSyncingReq struct {
}
type IsSyncingRsp struct {
	IsSyncing bool
}
