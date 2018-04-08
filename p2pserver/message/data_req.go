package message

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/Ontology/common"
	"github.com/Ontology/common/log"
)

type DataReq struct {
	MsgHdr
	DataType common.InventoryType
	Hash     common.Uint256
}

func (msg DataReq) Serialization() ([]byte, error) {
	hdrBuf, err := msg.MsgHdr.Serialization()
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(hdrBuf)
	err = binary.Write(buf, binary.LittleEndian, msg.DataType)
	if err != nil {
		return nil, err
	}
	msg.Hash.Serialize(buf)
	return buf.Bytes(), err
}

func (msg *DataReq) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, &(msg.MsgHdr))
	if err != nil {
		log.Warn("Parse dataReq message hdr error")
		return errors.New("Parse dataReq message hdr error ")
	}

	err = binary.Read(buf, binary.LittleEndian, &(msg.DataType))
	if err != nil {
		log.Warn("Parse dataReq message dataType error")
		return errors.New("Parse dataReq message dataType error ")
	}

	err = msg.Hash.Deserialize(buf)
	if err != nil {
		log.Warn("Parse dataReq message hash error")
		return errors.New("Parse dataReq message hash error ")
	}
	return nil
}
