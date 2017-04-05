package dbft

import (
	. "DNA/common"
	"DNA/common/log"
	ser "DNA/common/serialization"
	tx "DNA/core/transaction"
	. "DNA/errors"
	"fmt"
	"io"
)

type PrepareRequest struct {
	msgData ConsensusMessageData

	Nonce                  uint64
	NextMiner              Uint160
	TransactionHashes      []Uint256
	BookkeepingTransaction *tx.Transaction
	Signature              []byte
}

func (pr *PrepareRequest) Serialize(w io.Writer) error {
	log.Trace()
	pr.msgData.Serialize(w)
	err := ser.WriteVarUint(w, pr.Nonce)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "PrepareRequest Execute WriteVarUint failed.")
	}
	_, err = pr.NextMiner.Serialize(w)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "PrepareRequest Execute NextMiner.Serialize failed.")
	}
	//Serialize  Transaction's hashes
	length := uint64(len(pr.TransactionHashes))
	err = ser.WriteVarUint(w, length)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "PrepareRequest Execute WriteVarUint. failed.")
	}
	for _, txHash := range pr.TransactionHashes {
		_, err = txHash.Serialize(w)
		if err != nil {
			return NewDetailErr(err, ErrNoCode, "PrepareRequest Execute txHash.Serialize. failed.")
		}
	}

	err = pr.BookkeepingTransaction.Serialize(w)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "PrepareRequest Execute BookkeepingTransaction Serialize failed.")
	}
	//* test use
	//buf :=bytes.NewBuffer([]byte{})
	//err =pr.BookkeepingTransaction.Serialize(buf)
	//if err != nil {
	//	return NewDetailErr(err, ErrNoCode, "PrepareRequest Execute BookkeepingTransaction Serialize failed.")
	//}
	//fmt.Println("PrepareRequest Serialize BookkeepingTransaction=",buf.Bytes() )
	//fmt.Println("TxType",pr.BookkeepingTransaction.TxType         )
	//fmt.Println("PayloadVersion",pr.BookkeepingTransaction.PayloadVersion)
	//fmt.Println(pr.BookkeepingTransaction.Payload        )
	//fmt.Println(pr.BookkeepingTransaction.Nonce          )
	//for _, v := range pr.BookkeepingTransaction.Attributes {
	//	fmt.Println("Attributes",v)
	//}
	//for _, v := range pr.BookkeepingTransaction.UTXOInputs {
	//	fmt.Println("UTXOInputs",v)
	//}
	//for _, v := range pr.BookkeepingTransaction.BalanceInputs {
	//	fmt.Println("BalanceInputs",v)
	//}
	//for _, v := range pr.BookkeepingTransaction.Outputs {
	//	fmt.Println("Outputs",v)
	//}
	//for _, v := range pr.BookkeepingTransaction.Programs {
	//	fmt.Println("Programs",v)
	//}
	err = ser.WriteVarBytes(w, pr.Signature)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "PrepareRequest Execute ser.WriteVarBytes failed.")
	}
	return nil
}

//read data to reader
func (pr *PrepareRequest) Deserialize(r io.Reader) error {
	log.Trace()
	pr.msgData = ConsensusMessageData{}
	pr.msgData.Deserialize(r)
	pr.Nonce, _ = ser.ReadVarUint(r, 0)
	pr.NextMiner = Uint160{}
	pr.NextMiner.Deserialize(r)

	//TransactionHashes
	Len, err := ser.ReadVarUint(r, 0)
	if err != nil {
		return err
	}

	if Len == 0 {
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
	}
	pr.BookkeepingTransaction.Deserialize(r)
	if pr.BookkeepingTransaction.Hash() != pr.TransactionHashes[0] {
		log.Debug("pr.BookkeepingTransaction.Hash()=", pr.BookkeepingTransaction.Hash())
		log.Debug("pr.TransactionHashes[0]=", pr.TransactionHashes[0])
		//buf :=bytes.NewBuffer([]byte{})
		//pr.BookkeepingTransaction.Serialize(buf)
		//fmt.Println("PrepareRequest Deserialize cxt.Transactions[cxt.TransactionHashes[0]=",buf.Bytes() )
		//fmt.Println("TxType",pr.BookkeepingTransaction.TxType         )
		//fmt.Println("PayloadVersion",pr.BookkeepingTransaction.PayloadVersion)
		//fmt.Println(pr.BookkeepingTransaction.Payload        )
		//fmt.Println(pr.BookkeepingTransaction.Nonce          )
		//for _, v := range pr.BookkeepingTransaction.Attributes {
		//	fmt.Println("Attributes",v)
		//}
		//for _, v := range pr.BookkeepingTransaction.UTXOInputs {
		//	fmt.Println("UTXOInputs",v)
		//}
		//for _, v := range pr.BookkeepingTransaction.BalanceInputs {
		//	fmt.Println("BalanceInputs",v)
		//}
		//for _, v := range pr.BookkeepingTransaction.Outputs {
		//	fmt.Println("Outputs",v)
		//}
		//for _, v := range pr.BookkeepingTransaction.Programs {
		//	fmt.Println("Programs",v)
		//}

		return NewDetailErr(nil, ErrNoCode, "The Bookkeeping Transaction data is incorrect.")

	}
	pr.Signature, err = ser.ReadVarBytes(r)
	if err != nil {
		fmt.Printf("Parse the Signature error\n")
		return err
	}

	return nil
}

func (pr *PrepareRequest) Type() ConsensusMessageType {
	log.Trace()
	return pr.ConsensusMessageData().Type
}

func (pr *PrepareRequest) ViewNumber() byte {
	log.Trace()
	return pr.msgData.ViewNumber
}

func (pr *PrepareRequest) ConsensusMessageData() *ConsensusMessageData {
	log.Trace()
	return &(pr.msgData)
}
