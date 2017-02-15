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
func (pres *PrepareResponse) Deserialize(r io.Reader) error{
	err := pres.msgData.Deserialize(r)
	if err != nil {
		return err
	}
	pres.Signature,err = ser.ReadBytes(r,64)
	if err != nil {
		return err
	}
	return nil

}

func (pres *PrepareResponse) Type() ConsensusMessageType{
	return pres.ConsensusMessageData().Type
}

func (pres *PrepareResponse) ViewNumber() byte{
	return pres.msgData.ViewNumber
}

func (pres *PrepareResponse) ConsensusMessageData() *ConsensusMessageData{
	return pres.msgData
}