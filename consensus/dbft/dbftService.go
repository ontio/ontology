package dbft

import (
	cl "DNA/client"
	. "DNA/common"
	"DNA/common/log"
	"DNA/config"
	con "DNA/consensus"
	ct "DNA/core/contract"
	"DNA/core/contract/program"
	"DNA/core/ledger"
	_ "DNA/core/signature"
	sig "DNA/core/signature"
	tx "DNA/core/transaction"
	"DNA/core/transaction/payload"
	va "DNA/core/validation"
	. "DNA/errors"
	"DNA/events"
	"DNA/net"
	msg "DNA/net/message"
	"errors"
	"fmt"
	"time"
)

var GenBlockTime = (2 * time.Second)

type DbftService struct {
	context           ConsensusContext
	Client            cl.Client
	timer             *time.Timer
	timerHeight       uint32
	timeView          byte
	blockReceivedTime time.Time
	logDictionary     string
	started           bool
	localNet          net.Neter

	newInventorySubscriber          events.Subscriber
	blockPersistCompletedSubscriber events.Subscriber
}

func NewDbftService(client cl.Client, logDictionary string, localNet net.Neter) *DbftService {
	Trace()

	ds := &DbftService{
		Client:        client,
		timer:         time.NewTimer(time.Second * 15),
		started:       false,
		localNet:      localNet,
		logDictionary: logDictionary,
	}

	if !ds.timer.Stop() {
		<-ds.timer.C
	}
	Trace()
	go ds.timerRoutine()
	return ds
}

func (ds *DbftService) AddTransaction(TX *tx.Transaction, needVerify bool) error {
	Trace()

	//check whether the new TX already exist in ledger
	if ledger.DefaultLedger.Blockchain.ContainsTransaction(TX.Hash()) {
		log.Warn(fmt.Sprintf("[AddTransaction] TX already Exist: %v", TX.Hash()))
		ds.RequestChangeView()
		return errors.New("TX already Exist.")
	}

	//verify the TX
	if needVerify {
		err := va.VerifyTransaction(TX, ledger.DefaultLedger, ds.context.GetTransactionList())
		if err != nil {
			log.Warn(fmt.Sprintf("[AddTransaction] TX Verfiy failed: %v", TX.Hash()))
			ds.RequestChangeView()
			return errors.New("TX Verfiy failed.")
		}
	}

	//check the TX policy
	//checkPolicy :=  ds.CheckPolicy(TX)

	//set TX to current context
	ds.context.Transactions[TX.Hash()] = TX

	//if enough TXs already added to context, build block and sign/relay
	if len(ds.context.TransactionHashes) == len(ds.context.Transactions) {

		minerAddress, err := ledger.GetMinerAddress(ds.context.Miners)
		if err != nil {
			return NewDetailErr(err, ErrNoCode, "[DbftService] ,GetMinerAddress failed")
		}

		if minerAddress == ds.context.NextMiner {
			log.Info("send prepare response")
			ds.context.State |= SignatureSent
			miner, err := ds.Client.GetAccount(ds.context.Miners[ds.context.MinerIndex])
			if err != nil {
				return NewDetailErr(err, ErrNoCode, "[DbftService] ,GetAccount failed.")
			}
			//sig.SignBySigner(ds.context.MakeHeader(), miner)
			ds.context.Signatures[ds.context.MinerIndex], err = sig.SignBySigner(ds.context.MakeHeader(), miner)
			if err != nil {
				log.Error("[DbftService], SignBySigner failed.")
				return NewDetailErr(err, ErrNoCode, "[DbftService], SignBySigner failed.")
			}
			payload := ds.context.MakePrepareResponse(ds.context.Signatures[ds.context.MinerIndex])
			ds.SignAndRelay(payload)
			err = ds.CheckSignatures()
			if err != nil {
				return NewDetailErr(err, ErrNoCode, "[DbftService] ,CheckSignatures failed.")
			}
		} else {
			ds.RequestChangeView()
			return errors.New("No valid Next Miner.")

		}
	}
	return nil
}

func (ds *DbftService) BlockPersistCompleted(v interface{}) {
	Trace()
	if block, ok := v.(*ledger.Block); ok {
		log.Info(fmt.Sprintf("persist block: %d", block.Hash()))
	}

	ds.blockReceivedTime = time.Now()

	go ds.InitializeConsensus(0)
}

func (ds *DbftService) CheckExpectedView(viewNumber byte) {
	Trace()
	if ds.context.ViewNumber == viewNumber {
		return
	}

	//check the count for same view number
	count := 0
	for _, expectedViewNumber := range ds.context.ExpectedView {
		if expectedViewNumber == viewNumber {
			count++
		}
	}

	M := ds.context.M()
	if count >= M {
		log.Debug("[CheckExpectedView] Begin InitializeConsensus.")
		go ds.InitializeConsensus(viewNumber)
		//ds.InitializeConsensus(viewNumber)
	}
}

func (ds *DbftService) CheckPolicy(transaction *tx.Transaction) error {
	//TODO: CheckPolicy

	return nil
}

func (ds *DbftService) CheckSignatures() error {
	Trace()

	//check have enought signatures and all required TXs already in context
	if ds.context.GetSignaturesCount() >= ds.context.M() && ds.context.CheckTxHashesExist() {

		//get current index's hash
		ep, err := ds.context.Miners[ds.context.MinerIndex].EncodePoint(true)
		if err != nil {
			return NewDetailErr(err, ErrNoCode, "[DbftService] ,EncodePoint failed")
		}
		codehash, err := ToCodeHash(ep)
		if err != nil {
			return NewDetailErr(err, ErrNoCode, "[DbftService] ,ToCodeHash failed")
		}

		//create multi-sig contract with all miners
		contract, err := ct.CreateMultiSigContract(codehash, ds.context.M(), ds.context.Miners)
		if err != nil {
			return err
		}

		//build block
		block := ds.context.MakeHeader()

		//sign the block with all miners and add signed contract to context
		cxt := ct.NewContractContext(block)
		for i, j := 0, 0; i < len(ds.context.Miners) && j < ds.context.M(); i++ {
			if ds.context.Signatures[i] != nil {
				err := cxt.AddContract(contract, ds.context.Miners[i], ds.context.Signatures[i])
				if err != nil {
					log.Error("[CheckSignatures] Multi-sign add contract error:", err.Error())
					return NewDetailErr(err, ErrNoCode, "[DbftService], CheckSignatures AddContract failed.")
				}
				j++
			}
		}
		//set signed program to the block
		cxt.Data.SetPrograms(cxt.GetPrograms())

		block.Transcations = ds.context.GetTXByHashes()

		if err := ds.localNet.Xmit(block); err != nil {
			log.Info(fmt.Sprintf("[CheckSignatures] Xmit block Error: %s, blockHash: %d", err.Error(), block.Hash()))
		}
		ds.context.State |= BlockSent
	}
	return nil
}

func (ds *DbftService) CreateBookkeepingTransaction(nonce uint64) *tx.Transaction {
	Trace()

	//TODO: sysfee

	return &tx.Transaction{
		TxType:         tx.BookKeeping,
		PayloadVersion: 0x2,
		Payload:        &payload.BookKeeping{},
		Nonce:          nonce, //TODO: update the nonce
		Attributes:     []*tx.TxAttribute{},
		UTXOInputs:     []*tx.UTXOTxInput{},
		BalanceInputs:  []*tx.BalanceTxInput{},
		Outputs:        []*tx.TxOutput{},
		Programs:       []*program.Program{},
	}
}

func (ds *DbftService) ChangeViewReceived(payload *msg.ConsensusPayload, message *ChangeView) {
	Trace()
	log.Info(fmt.Sprintf("Change View Received: height=%d View=%d index=%d nv=%d", payload.Height, message.ViewNumber(), payload.MinerIndex, message.NewViewNumber))

	if message.NewViewNumber <= ds.context.ExpectedView[payload.MinerIndex] {
		return
	}

	ds.context.ExpectedView[payload.MinerIndex] = message.NewViewNumber

	ds.CheckExpectedView(message.NewViewNumber)
}

func (ds *DbftService) Halt() error {
	Trace()
	log.Info("DBFT Stop")
	if ds.timer != nil {
		ds.timer.Stop()
	}

	if ds.started {
		ledger.DefaultLedger.Blockchain.BCEvents.UnSubscribe(events.EventBlockPersistCompleted, ds.blockPersistCompletedSubscriber)
		ds.localNet.GetEvent("consensus").UnSubscribe(events.EventNewInventory, ds.newInventorySubscriber)
	}
	return nil
}

func (ds *DbftService) InitializeConsensus(viewNum byte) error {
	log.Debug("[InitializeConsensus] Start InitializeConsensus.")
	Trace()
	ds.context.contextMu.Lock()
	defer ds.context.contextMu.Unlock()

	log.Debug("[InitializeConsensus] viewNum: ", viewNum)

	if viewNum == 0 {
		ds.context.Reset(ds.Client, ds.localNet)
	} else {
		ds.context.ChangeView(viewNum)
	}

	if ds.context.MinerIndex < 0 {
		log.Error("Miner Index incorrect ", ds.context.MinerIndex)
		return NewDetailErr(errors.New("Miner Index incorrect"), ErrNoCode, "")
	}

	if ds.context.MinerIndex == int(ds.context.PrimaryIndex) {

		//primary peer
		Trace()
		ds.context.State |= Primary
		ds.timerHeight = ds.context.Height
		ds.timeView = viewNum
		span := time.Now().Sub(ds.blockReceivedTime)
		if span > GenBlockTime {
			//TODO: double check the is the stop necessary
			ds.timer.Stop()
			ds.timer.Reset(0)
			//go ds.Timeout()
		} else {
			ds.timer.Stop()
			ds.timer.Reset(GenBlockTime - span)
		}
	} else {

		//backup peer
		ds.context.State = Backup
		ds.timerHeight = ds.context.Height
		ds.timeView = viewNum

		ds.timer.Stop()
		ds.timer.Reset(GenBlockTime << (viewNum + 1))
	}
	return nil
}

func (ds *DbftService) LocalNodeNewInventory(v interface{}) {
	Trace()
	if inventory, ok := v.(Inventory); ok {
		if inventory.Type() == CONSENSUS {
			payload, ret := inventory.(*msg.ConsensusPayload)
			if ret == true {
				ds.NewConsensusPayload(payload)
			}
		} else if inventory.Type() == TRANSACTION {
			transaction, isTransaction := inventory.(*tx.Transaction)
			if isTransaction {
				ds.NewTransactionPayload(transaction)
			}
		}
	}
}

//TODO: add invenory receiving

func (ds *DbftService) NewConsensusPayload(payload *msg.ConsensusPayload) {
	Trace()
	ds.context.contextMu.Lock()
	defer ds.context.contextMu.Unlock()

	//if payload from current peer, ignore it
	if int(payload.MinerIndex) == ds.context.MinerIndex {
		return
	}

	//if payload is not same height with current contex, ignore it
	if payload.Version != ContextVersion || payload.PrevHash != ds.context.PrevHash || payload.Height != ds.context.Height {
		return
	}

	if int(payload.MinerIndex) >= len(ds.context.Miners) {
		return
	}

	message, err := DeserializeMessage(payload.Data)
	if err != nil {
		log.Error(fmt.Sprintf("DeserializeMessage failed: %s\n", err))
		return
	}

	if message.ViewNumber() != ds.context.ViewNumber && message.Type() != ChangeViewMsg {
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

func (ds *DbftService) NewTransactionPayload(transaction *tx.Transaction) error {
	Trace()
	ds.context.contextMu.Lock()
	defer ds.context.contextMu.Unlock()

	if !ds.context.State.HasFlag(Backup) || !ds.context.State.HasFlag(RequestReceived) || ds.context.State.HasFlag(SignatureSent) {
		return NewDetailErr(errors.New("Consensus State is incorrect."), ErrNoCode, "")
	}

	if _, hasTx := ds.context.Transactions[transaction.Hash()]; hasTx {
		return NewDetailErr(errors.New("The transaction already exist."), ErrNoCode, "")
	}

	if !ds.context.HasTxHash(transaction.Hash()) {
		return NewDetailErr(errors.New("The transaction hash is not exist."), ErrNoCode, "")
	}
	return ds.AddTransaction(transaction, true)
}

func (ds *DbftService) PrepareRequestReceived(payload *msg.ConsensusPayload, message *PrepareRequest) {
	Trace()
	log.Info(fmt.Sprintf("Prepare Request Received: height=%d View=%d index=%d tx=%d", payload.Height, message.ViewNumber(), payload.MinerIndex, len(message.TransactionHashes)))

	if !ds.context.State.HasFlag(Backup) || ds.context.State.HasFlag(RequestReceived) {
		return
	}

	if uint32(payload.MinerIndex) != ds.context.PrimaryIndex {
		return
	}
	header, err := ledger.DefaultLedger.Blockchain.GetHeader(ds.context.PrevHash)
	if err != nil {
		log.Info("PrepareRequestReceived GetHeader failed with ds.context.PrevHash", ds.context.PrevHash)
	}

	Trace()
	//TODO Add Error Catch
	prevBlockTimestamp := header.Blockdata.Timestamp
	if payload.Timestamp <= prevBlockTimestamp || payload.Timestamp > uint32(time.Now().Add(time.Minute*10).Unix()) {
		log.Info(fmt.Sprintf("Prepare Reques tReceived: Timestamp incorrect: %d", payload.Timestamp))
		return
	}

	ds.context.State |= RequestReceived
	ds.context.Timestamp = payload.Timestamp
	ds.context.Nonce = message.Nonce
	ds.context.NextMiner = message.NextMiner
	ds.context.TransactionHashes = message.TransactionHashes
	ds.context.Transactions = make(map[Uint256]*tx.Transaction)

	_, err = va.VerifySignature(ds.context.MakeHeader(), ds.context.Miners[payload.MinerIndex], message.Signature)
	if err != nil {
		log.Warn("PrepareRequestReceived VerifySignature failed.", err)
		return
	}

	ds.context.Signatures = make([][]byte, len(ds.context.Miners))
	ds.context.Signatures[payload.MinerIndex] = message.Signature

	mempool := ds.localNet.GetTxnPool(true)
	for _, hash := range ds.context.TransactionHashes[1:] {
		if transaction, ok := mempool[hash]; ok {
			if err := ds.AddTransaction(transaction, false); err != nil {
				log.Info("PrepareRequestReceived AddTransaction failed.")
				return
			}
		}
	}

	if err := ds.AddTransaction(message.BookkeepingTransaction, true); err != nil {
		log.Warn("PrepareRequestReceived AddTransaction failed", err)
		return
	}

	//TODO: LocalNode allow hashes (add Except method)
	//AllowHashes(ds.context.TransactionHashes)
	log.Info("Prepare Requst finished")
	if len(ds.context.Transactions) < len(ds.context.TransactionHashes) {
		ds.localNet.SynchronizeTxnPool()
	}
}

func (ds *DbftService) PrepareResponseReceived(payload *msg.ConsensusPayload, message *PrepareResponse) {
	Trace()

	log.Info(fmt.Sprintf("Prepare Response Received: height=%d View=%d index=%d", payload.Height, message.ViewNumber(), payload.MinerIndex))

	if ds.context.State.HasFlag(BlockSent) {
		return
	}

	//if the signature already exist, needn't handle again
	if ds.context.Signatures[payload.MinerIndex] != nil {
		return
	}

	header := ds.context.MakeHeader()
	if header == nil {
		return
	}
	if _, err := va.VerifySignature(header, ds.context.Miners[payload.MinerIndex], message.Signature); err != nil {
		return
	}

	ds.context.Signatures[payload.MinerIndex] = message.Signature
	ds.CheckSignatures()
	log.Info("Prepare Response finished")
}

func (ds *DbftService) RefreshPolicy() {
	Trace()
	con.DefaultPolicy.Refresh()
}

func (ds *DbftService) RequestChangeView() {
	Trace()
	// FIXME if there is no save block notifcation, when the timeout call this function it will crash
	ds.context.ExpectedView[ds.context.MinerIndex] = ds.context.ExpectedView[ds.context.MinerIndex] + 1
	log.Info(fmt.Sprintf("Request change view: height=%d View=%d nv=%d state=%s", ds.context.Height, ds.context.ViewNumber, ds.context.ExpectedView[ds.context.MinerIndex], ds.context.GetStateDetail()))

	ds.timer.Stop()
	ds.timer.Reset(GenBlockTime << (ds.context.ExpectedView[ds.context.MinerIndex] + 1))

	ds.SignAndRelay(ds.context.MakeChangeView())
	ds.CheckExpectedView(ds.context.ExpectedView[ds.context.MinerIndex])
}

func (ds *DbftService) SignAndRelay(payload *msg.ConsensusPayload) {
	Trace()

	prohash, err := payload.GetProgramHashes()
	if err != nil {
		log.Debug("[SignAndRelay] payload.GetProgramHashes failed: ", err.Error())
		return
	}
	log.Debug("[SignAndRelay] ConsensusPayload Program Hashes: ", prohash)

	ctCxt := ct.NewContractContext(payload)

	ret := ds.Client.Sign(ctCxt)
	if ret == false {
		log.Warn("[SignAndRelay] Sign contract failure")
	}
	prog := ctCxt.GetPrograms()
	if prog == nil {
		log.Warn("[SignAndRelay] Get programe failure")
	}
	payload.SetPrograms(prog)
	ds.localNet.Xmit(payload)
}

func (ds *DbftService) Start() error {
	Trace()
	ds.started = true

	if config.Parameters.GenBlockTime > 0 {
		GenBlockTime = time.Duration(config.Parameters.GenBlockTime) * time.Second
	}

	ds.blockPersistCompletedSubscriber = ledger.DefaultLedger.Blockchain.BCEvents.Subscribe(events.EventBlockPersistCompleted, ds.BlockPersistCompleted)
	ds.newInventorySubscriber = ds.localNet.GetEvent("consensus").Subscribe(events.EventNewInventory, ds.LocalNodeNewInventory)

	go ds.InitializeConsensus(0)
	return nil
}

func (ds *DbftService) Timeout() {
	Trace()
	ds.context.contextMu.Lock()
	defer ds.context.contextMu.Unlock()
	if ds.timerHeight != ds.context.Height || ds.timeView != ds.context.ViewNumber {
		return
	}

	log.Info("Timeout: height: ", ds.timerHeight, " View: ", ds.timeView, " State: ", ds.context.GetStateDetail())

	////temp change view number test
	//if ledger.DefaultLedger.Blockchain.BlockHeight > 2 {
	//	ds.RequestChangeView()
	//	return
	//}

	if ds.context.State.HasFlag(Primary) && !ds.context.State.HasFlag(RequestSent) {

		//parimary peer send the prepare request
		log.Info("Send prepare request: height: ", ds.timerHeight, " View: ", ds.timeView, " State: ", ds.context.GetStateDetail())
		ds.context.State |= RequestSent
		if !ds.context.State.HasFlag(SignatureSent) {

			//do signature
			now := uint32(time.Now().Unix())
			header, _ := ledger.DefaultLedger.Blockchain.GetHeader(ds.context.PrevHash)

			//set context Timestamp
			blockTime := header.Blockdata.Timestamp + 1
			if blockTime > now {
				ds.context.Timestamp = blockTime
			} else {
				ds.context.Timestamp = now
			}

			ds.context.Nonce = GetNonce()
			transactionsPool := ds.localNet.GetTxnPool(true) //TODO: add policy

			//TODO: add max TX limitation

			//convert txPool to tx list
			transactions := []*tx.Transaction{}

			//add new book keeping TX first
			txBookkeeping := ds.CreateBookkeepingTransaction(ds.context.Nonce)
			transactions = append(transactions, txBookkeeping)

			//add TXs from mem pool
			for _, tx := range transactionsPool {
				transactions = append(transactions, tx)
			}

			//add Transaction hashes
			trxhashes := []Uint256{}
			txMap := make(map[Uint256]*tx.Transaction)
			for _, tx := range transactions {
				txHash := tx.Hash()
				trxhashes = append(trxhashes, txHash)
				txMap[txHash] = tx
			}

			ds.context.TransactionHashes = trxhashes
			ds.context.Transactions = txMap

			//build block and sign
			ds.context.NextMiner, _ = ledger.GetMinerAddress(ds.context.Miners)
			block := ds.context.MakeHeader()
			account, _ := ds.Client.GetAccount(ds.context.Miners[ds.context.MinerIndex]) //TODO: handle error
			ds.context.Signatures[ds.context.MinerIndex], _ = sig.SignBySigner(block, account)

		}
		payload := ds.context.MakePrepareRequest()
		ds.SignAndRelay(payload)
		ds.timer.Stop()
		ds.timer.Reset(GenBlockTime << (ds.timeView + 1))
	} else if (ds.context.State.HasFlag(Primary) && ds.context.State.HasFlag(RequestSent)) || ds.context.State.HasFlag(Backup) {
		ds.RequestChangeView()
	}

}

func (ds *DbftService) timerRoutine() {
	Trace()
	for {
		select {
		case <-ds.timer.C:
			log.Debug("******Get a timeout notice")
			go ds.Timeout()
		}
	}
}
