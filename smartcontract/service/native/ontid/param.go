package ontid

import (
	"fmt"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

type RegIdWithPublicKeyParam struct {
	OntID  string
	PubKey string
	Access string
	Proof  string
}

func (this *RegIdWithPublicKeyParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteString(this.OntID)
	sink.WriteString(this.PubKey)
	sink.WriteString(this.Access)
	sink.WriteString(this.Proof)
}

func (this *RegIdWithPublicKeyParam) Deserialization(source *common.ZeroCopySource) error {
	ontID, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize ontID error: %v", err)
	}
	pubKey, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize pubKey error: %v", err)
	}
	access, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize access error: %v", err)
	}
	proof, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize proof error: %v", err)
	}
	this.OntID = ontID
	this.PubKey = pubKey
	this.Access = access
	this.Proof = proof
	return nil
}
