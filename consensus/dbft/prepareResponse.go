package dbft

import (
	"io"
	ser "GoOnchain/common/serialization"
	. "GoOnchain/common"
)

type PrepareResponse struct {
	msgData ConsensusMessageData
	Signature []byte
}

func (pres *PrepareResponse) Serialize(w io.Writer){
	Trace()
	pres.msgData.Serialize(w)
	w.Write(pres.Signature)
}

//read data to reader
func (pres *PrepareResponse) Deserialize(r io.Reader) error{
	Trace()
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
	Trace()
	return pres.ConsensusMessageData().Type
}

func (pres *PrepareResponse) ViewNumber() byte{
	Trace()
	return pres.msgData.ViewNumber
}

func (pres *PrepareResponse) ConsensusMessageData() *ConsensusMessageData{
	Trace()
	return &(pres.msgData)
}
