package message

import (
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/payload"
)

type MetaDataEvent struct {
	Version  uint32
	Height   uint32
	MetaData *payload.MetaDataCode
}

type ContractEvent struct {
	Version  uint32
	Height   uint32
	Contract common.Address
}
