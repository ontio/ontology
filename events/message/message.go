package message

import (
	"github.com/Ontology/common"
	"github.com/Ontology/core/types"
	"github.com/Ontology/p2pserver/peer"
)

const (
	TopicSaveBlockComplete       = "svblkcmp"
	TopicNewInventory            = "newinv"
	TopicNodeDisconnect          = "noddis"
	TopicNodeConsensusDisconnect = "nodcnsdis"
	TopicSmartCodeEvent          = "scevt"
)

type SaveBlockCompleteMsg struct {
	Block *types.Block
}

type NewInventoryMsg struct {
	Inventory *common.Inventory
}

type NodeDisconnectMsg struct {
	Peer *peer.Peer
}

type NodeConsensusDisconnectMsg struct {
	Peer *peer.Peer
}

type SmartCodeEventMsg struct {
	Event *types.SmartCodeEvent
}
