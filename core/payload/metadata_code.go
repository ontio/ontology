package payload

import (
	"io"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
)

type MetaDataCode struct {
	OntVersion uint64
	Contract   common.Address
	Owner      common.Address
	AllShard   bool
	IsFrozen   bool
	ShardId    uint64
}

func NewDefaultMetaData() *MetaDataCode {
	return &MetaDataCode{
		OntVersion: common.VERSION_SUPPORT_SHARD,
		Contract:   common.ADDRESS_EMPTY,
		Owner:      common.ADDRESS_EMPTY,
		AllShard:   false,
		IsFrozen:   false,
		ShardId:    0,
	}
}

func (this *MetaDataCode) Serialize(w io.Writer) error {
	sink := common.NewZeroCopySink(0)
	this.Serialization(sink)
	err := serialization.WriteVarBytes(w, sink.Bytes())
	return err
}
func (this *MetaDataCode) Deserialize(r io.Reader) error {
	data, err := serialization.ReadVarBytes(r)
	if err != nil {
		return err
	}
	source := common.NewZeroCopySource(data)
	err = this.Deserialization(source)
	return err
}

func (this *MetaDataCode) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint64(this.OntVersion)
	sink.WriteAddress(this.Contract)
	sink.WriteAddress(this.Owner)
	sink.WriteBool(this.AllShard)
	sink.WriteBool(this.IsFrozen)
	sink.WriteUint64(this.ShardId)
}

func (this *MetaDataCode) Deserialization(source *common.ZeroCopySource) error {
	var irr, eof bool
	this.OntVersion, eof = source.NextUint64()
	this.Contract, eof = source.NextAddress()
	this.Owner, eof = source.NextAddress()
	if eof {
		return io.ErrUnexpectedEOF
	}
	this.AllShard, irr, eof = source.NextBool()
	if irr {
		return common.ErrIrregularData
	}

	if eof {
		return io.ErrUnexpectedEOF
	}
	this.IsFrozen, irr, eof = source.NextBool()
	if irr {
		return common.ErrIrregularData
	}

	if eof {
		return io.ErrUnexpectedEOF
	}
	this.ShardId, eof = source.NextUint64()
	if eof {
		return io.ErrUnexpectedEOF
	}
	return nil
}
