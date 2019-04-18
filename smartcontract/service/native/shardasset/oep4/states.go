package oep4

import (
	"fmt"
	"io"
	"math/big"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

//type AssetId uint64

type Oep4 struct {
	Name        string
	Symbol      string
	Decimals    uint64
	TotalSupply *big.Int
}

func (this *Oep4) Serialization(sink *common.ZeroCopySink) {
	sink.WriteString(this.Name)
	sink.WriteString(this.Symbol)
	sink.WriteUint64(this.Decimals)
	sink.WriteVarBytes(common.BigIntToNeoBytes(this.TotalSupply))
}

func (this *Oep4) Deserialization(source *common.ZeroCopySource) error {
	var irr, eof bool
	this.Name, _, irr, eof = source.NextString()
	if irr {
		return common.ErrIrregularData
	}
	if eof {
		return io.ErrUnexpectedEOF
	}
	this.Symbol, _, irr, eof = source.NextString()
	if irr {
		return common.ErrIrregularData
	}
	if eof {
		return io.ErrUnexpectedEOF
	}
	this.Decimals, eof = source.NextUint64()
	if eof {
		return io.ErrUnexpectedEOF
	}
	supply, _, irr, eof := source.NextVarBytes()
	if irr {
		return common.ErrIrregularData
	}
	if eof {
		return io.ErrUnexpectedEOF
	}
	this.TotalSupply = common.BigIntFromNeoBytes(supply)
	return nil
}

const (
	XSHARD_TRANSFER_PENDING  uint8 = 0x06
	XSHARD_TRANSFER_COMPLETE uint8 = 0x07
)

type XShardTransferState struct {
	ToShard   types.ShardID
	ToAccount common.Address
	Amount    *big.Int
	Status    uint8
}

func (this *XShardTransferState) Serialization(sink *common.ZeroCopySink) {
	utils.SerializationShardId(sink, this.ToShard)
	sink.WriteAddress(this.ToAccount)
	sink.WriteVarBytes(common.BigIntToNeoBytes(this.Amount))
	sink.WriteUint8(this.Status)
}

func (this *XShardTransferState) Deserialization(source *common.ZeroCopySource) error {
	var err error = nil
	this.ToShard, err = utils.DeserializationShardId(source)
	if err != nil {
		return fmt.Errorf("deserialization: read to shard failed, err: %s", err)
	}
	var irr, eof bool
	this.ToAccount, eof = source.NextAddress()
	if eof {
		return io.ErrUnexpectedEOF
	}
	amount, _, irr, eof := source.NextVarBytes()
	if irr {
		return common.ErrIrregularData
	}
	if eof {
		return io.ErrUnexpectedEOF
	}
	this.Amount = common.BigIntFromNeoBytes(amount)
	this.Status, eof = source.NextUint8()
	if eof {
		return io.ErrUnexpectedEOF
	}
	return nil
}
