package dbft

import (
	. "GoOnchain/common"
	"GoOnchain/crypto"
	tx "GoOnchain/core/transaction"
	 "GoOnchain/core/ledger"
	msg "GoOnchain/net/message"
	ser "GoOnchain/common/serialization"
	cl "GoOnchain/client"
	"fmt"
)

const ContextVersion uint32 = 0

type ConsensusContext struct {

	State ConsensusState
	PrevHash Uint256
	Height uint32
	ViewNumber byte
	Miners []*crypto.PubKey
	MinerIndex int
	PrimaryIndex uint32
	Timestamp uint32
	Nonce uint64
	NextMiner Uint160
	TransactionHashes []Uint256
	Transactions map[Uint256]*tx.Transaction
	Signatures [][]byte
	ExpectedView []byte

	txlist []*tx.Transaction

	header *ledger.Block
}

func (cxt *ConsensusContext)  M() int {
	Trace()
	return len(cxt.Miners) - (len(cxt.Miners) - 1) / 3
}

func NewConsensusContext() *ConsensusContext {
	Trace()
	return  &ConsensusContext{
	}
}

func (cxt *ConsensusContext)  ChangeView(viewNum byte)  {
	Trace()
	p := (cxt.Height - uint32(viewNum)) % uint32(len(cxt.Miners))
	cxt.State &= SignatureSent
	cxt.ViewNumber = viewNum
	if p >= 0 {
		cxt.PrimaryIndex = uint32(p)
	} else {
		cxt.PrimaryIndex = uint32(p) + uint32(len(cxt.Miners))
	}

	if cxt.State == Initial{
		cxt.TransactionHashes = nil
		cxt.Signatures = make([][]byte,len(cxt.Miners))
	}
	cxt.header = nil
}

func (cxt *ConsensusContext)  HasTxHash(txHash Uint256) bool {
	Trace()
	for _, hash :=  range cxt.TransactionHashes{
		if hash == txHash {
			return true
		}
	}
	return false
}

func (cxt *ConsensusContext)  MakeChangeView() *msg.ConsensusPayload {
	Trace()
	cv := &ChangeView{
		NewViewNumber: cxt.ExpectedView[cxt.MinerIndex],
	}
	cv.msgData.Type = ChangeViewMsg
	return cxt.MakePayload(cv)
}

func (cxt *ConsensusContext)  MakeHeader() *ledger.Block {
	Trace()
	if cxt.TransactionHashes == nil {
		return nil
	}

	txRoot,_ := crypto.ComputeRoot(cxt.TransactionHashes)


	if cxt.header == nil{
		blockData := &ledger.Blockdata{
			Version: ContextVersion,
			PrevBlockHash: cxt.PrevHash,
			TransactionsRoot: txRoot,
			Timestamp: cxt.Timestamp,
			Height: cxt.Height,
			ConsensusData: cxt.Nonce,
			NextMiner: cxt.NextMiner,
		}
		cxt.header = &ledger.Block{
			Blockdata: blockData,
			Transcations: []*tx.Transaction{},
		}
	}
	return cxt.header
}

func (cxt *ConsensusContext)  MakePayload(message ConsensusMessage) *msg.ConsensusPayload{
	Trace()
	message.ConsensusMessageData().ViewNumber = cxt.ViewNumber
	return &msg.ConsensusPayload{
		Version: ContextVersion,
		PrevHash: cxt.PrevHash,
		Height: cxt.Height,
		MinerIndex: uint16(cxt.MinerIndex),
		Timestamp: cxt.Timestamp,
		Data: ser.ToArray(message),
	}
}

func (cxt *ConsensusContext)  MakePrepareRequest() *msg.ConsensusPayload{
	Trace()
	fmt.Println("cxt.TransactionHashes[0]",cxt.TransactionHashes[0])
	preReq := &PrepareRequest{
		Nonce: cxt.Nonce,
		NextMiner: cxt.NextMiner,
		TransactionHashes: cxt.TransactionHashes,
		BookkeepingTransaction: cxt.Transactions[cxt.TransactionHashes[0]],
		Signature: cxt.Signatures[cxt.MinerIndex],
	}
	preReq.msgData.Type = PrepareRequestMsg
	return cxt.MakePayload(preReq)
}

func (cxt *ConsensusContext)  MakePerpareResponse(signature []byte) *msg.ConsensusPayload{
	Trace()
	preRes := &PrepareResponse{
		Signature: signature,
	}
	preRes.msgData.Type = PrepareResponseMsg
	return cxt.MakePayload(preRes)
}

func (cxt *ConsensusContext)  GetSignaturesCount() (count int){
	Trace()
	count = 0
	for _,sig := range cxt.Signatures {
		if sig != nil {
			count += 1
		}
	}
	return count
}

func (cxt *ConsensusContext)  GetTransactionList()  []*tx.Transaction{
	Trace()
	if cxt.txlist == nil{
		cxt.txlist = []*tx.Transaction{}
		fmt.Println("cxt.Transactions=",cxt.Transactions)
		for _,TX := range cxt.Transactions {
			cxt.txlist = append(cxt.txlist,TX)
			fmt.Println("transaction added to cxt.Transactions.",TX)
		}
		fmt.Println("len cxt.transacionts",len(cxt.Transactions))
	}
	return cxt.txlist
}

func (cxt *ConsensusContext)  GetTXByHashes()  []*tx.Transaction{
	Trace()
	TXs := []*tx.Transaction{}
	for _,hash := range cxt.TransactionHashes {
		if TX,ok:=cxt.Transactions[hash]; ok{
			TXs = append(TXs,TX)
		}
	}
	return TXs
}

func (cxt *ConsensusContext)  CheckTxHashesExist() bool {
	Trace()
	for _,hash := range cxt.TransactionHashes {
		if _,ok:=cxt.Transactions[hash]; !ok{
			return false
		}
	}
	return true
}

func (cxt *ConsensusContext) Reset(client *cl.Client){
	Trace()
	cxt.State = Initial
	cxt.PrevHash = ledger.DefaultLedger.Blockchain.CurrentBlockHash()
	cxt.Height = ledger.DefaultLedger.Blockchain.BlockHeight + 1
	cxt.ViewNumber = 0
	cxt.Miners = ledger.DefaultLedger.Blockchain.GetMiners()
	cxt.MinerIndex = -1

	minerLen := len(cxt.Miners)
	cxt.PrimaryIndex = cxt.Height % uint32(minerLen)
	cxt.TransactionHashes = nil
	cxt.Signatures = make([][]byte,minerLen)
	cxt.ExpectedView = make([]byte,minerLen)

	fmt.Println("minerLen=",minerLen)
	    for _, v := range cxt.Miners {
		    pubkey, _ := v.EncodePoint(true)
	           fmt.Println("Miners pub key =",pubkey)
	        }
	for i:=0;i<minerLen ;i++  {
		if client.ContainsAccount(cxt.Miners[i]){
			fmt.Println("Runed.")
			cxt.MinerIndex = i
			break
		}
	}
	fmt.Println("cxt.MinerIndex = ",cxt.MinerIndex)
	cxt.header = nil
}
