package dbft

import (
	"time"
	"sync"
	. "GoOnchain/errors"
	. "GoOnchain/common"
	"errors"
	"GoOnchain/net"
	msg "GoOnchain/net/message"
	tx "GoOnchain/core/transaction"
	va "GoOnchain/core/validation"
	sig "GoOnchain/core/signature"
	ct "GoOnchain/core/contract"
	_ "GoOnchain/core/signature"
	"GoOnchain/core/ledger"
	con "GoOnchain/consensus"
	cl "GoOnchain/client"
	"GoOnchain/events"
	"fmt"
)

const TimePerBlock = 15
const SecondsPerBlock = 15

type DbftService struct {
	context ConsensusContext
	mu           sync.Mutex
	Client *cl.Client
	timer *time.Timer
	timerHeight uint32
	timeView byte
	blockReceivedTime time.Time
	logDictionary string
	started bool
	localNet *net.Net

	newInventorySubscriber events.Subscriber
	blockPersistCompletedSubscriber events.Subscriber
}

func NewDbftService(client *cl.Client,logDictionary string) *DbftService {
	return &DbftService{
		//localNode: localNode,
		Client: client,
		timer: time.NewTimer(time.Second*15),
		started: false,
		logDictionary: logDictionary,
	}
}

func (ds *DbftService) AddTransaction(TX *tx.Transaction) error{

	hasTx := ledger.DefaultLedger.Blockchain.ContainsTransaction(TX.Hash())
	verifyTx := va.VerifyTransaction(TX,ledger.DefaultLedger,ds.context.GetTransactionList())
	checkPolicy :=  ds.CheckPolicy(TX)
	if hasTx || (verifyTx != nil) || (checkPolicy != nil) {

		con.Log(fmt.Sprintf("reject tx: %s",TX.Hash()))
		ds.RequestChangeView()
		return errors.New("Transcation is invalid.")
	}

	ds.context.Transactions[TX.Hash()] = TX
	if len(ds.context.TransactionHashes) == len(ds.context.Transactions) {

		//Get Miner list
		txlist := ds.context.GetTransactionList()
		minerAddress := ledger.GetMinerAddress(ledger.DefaultLedger.Blockchain.GetMinersByTXs(txlist))

		if minerAddress == ds.context.NextMiner{
			con.Log("send perpare response")
			ds.context.State |= SignatureSent
			sig.SignBySigner(ds.context.MakeHeader(),ds.Client.GetAccount(ds.context.Miners[ds.context.MinerIndex]))
			ds.SignAndRelay(ds.context.MakePerpareResponse(ds.context.Signatures[ds.context.MinerIndex]))
			ds.CheckSignatures()
		} else {
			ds.RequestChangeView()
			return errors.New("No valid Next Miner.")
		}
	}
	return nil
}

func (ds *DbftService) BlockPersistCompleted(v interface{}){
	ds.blockReceivedTime = time.Now()
	ds.InitializeConsensus(0)
}

func (ds *DbftService) CheckSignatures() error{

	if ds.context.GetSignaturesCount() >= ds.context.M() && ds.context.CheckTxHashesExist() {
		contract,err := ct.CreateMultiSigContract(ToCodeHash(ds.context.Miners[ds.context.MinerIndex].EncodePoint(true)),ds.context.M(),ds.context.Miners)
		if err != nil{
			return err
		}

		block := ds.context.MakeHeader()
		cxt := ct.NewContractContext(block)

		for i,j :=0,0; i < len(ds.context.Miners) && j < ds.context.M() ; i++ {
			if ds.context.Signatures[i] != nil{
				cxt.AddContract(contract,ds.context.Miners[i],ds.context.Signatures[i])
				j++
			}
		}

		cxt.Data.SetPrograms(cxt.GetPrograms())
		block.Transcations = ds.context.GetTXByHashes()

		con.Log(fmt.Sprintf("relay block: %s", block.Hash()))

		if err := ds.localNet.Relay(block); err != nil{
			con.Log(fmt.Sprintf("reject block: %s", block.Hash()))
		}

		ds.context.State |= BlockSent

	}
	return nil
}

func (ds *DbftService) CreateBookkeepingTransaction(txs map[Uint256]*tx.Transaction,nonce uint64) *tx.Transaction {
	return &tx.Transaction{
		TxType: tx.Bookkeeping,
	}
}

func (ds *DbftService) ChangeViewReceived(payload *msg.ConsensusPayload,message *ChangeView){
	con.Log(fmt.Sprintf("Change View Received: height=%d View=%d index=%d nv=%d",payload.Height,message.ViewNumber(),payload.MinerIndex,message.NewViewNumber))

	if message.NewViewNumber <= ds.context.ExpectedView[payload.MinerIndex] {
		return
	}

	ds.context.ExpectedView[payload.MinerIndex] = message.NewViewNumber
	ds.CheckExpectedView(message.NewViewNumber)
}

func (ds *DbftService) CheckExpectedView(viewNumber byte){
	if ds.context.ViewNumber == viewNumber {
		return
	}

	if len(ds.context.ExpectedView) >= ds.context.M(){
		ds.InitializeConsensus(viewNumber)
	}
}

func (ds *DbftService) CheckPolicy(transaction *tx.Transaction) error{
	//TODO: CheckPolicy

	return nil
}


func (ds *DbftService) Halt() error  {
	if ds.timer != nil {
		ds.timer.Stop()
	}

	if ds.started {
		ledger.DefaultLedger.Blockchain.BCEvents.UnSubscribe(events.EventBlockPersistCompleted,ds.blockPersistCompletedSubscriber)
		ds.localNet.GetEvent("consensus").UnSubscribe(events.EventNewInventory,ds.newInventorySubscriber)
	}
	return nil
}

func (ds *DbftService) InitializeConsensus(viewNum byte) error  {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if viewNum == 0 {
		ds.context.Reset(ds.Client)
	} else {
		ds.context.ChangeView(viewNum)
	}

	if ds.context.MinerIndex < 0 {
		return NewDetailErr(errors.New("Miner Index incorrect"),ErrNoCode,"")
	}

	if ds.context.MinerIndex == int(ds.context.PrimaryIndex) {
		ds.context.State |= Primary
		ds.timerHeight = ds.context.Height
		ds.timeView = viewNum

		span := time.Now().Sub(ds.blockReceivedTime)

		if span > TimePerBlock {
			ds.Timeout()
		} else {
			time.AfterFunc(TimePerBlock-span,ds.Timeout)
		}
	} else {
		ds.context.State = Backup
		ds.timerHeight = ds.context.Height
		ds.timeView = viewNum
	}
	return nil
}

func (ds *DbftService) LocalNodeNewInventory(v interface{}){
	if inventory,ok := v.(msg.Inventory);ok {
		if inventory.Type() == msg.Consensus {
			payload, isConsensusPayload := inventory.(*msg.ConsensusPayload)
			if isConsensusPayload {
				ds.NewConsensusPayload(payload)
			}
		} else if inventory.Type() == msg.Transaction  {
			transaction, isTransaction := inventory.(*tx.Transaction)
			if isTransaction{
				ds.NewTransactionPayload(transaction)
			}
		}
	}
}

func (ds *DbftService) NewConsensusPayload(payload *msg.ConsensusPayload){
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if int(payload.MinerIndex) == ds.context.MinerIndex {return }

	if payload.Version != ContextVersion || payload.PrevHash != ds.context.PrevHash || payload.Height != ds.context.Height {
		return
	}

	if int(payload.MinerIndex) >= len(ds.context.Miners) {return }

	message,_ := DeserializeMessage(payload.Data)

	if message.ViewNumber() != ds.context.ViewNumber && message.Type() != ChangeViewMsg {
		return
	}

	switch message.Type() {
	case ChangeViewMsg:
		if cv, ok := message.(*ChangeView); ok {
			ds.ChangeViewReceived(payload,cv)
		}
		break
	case PrepareRequestMsg:
		if pr, ok := message.(*PrepareRequest); ok {
			ds.PrepareRequestReceived(payload,pr)
		}
		break
	case PrepareResponseMsg:
		if pres, ok := message.(*PrepareResponse); ok {
			ds.PrepareResponseReceived(payload,pres)
		}
		break
	}
}

func (ds *DbftService) NewTransactionPayload(transaction *tx.Transaction) error{
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if ds.context.State.HasFlag(Backup) || !ds.context.State.HasFlag(RequestReceived) || ds.context.State.HasFlag(SignatureSent) {
		return NewDetailErr(errors.New("Consensus State is incorrect."),ErrNoCode,"")
	}

	if _, hasTx := ds.context.Transactions[transaction.Hash()]; hasTx {
		return NewDetailErr(errors.New("The transaction already exist."),ErrNoCode,"")
	}

	if !ds.context.HasTxHash(transaction.Hash()) {
		return NewDetailErr(errors.New("The transaction hash is not exist."),ErrNoCode,"")
	}
	return ds.AddTransaction(transaction)
}

func (ds *DbftService) PrepareRequestReceived(payload *msg.ConsensusPayload,message *PrepareRequest) {
	con.Log(fmt.Sprintf("Prepare Request Received: height=%d View=%d index=%d tx=%d",payload.Height,message.ViewNumber(),payload.MinerIndex,len(message.TransactionHashes)))

	if ds.context.State.HasFlag(Backup) || ds.context.State.HasFlag(RequestReceived) {
		return
	}

	if uint32(payload.MinerIndex) != ds.context.PrimaryIndex {return }

	prevBlockTimestamp := ledger.DefaultLedger.Blockchain.GetHeader(ds.context.PrevHash).Blockdata.Timestamp
	if payload.Timestamp <= prevBlockTimestamp || payload.Timestamp > uint32(time.Now().Add(time.Minute*10).Unix()){
		con.Log(fmt.Sprintf("Timestamp incorrect: %d",payload.Timestamp))
		return
	}

	ds.context.State |= RequestReceived
	ds.context.Timestamp = payload.Timestamp
	ds.context.Nonce = message.Nonce
	ds.context.NextMiner = message.NextMiner
	ds.context.TransactionHashes = message.TransactionHashes
	ds.context.Transactions = make(map[Uint256]*tx.Transaction)

	if err := va.VerifySignature(ds.context.MakeHeader(),ds.context.Miners[payload.MinerIndex],message.Signature); err != nil {
		return
	}

	minerLen := len(ds.context.Miners)
	ds.context.Signatures = make([][]byte,minerLen)
	ds.context.Signatures[payload.MinerIndex] = message.Signature

	if err := ds.AddTransaction(message.BookkeepingTransaction); err != nil {return }

	mempool :=  ds.localNet.GetMemoryPool()
	for _, hash := range ds.context.TransactionHashes[1:] {
		if transaction,ok := mempool[hash]; ok{
			if err := ds.AddTransaction(transaction); err != nil {
				return
			}
		}
	}

	//TODO: LocalNode allow hashes (add Except method)
	//AllowHashes(ds.context.TransactionHashes)

	if len(ds.context.Transactions) < len(ds.context.TransactionHashes){
		ds.localNet.SynchronizeMemoryPool()
	}
}

func (ds *DbftService) PrepareResponseReceived(payload *msg.ConsensusPayload,message *PrepareResponse){

	con.Log(fmt.Sprintf("Prepare Response Received: height=%d View=%d index=%d",payload.Height,message.ViewNumber(),payload.MinerIndex))

	if ds.context.State.HasFlag(BlockSent)  {return}
	if ds.context.Signatures[payload.MinerIndex] != nil {return }

	header := ds.context.MakeHeader()
	if  header == nil {return }
	if err := va.VerifySignature(header,ds.context.Miners[payload.MinerIndex],message.Signature); err != nil {
		return
	}

	ds.context.Signatures[payload.MinerIndex] = message.Signature
	ds.CheckSignatures()
}

func  (ds *DbftService)  RefreshPolicy(){
	con.DefaultPolicy.Refresh()
}

func  (ds *DbftService)  RequestChangeView() {
	ds.context.ExpectedView[ds.context.MinerIndex]++
	con.Log(fmt.Sprintf("Request change view: height=%d View=%d nv=%d state=%d",ds.context.Height,ds.context.ViewNumber,ds.context.MinerIndex,ds.context.State))

	time.AfterFunc(SecondsPerBlock << (ds.context.ExpectedView[ds.context.MinerIndex]+1),ds.Timeout)
	ds.SignAndRelay(ds.context.MakeChangeView())
	ds.CheckExpectedView(ds.context.ExpectedView[ds.context.MinerIndex])
}

func (ds *DbftService) SignAndRelay(payload *msg.ConsensusPayload){

	ctCxt := ct.NewContractContext(payload)

	ds.Client.Sign(ctCxt)
	ctCxt.Data.SetPrograms(ctCxt.GetPrograms())
	ds.localNet.Relay(payload)
}

func (ds *DbftService) Start() error  {

	ds.started = true

	ds.newInventorySubscriber = ledger.DefaultLedger.Blockchain.BCEvents.Subscribe(events.EventBlockPersistCompleted,ds.BlockPersistCompleted)
	ds.blockPersistCompletedSubscriber = ds.localNet.GetEvent("consensus").Subscribe(events.EventNewInventory,ds.LocalNodeNewInventory)

	ds.InitializeConsensus(0)
	return nil
}

func (ds *DbftService) Timeout() {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if ds.timerHeight != ds.context.Height || ds.timeView != ds.context.ViewNumber {
		return
	}

	con.Log(fmt.Sprintf("Timeout: height=%d View=%d state=%d",ds.timerHeight,ds.timeView,ds.context.State))

	if ds.context.State.HasFlag(Primary) && !ds.context.State.HasFlag(RequestSent) {

		con.Log(fmt.Sprintf("Send prepare request: height=%d View=%d",ds.timerHeight,ds.timeView,ds.context.State))
		ds.context.State |= RequestSent
		if !ds.context.State.HasFlag(SignatureSent) {

			//set context Timestamp
			now := uint32(time.Now().Unix())
			blockTime := ledger.DefaultLedger.Blockchain.GetHeader(ds.context.PrevHash).Blockdata.Timestamp

			if blockTime > now {
				ds.context.Timestamp = blockTime
			} else {
				ds.context.Timestamp = now
			}

			ds.context.Nonce = GetNonce()
			transactions := ds.localNet.GetMemoryPool() //TODO: add policy

			ds.CreateBookkeepingTransaction(transactions,ds.context.Nonce)

			if ds.context.TransactionHashes == nil {
				ds.context.TransactionHashes = []Uint256{}
			}

			for _,TX := range transactions {
				ds.context.TransactionHashes = append(ds.context.TransactionHashes,TX.Hash())
			}
			ds.context.Transactions = transactions

			txlist := ds.context.GetTransactionList()
			ds.context.NextMiner = ledger.GetMinerAddress(ledger.DefaultLedger.Blockchain.GetMinersByTXs(txlist))

			block := ds.context.MakeHeader()
			account := ds.Client.GetAccount(ds.context.Miners[ds.context.MinerIndex])
			ds.context.Signatures[ds.context.MinerIndex] = sig.SignBySigner(block,account)
		}
		ds.SignAndRelay(ds.context.MakePrepareRequest())
		time.AfterFunc(SecondsPerBlock << (ds.timeView + 1), ds.Timeout)

	} else if ds.context.State.HasFlag(Primary) && ds.context.State.HasFlag(RequestSent) || ds.context.State.HasFlag(Backup){
		ds.RequestChangeView()
	}
}
