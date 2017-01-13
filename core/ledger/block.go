package ledger

import (
	tx "GoOnchain/core/transaction"
	"io"
	"GoOnchain/common/serialization"
	. "GoOnchain/common"
	. "GoOnchain/errors"
	"errors"
	"GoOnchain/crypto"
)

type Block struct {
	Blockdata *Blockdata
	Transcations []*tx.Transaction

	hash *Uint256
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

	txRoot,err := crypto.ComputeRoot(b.GetTransactionHashes());
	if err != nil{
		return err
	}

	if txRoot != b.Blockdata.TransactionsRoot { //TODO: change to compare
		return NewDetailErr(errors.New("Transaction Root is incorrect."),ErrNoCode,"")
	}

	return nil
}

func (b *Block) GetTransactionHashes() []Uint256 {
	//TODO: implement GetTransactionHashes
	return nil
}

func (b *Block) GetHash() Uint256  {

	if(b.hash == nil){
		//TODO: generate block hash
	}

	return *b.hash
}


