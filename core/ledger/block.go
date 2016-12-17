package ledger

import (
	tx "GoOnchain/core/transaction"
	"GoOnchain/common"
	"GoOnchain/core/contract/program"
	"io"
)

type Block struct {
	Blockheader *Blockheader
	Transcations []*tx.Transaction
}

type Blockheader struct {
	//TODO: define the Blockheader struct(define new uinttype)
	Version uint
	PrevBlockHash  common.Uint256
	TransactionsRoot common.Uint256
	Timestamp uint
	Height uint
	nonce uint64
	Program *program.Program

	hash common.Uint256
}

//Serialize the blockheader
func (bh *Blockheader) Serialize(w io.Writer)  {
	bh.SerializeUnsigned(w)
	w.Write([]byte{byte(1)})
	bh.Program.Serialize(w)
}

//Serialize the blockheader data without program
func (bh *Blockheader) SerializeUnsigned(w io.Writer) error  {
	//TODO: implement blockheader SerializeUnsigned

	return nil
}


func (tx *Blockheader) GetProgramHashes() ([]common.Uint160, error){
	//TODO: implement blockheader GetProgramHashes

	return nil, nil
}