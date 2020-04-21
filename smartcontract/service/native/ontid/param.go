package ontid

import (
	"fmt"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

type SetKeyAccessParam struct {
	OntId     []byte
	SetIndex  uint32
	Access    string
	SignIndex uint32
	Proof     []byte
}

func (this *SetKeyAccessParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes(this.OntId)
	sink.WriteVarUint(uint64(this.SetIndex))
	sink.WriteString(this.Access)
	sink.WriteVarUint(uint64(this.SignIndex))
	sink.WriteVarBytes(this.Proof)
}

func (this *SetKeyAccessParam) Deserialization(source *common.ZeroCopySource) error {
	ontId, err := utils.DecodeVarBytes(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize ontId error: %v", err)
	}
	setIndex, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize setIndex error: %v", err)
	}
	access, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize access error: %v", err)
	}
	signIndex, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize signIndex error: %v", err)
	}
	Proof, err := utils.DecodeVarBytes(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize Creator error: %v", err)
	}
	this.OntId = ontId
	this.SetIndex = uint32(setIndex)
	this.Access = access
	this.SignIndex = uint32(signIndex)
	this.Proof = Proof
	return nil
}

type ProofParam struct {
	ProofType      string
	Created        string
	Creator        string
	SignatureValue string
}

func (this *ProofParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteString(this.ProofType)
	sink.WriteString(this.Created)
	sink.WriteString(this.Creator)
	sink.WriteString(this.SignatureValue)
}

func (this *ProofParam) Deserialization(source *common.ZeroCopySource) error {
	ProofType, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize ProofType error: %v", err)
	}
	Created, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize Created error: %v", err)
	}
	Creator, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize Creator error: %v", err)
	}
	SignatureValue, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize SignatureValue error: %v", err)
	}
	this.ProofType = ProofType
	this.Created = Created
	this.Creator = Creator
	this.SignatureValue = SignatureValue
	return nil
}

type ServiceParam struct {
	OntId       []byte
	ServiceId   []byte
	ServiceInfo []byte
	Index       uint32
	Proof       []byte
}

func (this *ServiceParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteBytes(this.OntId)
	sink.WriteBytes(this.ServiceId)
	sink.WriteBytes(this.ServiceInfo)
	sink.WriteUint32(this.Index)
	sink.WriteBytes(this.Proof)
}

func (this *ServiceParam) Deserialization(source *common.ZeroCopySource) error {
	OntId, err := utils.DecodeVarBytes(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize ProofType error: %v", err)
	}
	ServiceId, err := utils.DecodeVarBytes(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize Created error: %v", err)
	}
	ServiceInfo, err := utils.DecodeVarBytes(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize Creator error: %v", err)
	}
	Index, err := utils.DecodeUint32(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize SignatureValue error: %v", err)
	}
	Proof, err := utils.DecodeVarBytes(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize Creator error: %v", err)
	}
	this.OntId = OntId
	this.ServiceId = ServiceId
	this.ServiceInfo = ServiceInfo
	this.Index = Index
	this.Proof = Proof
	return nil
}
