package ledger

import (
	. "GoOnchain/common"
	"GoOnchain/common/serialization"
	tx "GoOnchain/core/transaction"
	. "GoOnchain/errors"
	"io"
	"GoOnchain/core/contract/program"
	pl "GoOnchain/net/payload"
)

type Block struct {
	Blockdata    *Blockdata
	Transcations []*tx.Transaction

	hash *Uint256
}

func (b *Block) Serialize(w io.Writer) error {
	b.Blockdata.Serialize(w)
	err := serialization.WriteVarUint(w, uint64(len(b.Transcations)))
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Block item Transcations length serialization failed.")
	}
	for _, transaction := range b.Transcations {
		transaction.Serialize(w)
	}
	return nil
}

func (b *Block) Deserialize(r io.Reader) error {
	//b.Blockdata.Deserialize(r)
	b.Blockdata.DeserializeUnsigned(r)

	//Transactions
	var i uint64
	Len, err := serialization.ReadVarUint(r, 0)
	if err != nil {
		return err
	}
	for i = 0; i < Len; i++ {
		transaction := new(tx.Transaction)
		err = transaction.Deserialize(r)
		if err != nil {
			return err
		}
		b.Transcations = append(b.Transcations, transaction)
	}

	//TODO: merkleTree Compute Root
	//Wjj:  crypto/ComputeRoot ?

	return nil
}

func (b *Block) GetHash() Uint256 {

	if b.hash == nil {
		//TODO: generate block hash
	}

	return *b.hash
}

func  (b *Block) GetMessage() ([]byte){
	//TODO: GetMessage
	return  []byte{}
}

func (b *Block) GetProgramHashes() ([]Uint160, error){
	return nil,nil
}

func (b *Block) SetPrograms([]*program.Program){

}

func (b *Block) GetPrograms()  []*program.Program{
	return nil
}

func (b *Block) Hash() Uint256{
	//TODO:  Hash()
	return Uint256{}
}

func (b *Block) Verify() error{
	return nil
}

func (b *Block) InvertoryType() pl.InventoryType{
	return pl.Block
}