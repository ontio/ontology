package db

import (
	"github.com/Ontology/common"
	"github.com/Ontology/common/serialization"
	tx "github.com/Ontology/core/types"

	"io"
)

type TransactionMeta struct {
	BlockHeight uint32
	Spend       *FixedBitMap
}

func NewTransactionMeta(height uint32, outputs uint32) TransactionMeta {
	return TransactionMeta{
		BlockHeight: height,
		Spend:       NewFixedBitMap(outputs),
	}
}

func (self *TransactionMeta) DenoteSpent(index uint32) {
	self.Spend.Set(index)
}

func (self *TransactionMeta) DenoteUnspent(index uint32) {
	self.Spend.Unset(index)
}

func (self *TransactionMeta) Height() uint32 {
	return self.BlockHeight
}
func (self *TransactionMeta) IsSpent(idx uint32) bool {
	return self.Spend.Get(idx)
}

func (self *TransactionMeta) IsFullSpent() bool {
	return self.Spend.IsFullSet()
}

func (self *TransactionMeta) Serialize(w io.Writer) error {
	err := serialization.WriteUint32(w, self.BlockHeight)
	if err != nil {
		return err
	}

	err = self.Spend.Serialize(w)

	return err
}

func (self *TransactionMeta) Deserialize(r io.Reader) error {
	height, err := serialization.ReadUint32(r)
	if err != nil {
		return err
	}
	self.BlockHeight = height
	self.Spend = &FixedBitMap{}
	err = self.Spend.Deserialize(r)
	return err
}

type TransactionProvider interface {
	BestStateProvider
	ContainTransaction(hash common.Uint256) bool
	GetTransactionBytes(hash common.Uint256) ([]byte, error)
	GetTransaction(hash common.Uint256) (*tx.Transaction, error)
}

type TransactionMetaProvider interface {
	BestStateProvider
	GetTransactionMeta(hash common.Uint256) (TransactionMeta, error)
}
