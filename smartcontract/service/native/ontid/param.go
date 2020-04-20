package ontid

import (
	"fmt"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

type RegIdWithPublicKeyParam struct {
	OntID  []byte
	PubKey []byte
	Access string
	Proof  []byte
}

func (this *RegIdWithPublicKeyParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes(this.OntID)
	sink.WriteVarBytes(this.PubKey)
	sink.WriteString(this.Access)
	sink.WriteVarBytes(this.Proof)
}

func (this *RegIdWithPublicKeyParam) Deserialization(source *common.ZeroCopySource) error {
	ontID, err := utils.DecodeVarBytes(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarBytes, deserialize ontID error: %v", err)
	}
	pubKey, err := utils.DecodeVarBytes(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarBytes, deserialize pubKey error: %v", err)
	}
	access, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeString, deserialize access error: %v", err)
	}
	proof, err := utils.DecodeVarBytes(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarBytes, deserialize proof error: %v", err)
	}
	this.OntID = ontID
	this.PubKey = pubKey
	this.Access = access
	this.Proof = proof
	return nil
}

type OldRegIdWithPublicKeyParam struct {
	OntID  []byte
	PubKey []byte
}

func (this *OldRegIdWithPublicKeyParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes(this.OntID)
	sink.WriteVarBytes(this.PubKey)
}

func (this *OldRegIdWithPublicKeyParam) Deserialization(source *common.ZeroCopySource) error {
	ontID, err := utils.DecodeVarBytes(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarBytes, deserialize ontID error: %v", err)
	}
	pubKey, err := utils.DecodeVarBytes(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarBytes, deserialize pubKey error: %v", err)
	}
	this.OntID = ontID
	this.PubKey = pubKey
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
