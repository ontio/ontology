package dbft

import (
	"DNA/common/log"
	ser "DNA/common/serialization"
	tx "DNA/core/transaction"
	"bytes"
	"errors"
	"io"
)

type ConsensusMessage interface {
	ser.SerializableData
	Type() ConsensusMessageType
	ViewNumber() byte
	ConsensusMessageData() *ConsensusMessageData
}

type ConsensusMessageData struct {
	Type       ConsensusMessageType
	ViewNumber byte
}

func DeserializeMessage(data []byte) (ConsensusMessage, error) {
	log.Trace()
	msgType := ConsensusMessageType(data[0])

	r := bytes.NewReader(data)
	switch msgType {
	case PrepareRequestMsg:
		prMsg := &PrepareRequest{
			BookkeepingTransaction: new(tx.Transaction),
		}
		err := prMsg.Deserialize(r)
		if err != nil {
			log.Error("[DeserializeMessage] PrepareRequestMsg Deserialize Error: ", err.Error())
			return nil, err
		}
		return prMsg, nil

	case PrepareResponseMsg:
		presMsg := &PrepareResponse{}
		err := presMsg.Deserialize(r)
		if err != nil {
			log.Error("[DeserializeMessage] PrepareResponseMsg Deserialize Error: ", err.Error())
			return nil, err
		}
		return presMsg, nil
	case ChangeViewMsg:
		cv := &ChangeView{}
		err := cv.Deserialize(r)
		if err != nil {
			log.Error("[DeserializeMessage] ChangeViewMsg Deserialize Error: ", err.Error())
			return nil, err
		}
		return cv, nil

	}

	return nil, errors.New("The message is invalid.")
}

func (cd *ConsensusMessageData) Serialize(w io.Writer) {
	log.Trace()
	//ConsensusMessageType
	w.Write([]byte{byte(cd.Type)})

	//ViewNumber
	w.Write([]byte{byte(cd.ViewNumber)})

}

//read data to reader
func (cd *ConsensusMessageData) Deserialize(r io.Reader) error {
	log.Trace()
	//ConsensusMessageType
	var msgType [1]byte
	_, err := io.ReadFull(r, msgType[:])
	if err != nil {
		return err
	}
	cd.Type = ConsensusMessageType(msgType[0])

	//ViewNumber
	var vNumber [1]byte
	_, err = io.ReadFull(r, vNumber[:])
	if err != nil {
		return err
	}
	cd.ViewNumber = vNumber[0]

	return nil
}
