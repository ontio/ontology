package message

import (
	"github.com/Ontology/core/types"
	"github.com/Ontology/common"
	"github.com/Ontology/net/protocol"
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
	Node protocol.Noder
}

type NodeConsensusDisconnectMsg struct {
	Node protocol.Noder
}

type SmartCodeEventMsg struct {
	Event *types.SmartCodeEvent
}