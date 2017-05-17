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

const (
	INVDELAYTIME = 20 * time.Millisecond
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
		log.Info(fmt.Sprintf("persist block: %d", block.Hash()))
		err := ds.localNet.CleanSubmittedTransactions(block)
		if err != nil {
			log.Warn(err)
		}
		//log.Debug(fmt.Sprintf("persist block: %d with %d transactions\n", block.Hash(),len(trxHashToBeDelete)))
	}

	ds.blockReceivedTime = time.Now()

	go ds.InitializeConsensus(0)
}

func (ds *DbftService) CheckExpectedView(viewNumber byte) {
	log.Debug()
	if ds.context.State.HasFlag(BlockSent) {
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
		//fill transactions
		block.Transactions = ds.context.Transactions
		//set signed program to the block
		cxt.Data.SetPrograms(cxt.GetPrograms())

		hash := block.Hash()
		if !ledger.DefaultLedger.BlockInLedger(hash) {
			// save block
			if err := ledger.DefaultLedger.Blockchain.AddBlock(block); err != nil {
				log.Warn("Block saving error: ", hash)
				return err
			}

			// wait peers for saving block
			t := time.NewTimer(INVDELAYTIME)
			select {
			case <-t.C:
				// broadcast block hash
				if err := ds.localNet.Xmit(hash); err != nil {
					log.Warn("Block hash transmitting error: ", hash)
					return err
				}
			}
			ds.context.State |= BlockSent
		}
	}
	return nil
}

func (ds *DbftService) CreateBookkeepingTransaction(nonce uint64) *tx.Transaction {
	log.Debug()

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
	log.Debug()
	log.Info(fmt.Sprintf("Change View Received: height=%d View=%d index=%d nv=%d", payload.Height, message.ViewNumber(), payload.MinerIndex, message.NewViewNumber))

	if message.NewViewNumber <= ds.context.ExpectedView[payload.MinerIndex] {
		return
	}

	ds.context.ExpectedView[payload.MinerIndex] = message.NewViewNumber

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
		if ds.context.State.HasFlag(BlockSent) {
			return nil
		}
		ds.context.ChangeView(viewNum)
	}

	if ds.context.MinerIndex < 0 {
		log.Error("Miner Index incorrect ", ds.context.MinerIndex)
		return NewDetailErr(errors.New("Miner Index incorrect"), ErrNoCode, "")
	}

	if ds.context.MinerIndex == int(ds.context.PrimaryIndex) {

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

func (ds *DbftService) GetUnverifiedTxs(txs []*tx.Transaction) []*tx.Transaction {
	if len(ds.context.Transactions) == 0 {
		return nil
	}
	txpool := ds.localNet.GetTxnPool(false)
	ret := []*tx.Transaction{}
	for _, t := range txs {
		if _, ok := txpool[t.Hash()]; !ok {
			ret = append(ret, t)
		}
	}
	return ret
}

func VerifyTxs(txs []*tx.Transaction) error {
	for _, t := range txs {
		//TODO verify tx with transaction pool
		if err := va.VerifyTransaction(t); err != nil {
			return errors.New("Transaction verification failed")
		}
		if err := va.VerifyTransactionWithLedger(t, ledger.DefaultLedger); err != nil {
			return errors.New("Transaction verification with ledger failed")
		}
	}
	return nil
}

func (ds *DbftService) PrepareRequestReceived(payload *msg.ConsensusPayload, message *PrepareRequest) {
	log.Debug()
	log.Info(fmt.Sprintf("Prepare Request Received: height=%d View=%d index=%d tx=%d", payload.Height, message.ViewNumber(), payload.MinerIndex, len(message.Transactions)))

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
	ds.context.Transactions = message.Transactions

	//block header verification
	_, err = va.VerifySignature(ds.context.MakeHeader(), ds.context.Miners[payload.MinerIndex], message.Signature)
	if err != nil {
		log.Warn("PrepareRequestReceived VerifySignature failed.", err)
		return
	}

	ds.context.Signatures = make([][]byte, len(ds.context.Miners))
	ds.context.Signatures[payload.MinerIndex] = message.Signature

	//check if the transactions received are verified. If it already exists in transaction pool
	//then no need to verify it again. Otherwise, verify it.
	unverifyed := ds.GetUnverifiedTxs(ds.context.Transactions)
	if err := VerifyTxs(unverifyed); err != nil {
		log.Error("PrepareRequestReceived new transaction verification failed, will not sent Prepare Response", err)
		return
	}

	minerAddress, err := ledger.GetMinerAddress(ds.context.Miners)
	if err != nil {
		log.Error("[DbftService] GetMinerAddres failed")
		return
	}
	if minerAddress == ds.context.NextMiner {
		log.Info("send prepare response")
		ds.context.State |= SignatureSent
		miner, err := ds.Client.GetAccount(ds.context.Miners[ds.context.MinerIndex])
		if err != nil {
			log.Error("[DbftService] GetAccount failed")
			return

		}
		ds.context.Signatures[ds.context.MinerIndex], err = sig.SignBySigner(ds.context.MakeHeader(), miner)
		if err != nil {
			log.Error("[DbftService] SignBySigner failed")
			return
		}
		payload := ds.context.MakePrepareResponse(ds.context.Signatures[ds.context.MinerIndex])
		ds.SignAndRelay(payload)
	} else {
		ds.RequestChangeView()
		return
	}
	log.Info("Prepare Request finished")
}

func (ds *DbftService) PrepareResponseReceived(payload *msg.ConsensusPayload, message *PrepareResponse) {
	log.Debug()

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
	log.Debug()
	con.DefaultPolicy.Refresh()
}

func (ds *DbftService) RequestChangeView() {
	log.Debug()
	// FIXME if there is no save block notifcation, when the timeout call this function it will crash
	ds.context.ExpectedView[ds.context.MinerIndex] = ds.context.ExpectedView[ds.context.MinerIndex] + 1
	log.Info(fmt.Sprintf("Request change view: height=%d View=%d nv=%d state=%s", ds.context.Height,
		ds.context.ViewNumber, ds.context.ExpectedView[ds.context.MinerIndex], ds.context.GetStateDetail()))

	ds.timer.Stop()
	ds.timer.Reset(GenBlockTime << (ds.context.ExpectedView[ds.context.MinerIndex] + 1))

	ds.SignAndRelay(ds.context.MakeChangeView())
	ds.CheckExpectedView(ds.context.ExpectedView[ds.context.MinerIndex])
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

	if config.Parameters.GenBlockTime > 0 {
		GenBlockTime = time.Duration(config.Parameters.GenBlockTime) * time.Second
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
			transactionsPool := ds.localNet.GetTxnPool(false)
			//TODO: add policy
			//TODO: add max TX limitation

			txBookkeeping := ds.CreateBookkeepingTransaction(ds.context.Nonce)
			//add book keeping transaction first
			ds.context.Transactions = append(ds.context.Transactions, txBookkeeping)
			//add transactions from transaction pool
			for _, tx := range transactionsPool {
				ds.context.Transactions = append(ds.context.Transactions, tx)
			}
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
	log.Debug()
	for {
		select {
		case <-ds.timer.C:
			log.Debug("******Get a timeout notice")
			go ds.Timeout()
		}
	}
}
