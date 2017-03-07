package dbft

import (
	"io"
	ser "GoOnchain/common/serialization"
)

type ChangeView struct {
	msgData ConsensusMessageData
	NewViewNumber byte
}


func (cv *ChangeView) Serialize(w io.Writer)error{
	cv.msgData.Serialize(w)
	w.Write([]byte{cv.NewViewNumber})
	return nil
}

//read data to reader
func (cv *ChangeView) Deserialize(r io.Reader) error{
	 cv.msgData.Deserialize(r)
	viewNum,err := ser.ReadBytes(r,1)
	if err != nil {
		return err
	}
	cv.NewViewNumber = viewNum[0]
	return nil
}

func (cv *ChangeView) Type() ConsensusMessageType{
	return cv.ConsensusMessageData().Type
}

func (cv *ChangeView) ViewNumber() byte{
	return cv.msgData.ViewNumber
}

func (cv *ChangeView) ConsensusMessageData() *ConsensusMessageData{
	return &(cv.msgData)
}

