package dbft

import (
	"io"
	. "GoOnchain/common"
	. "GoOnchain/errors"
	ser "GoOnchain/common/serialization"
	tx "GoOnchain/core/transaction"
	"fmt"
)

type PrepareRequest struct {
	msgData ConsensusMessageData

	Nonce uint64
	NextMiner Uint160
	TransactionHashes []Uint256
	BookkeepingTransaction *tx.Transaction
	Signature []byte
}

func (pr *PrepareRequest) Serialize(w io.Writer){
	Trace()
	pr.msgData.Serialize(w)
	ser.WriteVarUint(w,pr.Nonce)
	pr.NextMiner.Serialize(w)

	//Serialize  Transaction's hashes
	len := uint64(len(pr.TransactionHashes))
	ser.WriteVarUint(w, len)
	for _, txHash := range pr.TransactionHashes {
		txHash.Serialize(w)
	}

	//pr.BookkeepingTransaction.Serialize(w)
	ser.WriteVarBytes(w,pr.Signature)
}

//read data to reader
func (pr *PrepareRequest) Deserialize(r io.Reader) error{
	Trace()
	pr.msgData.Deserialize(r)
	pr.Nonce,_ = ser.ReadVarUint(r,0)
	pr.NextMiner.Deserialize(r)

	//TransactionHashes
	Len, err := ser.ReadVarUint(r, 0)
	if err != nil {
		return err
	}

	if (Len == 0) {
		fmt.Printf("The hash len at consensus payload is 0\n")
	} else {
		pr.TransactionHashes = make([]Uint256, Len)
		for i := uint64(0); i < Len; i++ {
			hash := new(Uint256)
			err = hash.Deserialize(r)
			if err != nil {
			return err
			}
			pr.TransactionHashes[i] = *hash
		}
		if pr.BookkeepingTransaction.Hash() != pr.TransactionHashes[0] {
			return  NewDetailErr(nil,ErrNoCode,"The Bookkeeping Transaction data is incorrect.")

		}
	}
	//pr.BookkeepingTransaction.Deserialize(r)
	pr.Signature,err = ser.ReadVarBytes(r)
	if err != nil {
		fmt.Printf("Parse the Signature error\n")
		return err
	}
	fmt.Println("Signature deserialize complete")

	return nil
}

func (pr *PrepareRequest) Type() ConsensusMessageType{
	Trace()
	return pr.ConsensusMessageData().Type
}

func (pr *PrepareRequest) ViewNumber() byte{
	Trace()
	return pr.msgData.ViewNumber
}

func (pr *PrepareRequest) ConsensusMessageData() *ConsensusMessageData{
	Trace()
	return &(pr.msgData)
}
