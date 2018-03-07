package dbft

import (
	"bytes"
	"fmt"
	"time"

	. "github.com/Ontology/common"
	"github.com/Ontology/common/config"
	"github.com/Ontology/common/log"
	"github.com/Ontology/core"
	"github.com/Ontology/core/contract"
	"github.com/Ontology/core/contract/program"
	"github.com/Ontology/core/genesis"
	"github.com/Ontology/core/ledger"
	"github.com/Ontology/core/payload"
	"github.com/Ontology/core/signature"
	"github.com/Ontology/core/transaction/utxo"
	"github.com/Ontology/core/types"
	"github.com/Ontology/core/vote"
	"github.com/Ontology/crypto"
	"github.com/Ontology/events"
	"github.com/Ontology/net"
	msg "github.com/Ontology/net/message"
	"github.com/Ontology/core/ledger/ledgerevent"
	"github.com/Ontology/account"
	clientActor "github.com/Ontology/consensus/actor"
	ontErrors "github.com/Ontology/errors"
	"github.com/Ontology/eventbus/actor"
)

type DbftService struct {
	context           ConsensusContext
	Account           *account.Account
	timer             *time.Timer
	timerHeight       uint32
	timeView          byte
	blockReceivedTime time.Time
	logDictionary     string
	started           bool
	localNet          net.Neter
	poolActor         *clientActor.TxPoolActor

	newInventorySubscriber          events.Subscriber
	blockPersistCompletedSubscriber events.Subscriber
}

func NewDbftService(bkAccount *account.Account, logDictionary string, txpool *actor.PID) *DbftService {

	ds := &DbftService{
		Account:       bkAccount,
		timer:         time.NewTimer(time.Second * 15),
		started:       false,
		poolActor:     &clientActor.TxPoolActor{Pool: txpool},
		logDictionary: logDictionary,
	}

	if !ds.timer.Stop() {
		<-ds.timer.C
	}
	go ds.timerRoutine()
	return ds
}

func (ds *DbftService) BlockPersistCompleted(v interface{}) {
	if block, ok := v.(*types.Block); ok {
		log.Infof("persist block: %x", block.Hash())

		ds.localNet.Xmit(block.Hash())
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

func (ds *DbftService) CheckPolicy(transaction *types.Transaction) error {
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
			return ontErrors.NewDetailErr(err, ontErrors.ErrNoCode, "[DbftService] ,EncodePoint failed")
		}
		codehash := ToCodeHash(ep)

		//create multi-sig contract with all bookKeepers
		ct, err := contract.CreateMultiSigContract(codehash, ds.context.M(), ds.context.BookKeepers)
		if err != nil {
			log.Error("CheckSignatures CreateMultiSigContract error: ", err)
			return err
		}

		//build block
		block := ds.context.MakeHeader()
		//sign the block with all bookKeepers and add signed contract to context
		sb := program.NewProgramBuilder()

		sigs := make([]SignaturesData, ds.context.M())
		for i, j := 0, 0; i < len(ds.context.BookKeepers) && j < ds.context.M(); i++ {
			if ds.context.Signatures[i] != nil {
				sigs[j].Index = uint16(i)
				sigs[j].Signature = ds.context.Signatures[i]

				sb.PushData(ds.context.Signatures[i])
				j++
			}
		}
		//set signed program to the block
		block.Header.Program = &program.Program{
			Code:      ct.Code,
			Parameter: sb.ToArray(),
		}
		//fill transactions
		block.Transactions = ds.context.Transactions

		hash := block.Hash()
		isExist, err := ledger.DefLedger.IsContainBlock(&hash)
		if err != nil {
			log.Errorf("DefLedger.IsContainBlock Hash:%x error:%s", hash, err)
			return err
		}
		if !isExist {
			// save block
			if err := ledger.DefLedger.AddBlock(block); err != nil {
				log.Error(fmt.Sprintf("[CheckSignatures] Xmit block Error: %s, blockHash: %d", err.Error(), block.Hash()))
				return ontErrors.NewDetailErr(err, ontErrors.ErrNoCode, "[DbftService], CheckSignatures AddContract failed.")
			}

			ds.context.State |= BlockGenerated
			payload := ds.context.MakeBlockSignatures(sigs)
			ds.SignAndRelay(payload)
		}
	}
	return nil
}

func (ds *DbftService) CreateBookkeepingTransaction(nonce uint64, fee Fixed64) *types.Transaction {
	log.Debug()
	//TODO: sysfee
	bookKeepingPayload := &payload.BookKeeping{
		Nonce: uint64(time.Now().UnixNano()),
	}
	signatureRedeemScript, err := contract.CreateSignatureRedeemScript(ds.context.Owner)
	if err != nil {
		return nil
	}
	signatureRedeemScriptHashToCodeHash := ToCodeHash(signatureRedeemScript)
	if err != nil {
		return nil
	}
	outputs := []*utxo.TxOutput{}
	if fee > 0 {
		feeOutput := &utxo.TxOutput{
			AssetID:     genesis.ONGTokenID,
			Value:       fee,
			ProgramHash: signatureRedeemScriptHashToCodeHash,
		}
		outputs = append(outputs, feeOutput)
	}
	return &types.Transaction{
		TxType: types.BookKeeping,
		//PayloadVersion: payload.BookKeepingPayloadVersion,
		Payload:    bookKeepingPayload,
		Attributes: []*types.TxAttribute{},
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
		ledgerevent.DefLedgerEvt.UnSubscribe(events.EventBlockPersistCompleted, ds.blockPersistCompletedSubscriber)
		ds.localNet.GetEvent("consensus").UnSubscribe(events.EventNewInventory, ds.newInventorySubscriber)
	}
	return nil
}

func (ds *DbftService) InitializeConsensus(viewNum byte) error {
	log.Debug("[InitializeConsensus] Start InitializeConsensus.")
	ds.context.contextMu.Lock()
	defer ds.context.contextMu.Unlock()

	log.Debug("[InitializeConsensus] viewNum: ", viewNum)

	if viewNum == 0 {
		ds.context.Reset(ds.Account, ds.localNet)
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
		ds.context.State |= Primary
		ds.timerHeight = ds.context.Height
		ds.timeView = viewNum
		span := time.Now().Sub(ds.blockReceivedTime)
		if span > genesis.GenBlockTime {
			//TODO: double check the is the stop necessary
			ds.timer.Stop()
			ds.timer.Reset(0)
			//go ds.Timeout()
		} else {
			ds.timer.Stop()
			ds.timer.Reset(genesis.GenBlockTime - span)
		}
	} else {

		//backup peer
		ds.context.State = Backup
		ds.timerHeight = ds.context.Height
		ds.timeView = viewNum

		ds.timer.Stop()
		ds.timer.Reset(genesis.GenBlockTime << (viewNum + 1))
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

	if ds.context.State.HasFlag(BlockGenerated) {
		return
	}

	if ds.context.State.HasFlag(BlockGenerated) {
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

	err = payload.Verify()
	if err != nil {
		log.Warn(err.Error())
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
	case BlockSignaturesMsg:
		if blockSigs, ok := message.(*BlockSignatures); ok {
			ds.BlockSignaturesReceived(payload, blockSigs)
		}
		break
	}
}

func (ds *DbftService) PrepareRequestReceived(payload *msg.ConsensusPayload, message *PrepareRequest) {
	log.Info(fmt.Sprintf("Prepare Request Received: height=%d View=%d index=%d tx=%d", payload.Height, message.ViewNumber(), payload.BookKeeperIndex, len(message.Transactions)))

	if !ds.context.State.HasFlag(Backup) || ds.context.State.HasFlag(RequestReceived) {
		return
	}

	if uint32(payload.BookKeeperIndex) != ds.context.PrimaryIndex {
		return
	}

	header, err := ledger.DefLedger.GetHeaderByHash(&ds.context.PrevHash)
	if err != nil {
		log.Info("PrepareRequestReceived GetHeader failed with ds.context.PrevHash", ds.context.PrevHash)
	}

	//TODO Add Error Catch
	prevBlockTimestamp := header.Timestamp
	if payload.Timestamp <= prevBlockTimestamp || payload.Timestamp > uint32(time.Now().Add(time.Minute*10).Unix()) {
		log.Info(fmt.Sprintf("Prepare Reques tReceived: Timestamp incorrect: %d", payload.Timestamp))
		return
	}

	backupContext := ds.context

	ds.context.State |= RequestReceived
	ds.context.Timestamp = payload.Timestamp
	ds.context.Nonce = message.Nonce
	ds.context.NextBookKeeper = message.NextBookKeeper
	ds.context.Transactions = message.Transactions
	ds.context.header = nil

	buf := new(bytes.Buffer)
	ds.context.MakeHeader().SerializeUnsigned(buf)
	err = crypto.Verify(*ds.context.BookKeepers[payload.BookKeeperIndex], buf.Bytes(), message.Signature)
	if err != nil {
		log.Warn("PrepareRequestReceived VerifySignature failed.", err)
		ds.context = backupContext
		ds.RequestChangeView()
		return
	}

	ds.context.Signatures = make([][]byte, len(ds.context.BookKeepers))
	ds.context.Signatures[payload.BookKeeperIndex] = message.Signature

	//check if the transactions received are verified. If it already exists in transaction pool
	//then no need to verify it again. Otherwise, verify it.
	if err := ds.poolActor.VerifyBlock(ds.context.Transactions, ds.context.Height-1); err != nil {
		log.Error("PrepareRequestReceived new transaction verification failed, will not sent Prepare Response", err)
		ds.context = backupContext
		ds.RequestChangeView()

		return
	}

	ds.context.NextBookKeepers, err = vote.GetValidators(ds.context.Transactions)
	if err != nil {
		ds.context = backupContext
		log.Error("[PrepareRequestReceived] GetValidators failed")
		return
	}
	ds.context.NextBookKeeper, err = core.AddressFromBookKeepers(ds.context.NextBookKeepers)
	if err != nil {
		ds.context = backupContext
		log.Error("[PrepareRequestReceived] GetBookKeeperAddress failed")
		return
	}

	if ds.context.NextBookKeeper != message.NextBookKeeper {
		ds.context = backupContext
		ds.RequestChangeView()
		log.Error("[PrepareRequestReceived] Unmatched NextBookKeeper")
		return
	}

	log.Info("send prepare response")
	ds.context.State |= SignatureSent

	if ds.context.BookKeeperIndex == -1 {
		log.Error("[DbftService] GetAccount failed")
		return
	}

	signature, err := crypto.Sign(ds.Account.PrivKey(), buf.Bytes())
	if err != nil {
		log.Error("[DbftService] SignBySigner failed")
		return
	}
	ds.context.Signatures[ds.context.BookKeeperIndex] = signature

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
	buf := new(bytes.Buffer)
	header.SerializeUnsigned(buf)
	err := crypto.Verify(*ds.context.BookKeepers[payload.BookKeeperIndex], buf.Bytes(), message.Signature)
	if err != nil {
		return
	}

	ds.context.Signatures[payload.BookKeeperIndex] = message.Signature
	err = ds.CheckSignatures()
	if err != nil {
		log.Error("CheckSignatures failed")
		return
	}
	log.Info("Prepare Response finished")
}

func (ds *DbftService) BlockSignaturesReceived(payload *msg.ConsensusPayload, message *BlockSignatures) {
	log.Info(fmt.Sprintf("BlockSignatures Received: height=%d View=%d index=%d", payload.Height, message.ViewNumber(), payload.BookKeeperIndex))

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

	buf := new(bytes.Buffer)
	header.SerializeUnsigned(buf)

	for i := 0; i < len(message.Signatures); i++ {
		sigdata := message.Signatures[i]

		if ds.context.Signatures[sigdata.Index] != nil {
			continue
		}

		err := crypto.Verify(*ds.context.BookKeepers[sigdata.Index], buf.Bytes(), sigdata.Signature)
		if err != nil {
			continue
		}

		ds.context.Signatures[sigdata.Index] = sigdata.Signature
		if ds.context.GetSignaturesCount() >= ds.context.M() {
			log.Info("BlockSignatures got enough signatures")
			break
		}
	}

	err := ds.CheckSignatures()
	if err != nil {
		log.Error("CheckSignatures failed")
		return
	}
	log.Info("BlockSignatures finished")
}

func (ds *DbftService) RefreshPolicy() {
	log.Debug()
	//con.DefaultPolicy.Refresh()
}

func (ds *DbftService) RequestChangeView() {
	if ds.context.State.HasFlag(BlockGenerated) {
		return
	}
	// FIXME if there is no save block notifcation, when the timeout call this function it will crash
	if ds.context.ViewNumber > ds.context.ExpectedView[ds.context.BookKeeperIndex] {
		ds.context.ExpectedView[ds.context.BookKeeperIndex] = ds.context.ViewNumber + 1
	} else {
		ds.context.ExpectedView[ds.context.BookKeeperIndex] += 1
	}
	log.Info(fmt.Sprintf("Request change view: height=%d View=%d nv=%d state=%s", ds.context.Height,
		ds.context.ViewNumber, ds.context.ExpectedView[ds.context.BookKeeperIndex], ds.context.GetStateDetail()))

	ds.timer.Stop()
	ds.timer.Reset(genesis.GenBlockTime << (ds.context.ExpectedView[ds.context.BookKeeperIndex] + 1))

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

	context := contract.NewContractContext(payload)

	if prohash[0] != ds.Account.ProgramHash {
		log.Error("[SignAndRelay] wrong program hash")
	}

	sig, _ := signature.SignBySigner(context.Data, ds.Account)
	ct, _ := contract.CreateSignatureContract(ds.Account.PublicKey)
	context.AddContract(ct, ds.Account.PublicKey, sig)

	prog := context.GetPrograms()
	if prog == nil {
		log.Warn("[SignAndRelay] Get program failure")
	}
	payload.SetPrograms(prog)
	ds.localNet.Xmit(payload)
}

func (ds *DbftService) Start() error {
	log.Debug()
	ds.started = true

	if config.Parameters.GenBlockTime > config.MINGENBLOCKTIME {
		genesis.GenBlockTime = time.Duration(config.Parameters.GenBlockTime) * time.Second
	} else {
		log.Warn("The Generate block time should be longer than 2 seconds, so set it to be default 6 seconds.")
	}

	ds.blockPersistCompletedSubscriber = ledgerevent.DefLedgerEvt.Subscribe(events.EventBlockPersistCompleted, ds.BlockPersistCompleted)
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
			header, err :=ledger.DefLedger.GetHeaderByHash(&ds.context.PrevHash)
			if err != nil {
				log.Error("[Timeout] GetHeader error:", err)
			}
			//set context Timestamp
			blockTime := header.Timestamp + 1
			if blockTime > now {
				ds.context.Timestamp = blockTime
			} else {
				ds.context.Timestamp = now
			}

			ds.context.Nonce = GetNonce()
			txs := ds.poolActor.GetTxnPool(true, ds.context.Height-1)
			// todo : fix feesum calcuation
			feeSum := Fixed64(0)

			// TODO: increment checking txs

			txBookkeeping := ds.CreateBookkeepingTransaction(ds.context.Nonce, feeSum)
			//add book keeping transaction first
			ds.context.Transactions = append(ds.context.Transactions, txBookkeeping)
			//add transactions from transaction pool
			for _, txEntry := range txs {
				ds.context.Transactions = append(ds.context.Transactions, txEntry.Tx)
			}
			ds.context.NextBookKeepers, err = vote.GetValidators(ds.context.Transactions)
			if err != nil {
				log.Error("[Timeout] GetValidators failed", err.Error())
				return
			}
			ds.context.NextBookKeeper, err = core.AddressFromBookKeepers(ds.context.NextBookKeepers)
			if err != nil {
				log.Error("[Timeout] GetBookKeeperAddress failed")
				return
			}
			ds.context.header = nil
			//build block and sign
			block := ds.context.MakeHeader()
			buf := new(bytes.Buffer)
			block.SerializeUnsigned(buf)
			ds.context.Signatures[ds.context.BookKeeperIndex], _ = crypto.Sign(ds.Account.PrivKey(), buf.Bytes())
		}
		payload := ds.context.MakePrepareRequest()
		ds.SignAndRelay(payload)
		ds.timer.Stop()
		ds.timer.Reset(genesis.GenBlockTime << (ds.timeView + 1))
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
