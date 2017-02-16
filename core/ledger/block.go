package ledger

import (
	. "GoOnchain/common"
	"GoOnchain/common/serialization"
	"GoOnchain/core/contract/program"
	tx "GoOnchain/core/transaction"
	"GoOnchain/crypto"
	. "GoOnchain/errors"
	"io"
)

type Block struct {
	Blockdata    *Blockdata
	Transcations []*tx.Transaction

	hash *Uint256
}

func (b *Block) Serialize(w io.Writer) error {
	b.Blockdata.Serialize(w)
	err := serialization.WriteUint8(w, uint8(len(b.Transcations)))
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Block item Transcations length serialization failed.")
	}
	for _, transaction := range b.Transcations {
		temp := *transaction
		hash := temp.Hash()
		hash.Serialize(w)
	}
	return nil
}

func (b *Block) Deserialize(r io.Reader) error {
	if b.Blockdata == nil {
		b.Blockdata = new(Blockdata)
	}
	b.Blockdata.Deserialize(r)

	//Transactions
	var i uint8
	Len, err := serialization.ReadUint8(r)
	if err != nil {
		return err
	}
	var txhash Uint256
	var tharray []Uint256
	for i = 0; i < Len; i++ {
		txhash.Deserialize(r)
		transaction := new(tx.Transaction)
		transaction.SetHash(txhash)
		b.Transcations = append(b.Transcations, transaction)
		tharray = append(tharray, txhash)
	}

	b.Blockdata.TransactionsRoot, err = crypto.ComputeRoot(tharray)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Block Deserialize merkleTree compute failed")
	}

	return nil
}

func (b *Block) GetMessage() []byte {
	//TODO: GetMessage
	return []byte{}
}

func (b *Block) GetProgramHashes() ([]Uint160, error) {
	return nil, nil
}

func (b *Block) SetPrograms([]*program.Program) {

}

func (b *Block) GetPrograms() []*program.Program {
	return nil
}

func (b *Block) Hash() Uint256 {
	if b.hash == nil {
		b.hash = new(Uint256)
		*b.hash = b.Blockdata.Hash()
	}
	return *b.hash
}

func (b *Block) Verify() error {
	return nil
}

func (b *Block) Type() InventoryType {
	return BLOCK
}

func (bc *Blockchain) GetBlock(height uint32) (*Block, error) {
	temp := DefaultLedger.Store.GetBlockHash(height)
	bk, err := DefaultLedger.Store.GetBlock(temp.ToArray())
	if err != nil {
		return nil, err
	}
	return bk, nil
}
