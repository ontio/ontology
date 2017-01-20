package dbft

import (
	"io"
	. "GoOnchain/common"
	ser "GoOnchain/common/serialization"
	tx "GoOnchain/core/transaction"
)

type PrepareRequest struct {
	msgData *ConsensusMessageData

	Nonce uint64
	NextMiner Uint160
	TransactionHashes []Uint256
	MinerTransaction *tx.Transaction
	Signature []byte
}

func (pr *PrepareRequest) Serialize(w io.Writer){
	pr.msgData.Serialize(w)
	ser.WriteVarUint(w,pr.Nonce)
	pr.NextMiner.Serialize(w)
	//TODO: TransactionHashes
	pr.MinerTransaction.Serialize(w)
	ser.WriteVarBytes(w,pr.Signature)
}

//read data to reader
func (pr *PrepareRequest) Deserialize(r io.Reader){

	pr.msgData.Deserialize(r)
	pr.Nonce,_ = ser.ReadVarUint(r,0)
	pr.NextMiner.Deserialize(r)
	pr.MinerTransaction.Deserialize(r)
	//TODO: TransactionHashes
	if pr.MinerTransaction.Hash() != pr.TransactionHashes[0]{
		return
		//TODO: add error
	}
	pr.Signature,_ = ser.ReadBytes(r,64)

}

func (pr *PrepareRequest) Type() ConsensusMessageType{
	return PrepareRequestMsg
}

func (pr *PrepareRequest) ViewNumber() byte{
	return pr.msgData.ViewNumber
}

func (pr *PrepareRequest) ConsensusMessageData() *ConsensusMessageData{
	return pr.msgData
}