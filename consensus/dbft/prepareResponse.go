package dbft

import (
	"DNA/common/log"
	ser "DNA/common/serialization"
	"io"
)

type PrepareResponse struct {
	msgData   ConsensusMessageData
	Signature []byte
}

func (pres *PrepareResponse) Serialize(w io.Writer) error {
	log.Trace()
	pres.msgData.Serialize(w)
	w.Write(pres.Signature)
	return nil
}

//read data to reader
func (pres *PrepareResponse) Deserialize(r io.Reader) error {
	log.Trace()
	err := pres.msgData.Deserialize(r)
	if err != nil {
		return err
	}
	// Fixme the 64 should be defined as a unified const
	pres.Signature, err = ser.ReadBytes(r, 64)
	if err != nil {
		return err
	}
	return nil

}

func (pres *PrepareResponse) Type() ConsensusMessageType {
	log.Trace()
	return pres.ConsensusMessageData().Type
}

func (pres *PrepareResponse) ViewNumber() byte {
	log.Trace()
	return pres.msgData.ViewNumber
}

func (pres *PrepareResponse) ConsensusMessageData() *ConsensusMessageData {
	log.Trace()
	return &(pres.msgData)
}
