package message

import (
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/payload"
	"io"
)

type MetaDataEvent struct {
	Version  uint32
	Height   uint32
	MetaData *payload.MetaDataCode
}

type ContractEvent struct {
	Version       uint32
	DeployHeight  uint32
	Contract      *payload.DeployCode
	Destroyed     bool
	DestroyHeight uint32
}

func (this *ContractEvent) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint32(this.Version)
	sink.WriteUint32(this.DeployHeight)
	this.Contract.Serialization(sink)
	sink.WriteBool(this.Destroyed)
	sink.WriteUint32(this.DestroyHeight)
}

func (this *ContractEvent) Deserialization(source *common.ZeroCopySource) error {
	var eof bool
	this.Version, eof = source.NextUint32()
	this.DeployHeight, eof = source.NextUint32()
	if eof {
		return io.ErrUnexpectedEOF
	}
	this.Contract = &payload.DeployCode{}
	err := this.Contract.Deserialization(source)
	if err != nil {
		return err
	}
	var irr bool
	this.Destroyed, irr, eof = source.NextBool()
	if irr {
		return common.ErrIrregularData
	}
	if eof {
		return io.ErrUnexpectedEOF
	}
	this.DestroyHeight, eof = source.NextUint32()
	if eof {
		return io.ErrUnexpectedEOF
	}
	return nil
}
