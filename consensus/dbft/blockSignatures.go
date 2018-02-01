package dbft

import (
	"errors"
	ser "github.com/Ontology/common/serialization"
	"io"
)

type BlockSignatures struct {
	msgData    ConsensusMessageData
	Signatures []SignaturesData
}

type SignaturesData struct {
	Signature []byte
	Index     uint16
}

func (self *BlockSignatures) Serialize(w io.Writer) error {
	self.msgData.Serialize(w)
	if err := ser.WriteVarUint(w, uint64(len(self.Signatures))); err != nil {
		return errors.New("[BlockSignatures] serialization failed")
	}

	for i := 0; i < len(self.Signatures); i++ {
		if err := ser.WriteVarBytes(w, self.Signatures[i].Signature); err != nil {
			return errors.New("[BlockSignatures] serialization sig failed")
		}
		if err := ser.WriteUint16(w, self.Signatures[i].Index); err != nil {
			return errors.New("[BlockSignatures] serialization sig index failed")
		}
	}

	return nil
}

func (self *BlockSignatures) Deserialize(r io.Reader) error {
	err := self.msgData.Deserialize(r)
	if err != nil {
		return err
	}

	length, _ := ser.ReadVarUint(r, 0)
	self.Signatures = make([]SignaturesData, length)

	for i := uint64(0); i < length; i++ {
		self.Signatures[i].Signature, err = ser.ReadVarBytes(r)
		if err != nil {
			return err
		}
		self.Signatures[i].Index, err = ser.ReadUint16(r)
		if err != nil {
			return err
		}
	}

	return nil
}

func (self *BlockSignatures) Type() ConsensusMessageType {
	return self.ConsensusMessageData().Type
}

func (self *BlockSignatures) ViewNumber() byte {
	return self.msgData.ViewNumber
}

func (self *BlockSignatures) ConsensusMessageData() *ConsensusMessageData {
	return &(self.msgData)
}
