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
	localNet net.Neter

	newInventorySubscriber events.Subscriber
	blockPersistCompletedSubscriber events.Subscriber
}

func NewDbftService(client *cl.Client,logDictionary string,localNet net.Neter) *DbftService {
	Trace()
	return &DbftService{
		//localNode: localNode,
		Client: client,
		timer: time.NewTimer(time.Second*15),
		started: false,
		localNet:localNet,
		logDictionary: logDictionary,
	}
}

func (ds *DbftService) AddTransaction(TX *tx.Transaction) error{
	Trace()

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
		minerAddress,err := ledger.GetMinerAddress(ledger.DefaultLedger.Blockchain.GetMinersByTXs(txlist))
		if err != nil {
			return NewDetailErr(err,ErrNoCode,"[DbftService] ,GetMinerAddress failed")
		}

		if minerAddress == ds.context.NextMiner{
			con.Log("send perpare response")
			ds.context.State |= SignatureSent
			miner,err:=ds.Client.GetAccount(ds.context.Miners[ds.context.MinerIndex])
			if err != nil {
				return NewDetailErr(err,ErrNoCode,"[DbftService] ,GetAccount failed.")
			}
			sig.SignBySigner(ds.context.MakeHeader(),miner)
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
	Trace()
	ds.blockReceivedTime = time.Now()
	ds.InitializeConsensus(0)
}

func (ds *DbftService) CheckSignatures() error{
	Trace()
	if ds.context.GetSignaturesCount() >= ds.context.M() && ds.context.CheckTxHashesExist() {
		ep , err := ds.context.Miners[ds.context.MinerIndex].EncodePoint(true)
		if err != nil {
			return NewDetailErr(err,ErrNoCode,"[DbftService] ,EncodePoint failed")
		}
		codehash ,err := ToCodeHash(ep)
		if err != nil {
			return NewDetailErr(err,ErrNoCode,"[DbftService] ,ToCodeHash failed")
		}
		contract,err := ct.CreateMultiSigContract(codehash,ds.context.M(),ds.context.Miners)
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

		if err := ds.localNet.Xmit(block); err != nil{
			con.Log(fmt.Sprintf("reject block: %s", block.Hash()))
		}

		ds.context.State |= BlockSent

	}
	return nil
}

func (ds *DbftService) CreateBookkeepingTransaction(txs map[Uint256]*tx.Transaction,nonce uint64) *tx.Transaction {
	Trace()
	return &tx.Transaction{
		TxType: tx.BookKeeping,
	}
}

func (ds *DbftService) ChangeViewReceived(payload *msg.ConsensusPayload,message *ChangeView){
	Trace()
	con.Log(fmt.Sprintf("Change View Received: height=%d View=%d index=%d nv=%d",payload.Height,message.ViewNumber(),payload.MinerIndex,message.NewViewNumber))

	if message.NewViewNumber <= ds.context.ExpectedView[payload.MinerIndex] {
		return
	}

	ds.context.ExpectedView[payload.MinerIndex] = message.NewViewNumber
	ds.CheckExpectedView(message.NewViewNumber)
}

func (ds *DbftService) CheckExpectedView(viewNumber byte){
	Trace()
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
	Trace()
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
	Trace()
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if viewNum == 0 {
		ds.context.Reset(ds.Client)
	} else {
		ds.context.ChangeView(viewNum)
	}
	fmt.Println("ds.context.MinerIndex= ",ds.context.MinerIndex)
	if ds.context.MinerIndex < 0 {
		return NewDetailErr(errors.New("Miner Index incorrect"),ErrNoCode,"")
	}
	fmt.Println("ds.context.MinerIndex",ds.context.MinerIndex)
	fmt.Println("ds.context.PrimaryIndex",ds.context.PrimaryIndex)
	if ds.context.MinerIndex == int(ds.context.PrimaryIndex) {
		Trace()
		ds.context.State |= Primary
		ds.timerHeight = ds.context.Height
		ds.timeView = viewNum
		span := time.Now().Sub(ds.blockReceivedTime)
		Trace()
		if span > TimePerBlock {
			Trace()
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
	Trace()
	if inventory,ok := v.(Inventory);ok {
		if inventory.Type() == CONSENSUS {
			payload, isConsensusPayload := inventory.(*msg.ConsensusPayload)
			if isConsensusPayload {
				ds.NewConsensusPayload(payload)
			}
		} else if inventory.Type() == TRANSACTION  {
			transaction, isTransaction := inventory.(*tx.Transaction)
			if isTransaction{
				ds.NewTransactionPayload(transaction)
			}
		}
	}
}

func (ds *DbftService) NewConsensusPayload(payload *msg.ConsensusPayload){
	Trace()
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
	Trace()
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
	Trace()
	con.Log(fmt.Sprintf("Prepare Request Received: height=%d View=%d index=%d tx=%d",payload.Height,message.ViewNumber(),payload.MinerIndex,len(message.TransactionHashes)))

	if ds.context.State.HasFlag(Backup) || ds.context.State.HasFlag(RequestReceived) {
		return
	}

	if uint32(payload.MinerIndex) != ds.context.PrimaryIndex {return }
	header,_ := ledger.DefaultLedger.Blockchain.GetHeader(ds.context.PrevHash)
	/*
	* TODO Add Error Catch
	* 2017/2/27 luodanwg
	* */
	prevBlockTimestamp := header.Blockdata.Timestamp
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

	if _,err := va.VerifySignature(ds.context.MakeHeader(),ds.context.Miners[payload.MinerIndex],message.Signature); err != nil {
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
	Trace()

	con.Log(fmt.Sprintf("Prepare Response Received: height=%d View=%d index=%d",payload.Height,message.ViewNumber(),payload.MinerIndex))

	if ds.context.State.HasFlag(BlockSent)  {return}
	if ds.context.Signatures[payload.MinerIndex] != nil {return }

	header := ds.context.MakeHeader()
	if  header == nil {return }
	if _,err := va.VerifySignature(header,ds.context.Miners[payload.MinerIndex],message.Signature); err != nil {
		return
	}

	ds.context.Signatures[payload.MinerIndex] = message.Signature
	ds.CheckSignatures()
}

func  (ds *DbftService)  RefreshPolicy(){
	Trace()
	con.DefaultPolicy.Refresh()
}

func  (ds *DbftService)  RequestChangeView() {
	Trace()
	ds.context.ExpectedView[ds.context.MinerIndex]++
	con.Log(fmt.Sprintf("Request change view: height=%d View=%d nv=%d state=%d",ds.context.Height,ds.context.ViewNumber,ds.context.MinerIndex,ds.context.State))

	time.AfterFunc(SecondsPerBlock << (ds.context.ExpectedView[ds.context.MinerIndex]+1),ds.Timeout)
	ds.SignAndRelay(ds.context.MakeChangeView())
	ds.CheckExpectedView(ds.context.ExpectedView[ds.context.MinerIndex])
}

func (ds *DbftService) SignAndRelay(payload *msg.ConsensusPayload){
	Trace()
	ctCxt := ct.NewContractContext(payload)

	ds.Client.Sign(ctCxt)
	ctCxt.Data.SetPrograms(ctCxt.GetPrograms())
	ds.localNet.Xmit(payload)
}

func (ds *DbftService) Start() error  {
	Trace()
	ds.started = true

	ds.newInventorySubscriber = ledger.DefaultLedger.Blockchain.BCEvents.Subscribe(events.EventBlockPersistCompleted,ds.BlockPersistCompleted)
	ds.blockPersistCompletedSubscriber = ds.localNet.GetEvent("consensus").Subscribe(events.EventNewInventory,ds.LocalNodeNewInventory)

	ds.InitializeConsensus(0)
	return nil
}

func (ds *DbftService) Timeout() {
	Trace()
	//ds.mu.Lock()
	//defer ds.mu.Unlock()
	if ds.timerHeight != ds.context.Height || ds.timeView != ds.context.ViewNumber {
		return
	}
	fmt.Sprintf("Timeout: height=%d View=%d state=%d",ds.timerHeight,ds.timeView,ds.context.State)
	con.Log(fmt.Sprintf("Timeout: height=%d View=%d state=%d",ds.timerHeight,ds.timeView,ds.context.State))
	fmt.Println("ds.context.State.HasFlag(Primary)=",ds.context.State.HasFlag(Primary))
	fmt.Println("ds.context.State.HasFlag(RequestSent)=",ds.context.State.HasFlag(RequestSent))
	fmt.Println("ds.context.State.HasFlag(Backup)=",ds.context.State.HasFlag(Backup))

	if ds.context.State.HasFlag(Primary) && !ds.context.State.HasFlag(RequestSent) {
		con.Log(fmt.Sprintf("Send prepare request: height=%d View=%d",ds.timerHeight,ds.timeView,ds.context.State))
		ds.context.State |= RequestSent
		if !ds.context.State.HasFlag(SignatureSent) {

			//set context Timestamp
			now := uint32(time.Now().Unix())
			fmt.Println("ds.context.PrevHash",ds.context.PrevHash)
			header,_:= ledger.DefaultLedger.Blockchain.GetHeader(ds.context.PrevHash)
			fmt.Println(" ledger.DefaultLedger.Blockchain.GetHeader(ds.context.PrevHash)",header)
			/*
			* TODO Error Catch
			* 2017/2/27 luodanwg
			* */
			blockTime := header.Blockdata.Timestamp

			if blockTime > now {
				ds.context.Timestamp = blockTime
			} else {
				ds.context.Timestamp = now
			}

			ds.context.Nonce = GetNonce()
			transactions := ds.localNet.GetMemoryPool() //TODO: add policy

			txBookkeeping := ds.CreateBookkeepingTransaction(transactions,ds.context.Nonce)
			transactions[txBookkeeping.Hash()] = txBookkeeping

			if ds.context.TransactionHashes == nil {
				ds.context.TransactionHashes = []Uint256{}
			}

			ds.context.TransactionHashes = append(ds.context.TransactionHashes,txBookkeeping.Hash())
			for _,TX := range transactions {
				ds.context.TransactionHashes = append(ds.context.TransactionHashes,TX.Hash())
			}
			ds.context.Transactions = transactions

			txlist := ds.context.GetTransactionList()
			ds.context.NextMiner,_= ledger.GetMinerAddress(ledger.DefaultLedger.Blockchain.GetMinersByTXs(txlist))
			/*
			* TODO  add error catch
			* 2017/2/27 luodanwg
			* */
			block := ds.context.MakeHeader()
			account,_:= ds.Client.GetAccount(ds.context.Miners[ds.context.MinerIndex])
			/*
			* TODO add error catch
			* 2017/2/27 luodanwg
			* */
			ds.context.Signatures[ds.context.MinerIndex],_ = sig.SignBySigner(block,account)
			/*
			* TODO add error catch
			* 2017/2/27 luodanwg
			* */
		}
		ds.SignAndRelay(ds.context.MakePrepareRequest())
		time.AfterFunc(SecondsPerBlock << (ds.timeView + 1), ds.Timeout)

	} else if ds.context.State.HasFlag(Primary) && ds.context.State.HasFlag(RequestSent) || ds.context.State.HasFlag(Backup){
		ds.RequestChangeView()
	}
}
