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

func (this *MetaDataEvent) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint32(this.Version)
	sink.WriteUint32(this.Height)
	this.MetaData.Serialization(sink)
}

func (this *MetaDataEvent) Deserialization(source *common.ZeroCopySource) error {
	eof := false
	this.Version, eof = source.NextUint32()
	this.Height, eof = source.NextUint32()
	if eof {
		return io.ErrUnexpectedEOF
	}
	this.MetaData = payload.NewDefaultMetaData()
	return this.MetaData.Deserialization(source)
}
