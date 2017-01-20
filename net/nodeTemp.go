package net

import (
	. "GoOnchain/common"
	tx "GoOnchain/core/transaction"
	pl "GoOnchain/net/payload"
	"GoOnchain/events"
)


func AllowHashes(hashes []Uint256){
	//TODO: AllowHashes
}

const (
	EventNewInventory events.EventType = iota
)

//TODO: "node” need change to "Node" (be public)
type Node struct {
	NodeEvent *events.Event
}

func (node *Node) init(){
	node.NodeEvent  = events.NewEvent()
}


func (node *Node) GetMemoryPool() map[Uint256]*tx.Transaction{
	//TODO: GetMemoryPool
	return nil
}


func (node *Node) SynchronizeMemoryPool(){
	//TODO: SynchronizeMemoryPool
}

func (node *Node) Relay(inventory pl.Inventory) error{
	//TODO: Relay
	return nil
}
