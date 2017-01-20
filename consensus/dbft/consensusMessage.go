package dbft

import (
	ser "GoOnchain/common/serialization"
	"io"
)


type ConsensusMessage interface {
	ser.SerializableData
	Type() ConsensusMessageType
	ViewNumber() byte
	ConsensusMessageData() *ConsensusMessageData
}

type ConsensusMessageData struct {
	ViewNumber byte
}

func DeserializeMessage(data []byte) (ConsensusMessage, error){
	return nil,nil
}

func (cd *ConsensusMessageData) Serialize(w io.Writer){
	//TODO: Serialize

}

//read data to reader
func (cd *ConsensusMessageData) Deserialize(r io.Reader){
	//TODO： Deserialize
}

