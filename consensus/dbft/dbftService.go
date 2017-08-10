package dbft

import (
	cl "DNA/account"
	. "DNA/common"
	"DNA/common/config"
	"DNA/common/log"
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

const (
	MINGENBLOCKTIME = 6
)

var GenBlockTime = (MINGENBLOCKTIME * time.Second)

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
	log.Debug()

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
	log.Debug()
	go ds.timerRoutine()
	return ds
}

func (ds *DbftService) BlockPersistCompleted(v interface{}) {
	log.Debug()
	if block, ok := v.(*ledger.Block); ok {
		log.Infof("persist block: %x", block.Hash())
		err := ds.localNet.CleanSubmittedTransactions(block)
		if err != nil {
			log.Warn(err)
		}

		ds.localNet.Xmit(block.Hash())
		//log.Debug(fmt.Sprintf("persist block: %x with %d transactions\n", block.Hash(),len(trxHashToBeDelete)))
	}

	ds.blockReceivedTime = time.Now()

	go ds.InitializeConsensus(0)
}

func (ds *DbftService) CheckExpectedView(viewNumber byte) {
	log.Debug()
	if ds.context.State.HasFlag(BlockGenerated) {
		return
	}
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
	log.Debug()

	//check if get enough signatures
	if ds.context.GetSignaturesCount() >= ds.context.M() {

		//get current index's hash
		ep, err := ds.context.BookKeepers[ds.context.BookKeeperIndex].EncodePoint(true)
		if err != nil {
			return NewDetailErr(err, ErrNoCode, "[DbftService] ,EncodePoint failed")
		}
		codehash, err := ToCodeHash(ep)
		if err != nil {
			return NewDetailErr(err, ErrNoCode, "[DbftService] ,ToCodeHash failed")
		}

		//create multi-sig contract with all bookKeepers
		contract, err := ct.CreateMultiSigContract(codehash, ds.context.M(), ds.context.BookKeepers)
		if err != nil {
			log.Error("CheckSignatures CreateMultiSigContract error: ", err)
			return err
		}

		//build block
		block := ds.context.MakeHeader()
		//sign the block with all bookKeepers and add signed contract to context
		cxt := ct.NewContractContext(block)
		for i, j := 0, 0; i < len(ds.context.BookKeepers) && j < ds.context.M(); i++ {
			if ds.context.Signatures[i] != nil {
				err := cxt.AddContract(contract, ds.context.BookKeepers[i], ds.context.Signatures[i])
				if err != nil {
					log.Error("[CheckSignatures] Multi-sign add contract error:", err.Error())
					return NewDetailErr(err, ErrNoCode, "[DbftService], CheckSignatures AddContract failed.")
				}
				j++
			}
		}
		//fill transactions
		block.Transactions = ds.context.Transactions
		//set signed program to the block
		cxt.Data.SetPrograms(cxt.GetPrograms())

		hash := block.Hash()
		if !ledger.DefaultLedger.BlockInLedger(hash) {
			// save block
			if err := ledger.DefaultLedger.Blockchain.AddBlock(block); err != nil {
				log.Error(fmt.Sprintf("[CheckSignatures] Xmit block Error: %s, blockHash: %d", err.Error(), block.Hash()))
				return NewDetailErr(err, ErrNoCode, "[DbftService], CheckSignatures AddContract failed.")
			}

			ds.context.State |= BlockGenerated
		}
	}
	return nil
}

func (ds *DbftService) CreateBookkeepingTransaction(nonce uint64) *tx.Transaction {
	log.Debug()
	//TODO: sysfee
	bookKeepingPayload := &payload.BookKeeping{
		Nonce: uint64(time.Now().UnixNano()),
	}
	return &tx.Transaction{
		TxType:         tx.BookKeeping,
		PayloadVersion: payload.BookKeepingPayloadVersion,
		Payload:        bookKeepingPayload,
		Attributes:     []*tx.TxAttribute{},
		UTXOInputs:     []*tx.UTXOTxInput{},
		BalanceInputs:  []*tx.BalanceTxInput{},
		Outputs:        []*tx.TxOutput{},
		Programs:       []*program.Program{},
	}
}

func (ds *DbftService) ChangeViewReceived(payload *msg.ConsensusPayload, message *ChangeView) {
	log.Debug()
	log.Info(fmt.Sprintf("Change View Received: height=%d View=%d index=%d nv=%d", payload.Height, message.ViewNumber(), payload.BookKeeperIndex, message.NewViewNumber))

	if message.NewViewNumber <= ds.context.ExpectedView[payload.BookKeeperIndex] {
		return
	}

	ds.context.ExpectedView[payload.BookKeeperIndex] = message.NewViewNumber

	ds.CheckExpectedView(message.NewViewNumber)
}

func (ds *DbftService) Halt() error {
	log.Debug()
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
	log.Debug()
	ds.context.contextMu.Lock()
	defer ds.context.contextMu.Unlock()

	log.Debug("[InitializeConsensus] viewNum: ", viewNum)

	if viewNum == 0 {
		ds.context.Reset(ds.Client, ds.localNet)
	} else {
		if ds.context.State.HasFlag(BlockGenerated) {
			return nil
		}
		ds.context.ChangeView(viewNum)
	}

	if ds.context.BookKeeperIndex < 0 {
		log.Info("You aren't bookkeeper")
		return nil
	}

	if ds.context.BookKeeperIndex == int(ds.context.PrimaryIndex) {

		//primary peer
		log.Debug()
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
	log.Debug()
	if inventory, ok := v.(Inventory); ok {
		if inventory.Type() == CONSENSUS {
			payload, ret := inventory.(*msg.ConsensusPayload)
			if ret == true {
				ds.NewConsensusPayload(payload)
			}
		}
	}
}

//TODO: add invenory receiving

func (ds *DbftService) NewConsensusPayload(payload *msg.ConsensusPayload) {
	log.Debug()
	ds.context.contextMu.Lock()
	defer ds.context.contextMu.Unlock()

	//if payload from current peer, ignore it
	if int(payload.BookKeeperIndex) == ds.context.BookKeeperIndex {
		return
	}

	//if payload is not same height with current contex, ignore it
	if payload.Version != ContextVersion || payload.PrevHash != ds.context.PrevHash || payload.Height != ds.context.Height {
		return
	}

	if int(payload.BookKeeperIndex) >= len(ds.context.BookKeepers) {
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

func (ds *DbftService) GetUnverifiedTxs(txs []*tx.Transaction) []*tx.Transaction {
	if len(ds.context.Transactions) == 0 {
		return nil
	}
	txpool := ds.localNet.GetTxnPool(false)
	ret := []*tx.Transaction{}
	for _, t := range txs {
		if _, ok := txpool[t.Hash()]; !ok {
			if t.TxType != tx.BookKeeping {
				ret = append(ret, t)
			}
		}
	}
	return ret
}

func (ds *DbftService) VerifyTxs(txs []*tx.Transaction) error {
	for _, t := range txs {
		if ok := ds.localNet.AppendTxnPool(t); !ok {
			return errors.New("[dbftService] VerifyTxs failed when AppendTxnPool.")
		}
	}
	return nil
}

func (ds *DbftService) PrepareRequestReceived(payload *msg.ConsensusPayload, message *PrepareRequest) {
	log.Debug()
	log.Info(fmt.Sprintf("Prepare Request Received: height=%d View=%d index=%d tx=%d", payload.Height, message.ViewNumber(), payload.BookKeeperIndex, len(message.Transactions)))

	if !ds.context.State.HasFlag(Backup) || ds.context.State.HasFlag(RequestReceived) {
		return
	}

	if uint32(payload.BookKeeperIndex) != ds.context.PrimaryIndex {
		return
	}

	header, err := ledger.DefaultLedger.Blockchain.GetHeader(ds.context.PrevHash)
	if err != nil {
		log.Info("PrepareRequestReceived GetHeader failed with ds.context.PrevHash", ds.context.PrevHash)
	}

	//TODO Add Error Catch
	prevBlockTimestamp := header.Blockdata.Timestamp
	if payload.Timestamp <= prevBlockTimestamp || payload.Timestamp > uint32(time.Now().Add(time.Minute*10).Unix()) {
		log.Info(fmt.Sprintf("Prepare Reques tReceived: Timestamp incorrect: %d", payload.Timestamp))
		return
	}

	if ds.context.NextBookKeeper != message.NextBookKeeper {
		log.Info("PrepareRequestReceived: Get mismatched NextBookKeeper, RequestChangeView")
		ds.RequestChangeView()
		return
	}

	ds.context.State |= RequestReceived
	ds.context.Timestamp = payload.Timestamp
	ds.context.Nonce = message.Nonce
	ds.context.Transactions = message.Transactions
	ds.context.header = nil

	//block header verification
	_, err = va.VerifySignature(ds.context.MakeHeader(), ds.context.BookKeepers[payload.BookKeeperIndex], message.Signature)
	if err != nil {
		log.Warn("PrepareRequestReceived VerifySignature failed.", err)
		return
	}

	ds.context.Signatures = make([][]byte, len(ds.context.BookKeepers))
	ds.context.Signatures[payload.BookKeeperIndex] = message.Signature

	//check if the transactions received are verified. If it already exists in transaction pool
	//then no need to verify it again. Otherwise, verify it.
	unverifyed := ds.GetUnverifiedTxs(ds.context.Transactions)
	if err := ds.VerifyTxs(unverifyed); err != nil {
		log.Error("PrepareRequestReceived new transaction verification failed, will not sent Prepare Response", err)
		return
	}

	log.Info("send prepare response")
	ds.context.State |= SignatureSent
	bookKeeper, err := ds.Client.GetAccount(ds.context.BookKeepers[ds.context.BookKeeperIndex])
	if err != nil {
		log.Error("[DbftService] GetAccount failed")
		return
	}
	ds.context.Signatures[ds.context.BookKeeperIndex], err = sig.SignBySigner(ds.context.MakeHeader(), bookKeeper)
	if err != nil {
		log.Error("[DbftService] SignBySigner failed")
		return
	}
	payload = ds.context.MakePrepareResponse(ds.context.Signatures[ds.context.BookKeeperIndex])
	ds.SignAndRelay(payload)

	log.Info("Prepare Request finished")
}

func (ds *DbftService) PrepareResponseReceived(payload *msg.ConsensusPayload, message *PrepareResponse) {
	log.Debug()

	log.Info(fmt.Sprintf("Prepare Response Received: height=%d View=%d index=%d", payload.Height, message.ViewNumber(), payload.BookKeeperIndex))

	if ds.context.State.HasFlag(BlockGenerated) {
		return
	}

	//if the signature already exist, needn't handle again
	if ds.context.Signatures[payload.BookKeeperIndex] != nil {
		return
	}

	header := ds.context.MakeHeader()
	if header == nil {
		return
	}
	if _, err := va.VerifySignature(header, ds.context.BookKeepers[payload.BookKeeperIndex], message.Signature); err != nil {
		return
	}

	ds.context.Signatures[payload.BookKeeperIndex] = message.Signature
	err := ds.CheckSignatures()
	if err != nil {
		log.Error("CheckSignatures failed")
		return
	}
	log.Info("Prepare Response finished")
}

func (ds *DbftService) RefreshPolicy() {
	log.Debug()
	con.DefaultPolicy.Refresh()
}

func (ds *DbftService) RequestChangeView() {
	log.Debug()
	// FIXME if there is no save block notifcation, when the timeout call this function it will crash
	ds.context.ExpectedView[ds.context.BookKeeperIndex] = ds.context.ExpectedView[ds.context.BookKeeperIndex] + 1
	log.Info(fmt.Sprintf("Request change view: height=%d View=%d nv=%d state=%s", ds.context.Height,
		ds.context.ViewNumber, ds.context.ExpectedView[ds.context.BookKeeperIndex], ds.context.GetStateDetail()))

	ds.timer.Stop()
	ds.timer.Reset(GenBlockTime << (ds.context.ExpectedView[ds.context.BookKeeperIndex] + 1))

	ds.SignAndRelay(ds.context.MakeChangeView())
	ds.CheckExpectedView(ds.context.ExpectedView[ds.context.BookKeeperIndex])
}

func (ds *DbftService) SignAndRelay(payload *msg.ConsensusPayload) {
	log.Debug()

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
	log.Debug()
	ds.started = true

	if config.Parameters.GenBlockTime > MINGENBLOCKTIME {
		GenBlockTime = time.Duration(config.Parameters.GenBlockTime) * time.Second
	} else {
		log.Warn("The Generate block time should be longer than 6 seconds, so set it to be 6.")
	}

	ds.blockPersistCompletedSubscriber = ledger.DefaultLedger.Blockchain.BCEvents.Subscribe(events.EventBlockPersistCompleted, ds.BlockPersistCompleted)
	ds.newInventorySubscriber = ds.localNet.GetEvent("consensus").Subscribe(events.EventNewInventory, ds.LocalNodeNewInventory)

	go ds.InitializeConsensus(0)
	return nil
}

func (ds *DbftService) Timeout() {
	log.Debug()
	ds.context.contextMu.Lock()
	defer ds.context.contextMu.Unlock()
	if ds.timerHeight != ds.context.Height || ds.timeView != ds.context.ViewNumber {
		return
	}

	log.Info("Timeout: height: ", ds.timerHeight, " View: ", ds.timeView, " State: ", ds.context.GetStateDetail())

	if ds.context.State.HasFlag(Primary) && !ds.context.State.HasFlag(RequestSent) {
		//primary node send the prepare request
		log.Info("Send prepare request: height: ", ds.timerHeight, " View: ", ds.timeView, " State: ", ds.context.GetStateDetail())
		ds.context.State |= RequestSent
		if !ds.context.State.HasFlag(SignatureSent) {
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
			transactionsPool := ds.localNet.GetTxnPool(true)
			//TODO: add policy
			//TODO: add max TX limitation

			txBookkeeping := ds.CreateBookkeepingTransaction(ds.context.Nonce)
			//add book keeping transaction first
			ds.context.Transactions = append(ds.context.Transactions, txBookkeeping)
			//add transactions from transaction pool
			for _, tx := range transactionsPool {
				ds.context.Transactions = append(ds.context.Transactions, tx)
			}
			ds.context.header = nil
			//build block and sign
			block := ds.context.MakeHeader()
			account, _ := ds.Client.GetAccount(ds.context.BookKeepers[ds.context.BookKeeperIndex]) //TODO: handle error
			ds.context.Signatures[ds.context.BookKeeperIndex], _ = sig.SignBySigner(block, account)
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
	log.Debug()
	for {
		select {
		case <-ds.timer.C:
			log.Debug("******Get a timeout notice")
			go ds.Timeout()
		}
	}
}
