package dbft

import (
	"time"
	"sync"
	. "GoOnchain/errors"
	. "GoOnchain/common"
	"GoOnchain/common/log"
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
	"GoOnchain/core/transaction/payload"
	"GoOnchain/core/contract/program"
)

const (
	TimePerBlock = (15 * time.Second)
	SecondsPerBlock = (15 * time.Second)
)

type DbftService struct {
	context ConsensusContext
	mu           sync.Mutex
	Client cl.Client
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

func NewDbftService(client cl.Client,logDictionary string,localNet net.Neter) *DbftService {
	Trace()

	ds := &DbftService{
		//localNode: localNode,
		Client: client,
		timer: time.NewTimer(time.Second * 15),
		started: false,
		localNet:localNet,
		logDictionary: logDictionary,
	}

	if !ds.timer.Stop() {
		<-ds.timer.C
	}
	Trace()
	go ds.timerRoutine()
	return ds
}

func (ds *DbftService) AddTransaction(TX *tx.Transaction) error{
	Trace()

	hasTx := ledger.DefaultLedger.Blockchain.ContainsTransaction(TX.Hash())
	verifyTx := va.VerifyTransaction(TX,ledger.DefaultLedger,ds.context.GetTransactionList())
	checkPolicy :=  ds.CheckPolicy(TX)

	log.Debug("The return value is HasTx: ", hasTx, " verifyTX: ", verifyTx, " checkPolicy: ", checkPolicy)
	if hasTx || (verifyTx != nil) || (checkPolicy != nil) {
		log.Warn(fmt.Sprintf("Reject tx: %v", TX.Hash()))
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

		if minerAddress == ds.context.NextMiner {
			log.Debug("Send prepare response")
			ds.context.State |= SignatureSent
			miner,err:=ds.Client.GetAccount(ds.context.Miners[ds.context.MinerIndex])
			if err != nil {
				return NewDetailErr(err,ErrNoCode,"[DbftService] ,GetAccount failed.")
			}
			//sig.SignBySigner(ds.context.MakeHeader(), miner)
			ds.context.Signatures[ds.context.MinerIndex], err = sig.SignBySigner(ds.context.MakeHeader(), miner)
			if err != nil {
				log.Error("[DbftService], SignBySigner failed.")
				return NewDetailErr(err,ErrNoCode,"[DbftService], SignBySigner failed.")
			}
			payload := ds.context.MakePrepareResponse(ds.context.Signatures[ds.context.MinerIndex])
			ds.SignAndRelay(payload)
			err =ds.CheckSignatures()
			if err != nil {
				return NewDetailErr(err,ErrNoCode,"[DbftService] ,CheckSignatures failed.")
			}
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
	go ds.InitializeConsensus(0)
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
				err:=cxt.AddContract(contract,ds.context.Miners[i],ds.context.Signatures[i])
				if err != nil {
					return NewDetailErr(err, ErrNoCode, "[DbftService], CheckSignatures AddContract failed.")
				}
				j++
			}
		}
		cxt.Data.SetPrograms(cxt.GetPrograms())
		block.Transcations = ds.context.GetTXByHashes()

		log.Info(fmt.Sprintf("cxt.GetPrograms(): %d", cxt.GetPrograms()))
		log.Info(fmt.Sprintf("cxt.Data.GetPrograms(): %d", cxt.Data.GetPrograms()))
		log.Info(fmt.Sprintf("relay block: %d", block.Hash()))

		if err := ds.localNet.Xmit(block); err != nil{
			log.Warn(fmt.Sprintf("Reject block: %s", block.Hash()))
		}
		Trace()
		ds.context.State |= BlockSent

	}
	return nil
}

func (ds *DbftService) CreateBookkeepingTransaction(nonce uint64) *tx.Transaction {
	Trace()
	return &tx.Transaction{
		TxType: tx.BookKeeping,
		PayloadVersion: 0x2,
		Payload: &payload.MinerPayload{},
		Nonce: nonce,
		Attributes: []*tx.TxAttribute{},
		UTXOInputs:[]*tx.UTXOTxInput{},
		BalanceInputs:[]*tx.BalanceTxInput{},
		Outputs:[]*tx.TxOutput{},
		Programs:[]*program.Program{},
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
		go ds.InitializeConsensus(viewNumber)
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
		ledger.DefaultLedger.Blockchain.BCEvents.UnSubscribe(events.EventBlockPersistCompleted, ds.blockPersistCompletedSubscriber)
		ds.localNet.GetEvent("consensus").UnSubscribe(events.EventNewInventory, ds.newInventorySubscriber)
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

	if ds.context.MinerIndex < 0 {
		log.Error("Miner Index incorrect ", ds.context.MinerIndex)
		return NewDetailErr(errors.New("Miner Index incorrect"), ErrNoCode, "")
	}
	log.Debug("ds.context.MinerIndex ", ds.context.MinerIndex)
	log.Debug("ds.context.PrimaryIndex ", ds.context.PrimaryIndex)
	if ds.context.MinerIndex == int(ds.context.PrimaryIndex) {
		Trace()
		ds.context.State |= Primary
		ds.timerHeight = ds.context.Height
		ds.timeView = viewNum
		span := time.Now().Sub(ds.blockReceivedTime)
		if span > TimePerBlock {
			Trace()
			ds.timer.Stop()
			Trace()
			ds.timer.Reset(0)
			//go ds.Timeout()
		} else {
			Trace()
			ds.timer.Stop()
			log.Debug("The reset value is ", TimePerBlock - span)
			ds.timer.Reset(TimePerBlock - span)
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
			payload, ret := inventory.(*msg.ConsensusPayload)
			if (ret == true) {
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

	if int(payload.MinerIndex) == ds.context.MinerIndex {
		return
	}
	if payload.Version != ContextVersion || payload.PrevHash != ds.context.PrevHash || payload.Height != ds.context.Height {
		return
	}

	if int(payload.MinerIndex) >= len(ds.context.Miners) {
		return
	}

	message,err := DeserializeMessage(payload.Data)
	if err != nil {
		log.Error(fmt.Sprintf("DeserializeMessage failed: %s\n", err))
		return
	}

	if message.ViewNumber() != ds.context.ViewNumber && message.Type() != ChangeViewMsg {
		fmt.Printf("message.ViewNumber()=%d\n",message.ViewNumber())
		fmt.Printf("ds.context.ViewNumber=%d\n",ds.context.ViewNumber)
		fmt.Printf("message.Type()=%d\n",message.Type())
		fmt.Printf("ChangeViewMsg=%d\n",ChangeViewMsg)
		return
	}

	switch message.Type() {
	case ChangeViewMsg:
		if cv, ok := message.(*ChangeView); ok {
			ds.ChangeViewReceived(payload, cv)
		}
		break
	case PrepareRequestMsg:
		if pr, ok := message.(*PrepareRequest); ok {
			ds.PrepareRequestReceived(payload, pr)
		}
		break
	case PrepareResponseMsg:
		if pres, ok := message.(*PrepareResponse); ok {
			ds.PrepareResponseReceived(payload, pres)
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

func (ds *DbftService) PrepareRequestReceived(payload *msg.ConsensusPayload, message *PrepareRequest) {
	Trace()
	log.Info(fmt.Sprintf("Prepare Request Received: height=%d View=%d index=%d tx=%d",payload.Height,message.ViewNumber(),payload.MinerIndex,len(message.TransactionHashes)))

	if !ds.context.State.HasFlag(Backup) || ds.context.State.HasFlag(RequestReceived) {
		fmt.Println("PrepareRequestReceived ds.context.State.HasFlag(Backup)=",ds.context.State.HasFlag(Backup))
		fmt.Println("PrepareRequestReceived ds.context.State.HasFlag(RequestReceived)=",ds.context.State.HasFlag(RequestReceived))
		return
	}
	Trace()
	if uint32(payload.MinerIndex) != ds.context.PrimaryIndex {
		fmt.Println("PrepareRequestReceived uint32(payload.MinerIndex)=",uint32(payload.MinerIndex))
		fmt.Println("PrepareRequestReceived ds.context.PrimaryIndex=",ds.context.PrimaryIndex)
		return
	}
	header,err := ledger.DefaultLedger.Blockchain.GetHeader(ds.context.PrevHash)
	if err != nil {
		fmt.Println("PrepareRequestReceived GetHeader failed with ds.context.PrevHash",ds.context.PrevHash)
	}

	Trace()
	//TODO Add Error Catch
	prevBlockTimestamp := header.Blockdata.Timestamp
	if payload.Timestamp <= prevBlockTimestamp || payload.Timestamp > uint32(time.Now().Add(time.Minute*10).Unix()){
		con.Log(fmt.Sprintf("PrepareRequestReceived Timestamp incorrect: %d",payload.Timestamp))
		fmt.Println("PrepareRequestReceived payload.Timestamp=",payload.Timestamp,)
		fmt.Println("PrepareRequestReceived prevBlockTimestamp=",prevBlockTimestamp)
		fmt.Println("PrepareRequestReceived uint32(time.Now().Add(time.Minute*10).Unix()=",uint32(time.Now().Add(time.Minute*10).Unix()))
		return
	}

	ds.context.State |= RequestReceived
	ds.context.Timestamp = payload.Timestamp
	ds.context.Nonce = message.Nonce
	ds.context.NextMiner = message.NextMiner
	ds.context.TransactionHashes = message.TransactionHashes
	ds.context.Transactions = make(map[Uint256]*tx.Transaction)

	Trace()
	_, err = va.VerifySignature(ds.context.MakeHeader(), ds.context.Miners[payload.MinerIndex], message.Signature)
	if err != nil {
		log.Warn("PrepareRequestReceived VerifySignature failed.", err)
		return
	}

	minerLen := len(ds.context.Miners)
	ds.context.Signatures = make([][]byte, minerLen)
	ds.context.Signatures[payload.MinerIndex] = message.Signature
	Trace()
	if err := ds.AddTransaction(message.BookkeepingTransaction); err != nil {
		log.Warn("PrepareRequestReceived AddTransaction failed", err)
		return
	}
	Trace()
	mempool :=  ds.localNet.GetMemoryPool()
	for _, hash := range ds.context.TransactionHashes[1:] {
		if transaction,ok := mempool[hash]; ok{
			if err := ds.AddTransaction(transaction); err != nil {
				fmt.Println("PrepareRequestReceived AddTransaction failed.")
				return
			}
		}
	}

	//TODO: LocalNode allow hashes (add Except method)
	//AllowHashes(ds.context.TransactionHashes)
	log.Info("Prepare Requst finished")
	if len(ds.context.Transactions) < len(ds.context.TransactionHashes) {
		ds.localNet.SynchronizeMemoryPool()
	}
}

func (ds *DbftService) PrepareResponseReceived(payload *msg.ConsensusPayload,message *PrepareResponse){
	Trace()

	log.Info(fmt.Sprintf("Prepare Response Received: height=%d View=%d index=%d",payload.Height,message.ViewNumber(),payload.MinerIndex))

	if ds.context.State.HasFlag(BlockSent)  {return}
	if ds.context.Signatures[payload.MinerIndex] != nil {return }

	header := ds.context.MakeHeader()
	if  header == nil {return }
	if _,err := va.VerifySignature(header,ds.context.Miners[payload.MinerIndex],message.Signature); err != nil {
		return
	}

	ds.context.Signatures[payload.MinerIndex] = message.Signature
	ds.CheckSignatures()
	log.Info("Prepare Response finished")
}

func  (ds *DbftService) RefreshPolicy(){
	Trace()
	con.DefaultPolicy.Refresh()
}

func  (ds *DbftService) RequestChangeView() {
	Trace()
	// FIXME if there is no save block notifcation, when the timeout call this function it will crash
	ds.context.ExpectedView[ds.context.MinerIndex]++
	log.Info(fmt.Sprintf("Request change view: height=%d View=%d nv=%d state=%d",ds.context.Height,ds.context.ViewNumber,ds.context.MinerIndex,ds.context.State))

	ds.timer.Stop()
	ds.timer.Reset(SecondsPerBlock << (ds.context.ExpectedView[ds.context.MinerIndex]+1))

	ds.SignAndRelay(ds.context.MakeChangeView())
	ds.CheckExpectedView(ds.context.ExpectedView[ds.context.MinerIndex])
}

func (ds *DbftService) SignAndRelay(payload *msg.ConsensusPayload){
	Trace()
	ctCxt := ct.NewContractContext(payload)

	ret := ds.Client.Sign(ctCxt)
	if (ret == false) {
		log.Warn("Sign contract failure")
	}
	prog := ctCxt.GetPrograms()
	if (prog == nil) {
		log.Warn("Get programe failure")
	}
	payload.SetPrograms(prog)
	ds.localNet.Xmit(payload)
}

func (ds *DbftService) Start() error  {
	Trace()
	ds.started = true

	ds.newInventorySubscriber = ledger.DefaultLedger.Blockchain.BCEvents.Subscribe(events.EventBlockPersistCompleted,ds.BlockPersistCompleted)
	ds.blockPersistCompletedSubscriber = ds.localNet.GetEvent("consensus").Subscribe(events.EventNewInventory,ds.LocalNodeNewInventory)

	go ds.InitializeConsensus(0)
	return nil
}

func (ds *DbftService) Timeout() {
	Trace()
	ds.mu.Lock()
	defer ds.mu.Unlock()
	if ds.timerHeight != ds.context.Height || ds.timeView != ds.context.ViewNumber {
		return
	}
	log.Info("Timeout: height: ", ds.timerHeight, " View: ", ds.timeView, " State: ", ds.context.State)
	fmt.Printf(" ds.context.State %x\n", ds.context.State)
	fmt.Println("ds.context.State.HasFlag(Primary) ",ds.context.State.HasFlag(Primary))
	fmt.Println("ds.context.State.HasFlag(RequestSent) ",ds.context.State.HasFlag(RequestSent))
	fmt.Println("ds.context.State.HasFlag(Backup) ",ds.context.State.HasFlag(Backup))

	if ds.context.State.HasFlag(Primary) && !ds.context.State.HasFlag(RequestSent) {
		log.Info("Send prepare request: height: ", ds.timerHeight, " View: ", ds.timeView, " State: ", ds.context.State)
		ds.context.State |= RequestSent
		if !ds.context.State.HasFlag(SignatureSent) {

			//set context Timestamp
			now := uint32(time.Now().Unix())
			fmt.Println("ds.context.PrevHash",ds.context.PrevHash)
			header,_:= ledger.DefaultLedger.Blockchain.GetHeader(ds.context.PrevHash)
			fmt.Println(" ledger.DefaultLedger.Blockchain.GetHeader(ds.context.PrevHash)",header)

			// TODO Error Catch
			blockTime := header.Blockdata.Timestamp

			if blockTime > now {
				ds.context.Timestamp = blockTime
			} else {
				ds.context.Timestamp = now
			}

			ds.context.Nonce = GetNonce()
			transactions := ds.localNet.GetMemoryPool() //TODO: add policy

			txBookkeeping := ds.CreateBookkeepingTransaction(ds.context.Nonce)
			transactions[txBookkeeping.Hash()] = txBookkeeping

			//if ds.context.TransactionHashes == nil {
			//	ds.context.TransactionHashes = []Uint256{}
			//}
			trxhashes :=  []Uint256{}
			trxhashes = append(trxhashes,txBookkeeping.Hash())
			for _, v := range ds.context.TransactionHashes {
				trxhashes = append(trxhashes,v)
			}
			ds.context.TransactionHashes= trxhashes
			ds.context.Transactions = transactions

			txlist := ds.context.GetTransactionList()
			ds.context.NextMiner,_= ledger.GetMinerAddress(ledger.DefaultLedger.Blockchain.GetMinersByTXs(txlist))
			//TODO: add error catch
			block := ds.context.MakeHeader()
			account,_:= ds.Client.GetAccount(ds.context.Miners[ds.context.MinerIndex])
			//TODO: add error catch
			ds.context.Signatures[ds.context.MinerIndex], _ = sig.SignBySigner(block, account)
			//TODO: add error catch
		}
		payload := ds.context.MakePrepareRequest()
		ds.SignAndRelay(payload)
		ds.timer.Stop()
		ds.timer.Reset(SecondsPerBlock << (ds.timeView + 1))
	} else if ds.context.State.HasFlag(Primary) && ds.context.State.HasFlag(RequestSent) || ds.context.State.HasFlag(Backup){
		ds.RequestChangeView()
	}
}

func (ds *DbftService) timerRoutine () {
	Trace()
	for {
		select {
		case <-ds.timer.C:
			log.Debug("******Get a timeout notice")
			go ds.Timeout()
		}
	}
}
