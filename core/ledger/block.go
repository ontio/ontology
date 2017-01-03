package ledger

import (
	tx "GoOnchain/core/transaction"
	"io"
	"GoOnchain/common/serialization"
	"GoOnchain/common"
)

type Block struct {
	Blockdata *Blockdata
	Transcations []*tx.Transaction

	hash *common.Uint256
}

func (b *Block) Serialize(w io.Writer)  {
	b.Blockdata.Serialize(w)
}

func (b *Block) Deserialize(r io.Reader) error  {
	b.Blockdata.Deserialize(r)

	//Transactions
	Len := serialization.ReadVarInt(r)
	for i := 0; i < Len; i++ {
		transaction := new(tx.Transaction)
		err := transaction.Deserialize(r)
		if err != nil {
			return err
		}
		b.Transcations = append(b.Transcations,transaction)
	}

	//TODO: merkleTree Compute Root

	return nil
}

func (b *Block) GetHash() common.Uint256  {

	if(b.hash == nil){
		//TODO: generate block hash
	}

	return *b.hash
}


