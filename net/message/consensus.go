package message

import (
	"GoOnchain/common"
	"GoOnchain/core/contract/program"
	"GoOnchain/events"
	. "GoOnchain/net/protocol"
	"fmt"
)

type ConsensusPayload struct {
	Version    uint32
	PrevHash   common.Uint256
	Height     uint32
	MinerIndex uint16
	Timestamp  uint32
	Data       []byte
	Program    *program.Program

	hash common.Uint256
}

type consensus struct {
	msgHdr
	cons  ConsensusPayload
	event *events.Event
	//TBD
}

func (cp *ConsensusPayload) Hash() common.Uint256 {
	return common.Uint256{}
}

func (cp *ConsensusPayload) Verify() error {
	return nil
}

func (cp *ConsensusPayload) InvertoryType() common.InventoryType {
	return common.CONSENSUS
}

func (cp *ConsensusPayload) GetProgramHashes() ([]common.Uint160, error) {
	return nil, nil
}

func (cp *ConsensusPayload) SetPrograms([]*program.Program) {
}

func (cp *ConsensusPayload) GetPrograms() []*program.Program {
	return nil
}

func (cp *ConsensusPayload) GetMessage() []byte {
	//TODO: GetMessage
	return []byte{}
}

func (msg consensus) Handle(node Noder) error {
	common.Trace()
	fmt.Printf("RX Consensus message\n")
	if !node.ExistedID(msg.cons.hash) {
		if msg.event != nil {
			msg.event.Notify(events.EventNewInventory, msg.cons)
		}
	}
	return nil
}

func reqConsensusData(node Noder, hash common.Uint256) error {
	var msg dataReq
	msg.dataType = common.CONSENSUS
	// TODO handle the hash array case
	msg.hash = hash

	buf, _ := msg.Serialization()
	go node.Tx(buf)

	return nil
}
