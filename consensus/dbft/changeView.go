package dbft

import (
	"io"
)

type ChangeView struct {
	msgData *ConsensusMessageData

	NewViewNumber byte
}


func (cv *ChangeView) Serialize(w io.Writer){
	cv.msgData.Serialize(w)
	w.Write([]byte{cv.NewViewNumber})
}

//read data to reader
func (cv *ChangeView) Deserialize(r io.Reader){
	cv.msgData.Deserialize(r)
	//TODO: NewViewNumber (readByte)
}

func (cv *ChangeView) Type() ConsensusMessageType{
	return ChangeViewMsg
}

func (cv *ChangeView) ViewNumber() byte{
	return cv.msgData.ViewNumber
}

func (cv *ChangeView) ConsensusMessageData() *ConsensusMessageData{
	return cv.msgData
}

