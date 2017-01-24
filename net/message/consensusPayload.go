package message

import (
	. "GoOnchain/common"
	"GoOnchain/core/contract/program"
)

type ConsensusPayload struct {
	Version uint32
	PrevHash Uint256
	Height uint32
	MinerIndex uint16
	Timestamp uint32
	Data []byte
	Program *program.Program

	hash Uint256
}

func (cp *ConsensusPayload) Hash() Uint256{
	return Uint256{}
}

func (cp *ConsensusPayload) Verify() error{
	return nil
}

func (cp *ConsensusPayload) InvertoryType() InventoryType{
	return Consensus
}

func (cp *ConsensusPayload) GetProgramHashes() ([]Uint160, error){
	return nil,nil
}

func (cp *ConsensusPayload) SetPrograms([]*program.Program){
}

func (cp *ConsensusPayload) GetPrograms()  []*program.Program{
	return nil
}

func  (cp *ConsensusPayload) GetMessage() ([]byte){
	//TODO: GetMessage
	return  []byte{}
}


