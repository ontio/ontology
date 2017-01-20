package dbft

import (
	"io"
	ser "GoOnchain/common/serialization"
)

type PrepareResponse struct {
	msgData *ConsensusMessageData
	Signature []byte
}

func (pres *PrepareResponse) Serialize(w io.Writer){
	pres.msgData.Serialize(w)
	w.Write(pres.Signature)
}

//read data to reader
func (pres *PrepareResponse) Deserialize(r io.Reader){
	pres.msgData.Deserialize(r)
	pres.Signature,_ = ser.ReadBytes(r,64)

}

func (pres *PrepareResponse) Type() ConsensusMessageType{
	return PrepareResponseMsg
}

func (pres *PrepareResponse) ViewNumber() byte{
	return pres.msgData.ViewNumber
}

func (pres *PrepareResponse) ConsensusMessageData() *ConsensusMessageData{
	return pres.msgData
}