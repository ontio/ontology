package dbft

import (
	"time"
	"sync"
	. "GoOnchain/errors"
	. "GoOnchain/common"
	"errors"
	"GoOnchain/net"
	pl "GoOnchain/net/payload"
	inv "GoOnchain/net/inventory"
	tx "GoOnchain/core/transaction"
	va "GoOnchain/core/validation"
	sig "GoOnchain/core/signature"
	ct "GoOnchain/core/contract"
	_ "GoOnchain/core/signature"
	"GoOnchain/core/ledger"
	"GoOnchain/consensus"
	cl "GoOnchain/client"
	"GoOnchain/events"
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
	localNode *net.Node

	newInventorySubscriber events.Subscriber
	blockPersistCompletedSubscriber events.Subscriber
}

func NewDbftService(localNode *net.Node,client *cl.Client,logDictionary string) *DbftService {
	return &DbftService{
		localNode: localNode,
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
		//ADD Log: "reject tx"
		ds.RequestChangeView()
		return errors.New("Transcation is invalid.")
	}

	ds.context.Transactions[TX.Hash()] = TX
	if len(ds.context.TransactionHashes) == len(ds.context.Transactions) {

		//Get Miner list
		txlist := ds.context.GetTransactionList()
		minerAddress := ledger.GetMinerAddress(ledger.DefaultLedger.Blockchain.GetMinersByTXs(txlist))

		if minerAddress == ds.context.NextMiner{
			//TODO: add log "send prepare response"
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

func (ds *DbftService) ChangeViewReceived(payload *pl.ConsensusPayload,message *ChangeView){
	//TODO: add log

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

		//TODO: add log "relay block"

		if err := ds.localNode.Relay(block); err != nil{
			//TODO: add log "reject block"
		}

		ds.context.State |= BlockSent

	}
	return nil
}

func (ds *DbftService) Halt() error  {
	if ds.timer != nil {
		ds.timer.Stop()
	}

	if ds.started {
		ledger.DefaultLedger.Blockchain.BCEvents.UnSubscribe(ledger.EventBlockPersistCompleted,ds.blockPersistCompletedSubscriber)
		ds.localNode.NodeEvent.UnSubscribe(net.EventNewInventory,ds.newInventorySubscriber)
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
			ds.Timeout() //TODO: double check timer check
		} else {
			time.AfterFunc(TimePerBlock-span,ds.Timeout)//TODO: double check time usage
		}
	} else {
		ds.context.State = Backup
		ds.timerHeight = ds.context.Height
		ds.timeView = viewNum
		//ds.timer.Reset()
	}
	return nil
}

func (ds *DbftService) LocalNodeNewInventory(v interface{}){
	if inventory,ok := v.(inv.Inventory);ok {
		if inventory.InvertoryType() == inv.Consensus {
			payload, isConsensusPayload := inventory.(*pl.ConsensusPayload)
			if isConsensusPayload {
				ds.NewConsensusPayload(payload)
			}
		} else if inventory.InvertoryType() == inv.Transaction  {
			transaction, isTransaction := inventory.(*tx.Transaction)
			if isTransaction{
				ds.NewTransactionPayload(transaction)
			}
		}
	}
}

func (ds *DbftService) NewConsensusPayload(payload *pl.ConsensusPayload){
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

func (ds *DbftService) PrepareRequestReceived(payload *pl.ConsensusPayload,message *PrepareRequest) {
	//TODO: add log

	if ds.context.State.HasFlag(Backup) || ds.context.State.HasFlag(RequestReceived) {
		return
	}

	if uint32(payload.MinerIndex) != ds.context.PrimaryIndex {return }

	prevBlockTimestamp := ledger.DefaultLedger.Blockchain.GetHeader(ds.context.PrevHash).Blockdata.Timestamp
	if payload.Timestamp <= prevBlockTimestamp || payload.Timestamp > uint32(time.Now().Add(time.Minute*10).Unix()){
		//TODO: add log "Timestamp incorrect"
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

	if err := ds.AddTransaction(message.MinerTransaction); err != nil {return }

	mempool := ds.localNode.GetMemoryPool()
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
		ds.localNode.SynchronizeMemoryPool()
	}
}

func (ds *DbftService) PrepareResponseReceived(payload *pl.ConsensusPayload,message *PrepareResponse){
	//TODO: add log

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
	consensus.DefaultPolicy.Refresh()
}

func  (ds *DbftService)  RequestChangeView() {
	ds.context.ExpectedView[ds.context.MinerIndex]++
	//TODO: add log request change view

	time.AfterFunc(SecondsPerBlock << (ds.context.ExpectedView[ds.context.MinerIndex]+1),ds.Timeout) //TODO: double check timer
	ds.SignAndRelay(ds.context.MakeChangeView())
	ds.CheckExpectedView(ds.context.ExpectedView[ds.context.MinerIndex])
}

func (ds *DbftService) SignAndRelay(payload *pl.ConsensusPayload){

	ctCxt := ct.NewContractContext(payload)

	ds.Client.Sign(ctCxt)
	ctCxt.Data.SetPrograms(ctCxt.GetPrograms())
	ds.localNode.Relay(payload)
}

func (ds *DbftService) Start() error  {

	ds.started = true

	ds.newInventorySubscriber = ledger.DefaultLedger.Blockchain.BCEvents.Subscribe(ledger.EventBlockPersistCompleted,ds.BlockPersistCompleted)
	ds.blockPersistCompletedSubscriber = ds.localNode.NodeEvent.Subscribe(net.EventNewInventory,ds.LocalNodeNewInventory)

	ds.InitializeConsensus(0)
	return nil
}

func (ds *DbftService) Timeout() {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if ds.timerHeight != ds.context.Height || ds.timeView != ds.context.ViewNumber {
		return
	}

	//TODO: add log "timeout”

	if ds.context.State.HasFlag(Primary) && !ds.context.State.HasFlag(RequestSent) {
		//TODO: add log “send prepare request”

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
			transactions := ds.localNode.GetMemoryPool() //TODO: add policy
			//Insert miner transaction
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
		ds.SignAndRelay(ds.context.MakePerpareRequest())
		time.AfterFunc(SecondsPerBlock << (ds.timeView + 1), ds.Timeout) //TODO: double check change timer

	} else if ds.context.State.HasFlag(Primary) && ds.context.State.HasFlag(RequestSent) || ds.context.State.HasFlag(Backup){
		ds.RequestChangeView()
	}
}
