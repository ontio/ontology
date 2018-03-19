package dbft

import (
	"bytes"
	"fmt"
	"reflect"
	"time"

	"github.com/Ontology/account"
	. "github.com/Ontology/common"
	"github.com/Ontology/common/config"
	"github.com/Ontology/common/log"
	actorTypes "github.com/Ontology/consensus/actor"
	"github.com/Ontology/core/genesis"
	"github.com/Ontology/core/ledger"
	"github.com/Ontology/core/payload"
	"github.com/Ontology/core/types"
	"github.com/Ontology/core/vote"
	"github.com/Ontology/crypto"
	ontErrors "github.com/Ontology/errors"
	"github.com/Ontology/eventbus/actor"
	"github.com/Ontology/events"
	"github.com/Ontology/events/message"
	p2pmsg "github.com/Ontology/p2pserver/message"
)

type DbftService struct {
	context           ConsensusContext
	Account           *account.Account
	timer             *time.Timer
	timerHeight       uint32
	timeView          byte
	blockReceivedTime time.Time
	started           bool
	poolActor         *actorTypes.TxPoolActor
	p2p               *actorTypes.P2PActor

	pid *actor.PID
	sub *events.ActorSubscriber
	//blockPersistCompletedSubscriber events.Subscriber
}

func NewDbftService(bkAccount *account.Account, txpool, p2p *actor.PID) (*DbftService, error) {

	service := &DbftService{
		Account:   bkAccount,
		timer:     time.NewTimer(time.Second * 15),
		started:   false,
		poolActor: &actorTypes.TxPoolActor{Pool: txpool},
		p2p:       &actorTypes.P2PActor{P2P: p2p},
	}

	if !service.timer.Stop() {
		<-service.timer.C
	}

	go func() {
		for {
			select {
			case <-service.timer.C:
				log.Debug("******Get a timeout notice")
				service.pid.Tell(&actorTypes.TimeOut{})
			}
		}
	}()

	props := actor.FromProducer(func() actor.Actor {
		return service
	})

	pid, err := actor.SpawnNamed(props, "consensus_dbft")
	service.pid = pid

	service.sub = events.NewActorSubscriber(pid)
	return service, err
}

func (this *DbftService) Receive(context actor.Context) {
	if _, ok := context.Message().(*actorTypes.StartConsensus); this.started == false && ok == false {
		return
	}

	switch msg := context.Message().(type) {
	case *actor.Restarting:
		log.Warn("dbft actor restarting")
	case *actor.Stopping:
		log.Warn("dbft actor stopping")
	case *actor.Stopped:
		log.Warn("dbft actor stopped")
	case *actor.Started:
		log.Warn("dbft actor started")
	case *actor.Restart:
		log.Warn("dbft actor restart")
	case *actorTypes.StartConsensus:
		this.start()
	case *actorTypes.StopConsensus:
		this.halt()
	case *actorTypes.TimeOut:
		log.Info("dbft receive timeout")
		this.Timeout()
	case *message.SaveBlockCompleteMsg:
		this.handleBlockPersistCompleted(msg.Block)
	case *p2pmsg.ConsensusPayload:
		this.NewConsensusPayload(msg)

	default:
		log.Info("dbft actor: Unknown msg ", msg, "type", reflect.TypeOf(msg))
	}
}

func (this *DbftService) GetPID() *actor.PID {
	return this.pid
}
func (this *DbftService) Start() error {
	this.pid.Tell(&actorTypes.StartConsensus{})
	return nil
}

func (this *DbftService) Halt() error {
	this.pid.Tell(&actorTypes.StopConsensus{})
	return nil
}

func (self *DbftService) handleBlockPersistCompleted(block *types.Block) {
	log.Infof("persist block: %x", block.Hash())
	self.p2p.Xmit(block.Hash())

	self.blockReceivedTime = time.Now()

	self.InitializeConsensus(0)
}

func (ds *DbftService) BlockPersistCompleted(v interface{}) {
	if block, ok := v.(*types.Block); ok {
		log.Infof("persist block: %x", block.Hash())

		ds.p2p.Xmit(block.Hash())
	}

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
		ds.InitializeConsensus(viewNumber)
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
		//build block
		block := ds.context.MakeHeader()
		sigs := make([]SignaturesData, ds.context.M())
		for i, j := 0, 0; i < len(ds.context.BookKeepers) && j < ds.context.M(); i++ {
			if ds.context.Signatures[i] != nil {
				sig := ds.context.Signatures[i]
				sigs[j].Index = uint16(i)
				sigs[j].Signature = sig

				block.Header.SigData = append(block.Header.SigData, sig)
				j++
			}
		}

		block.Header.BookKeepers = ds.context.BookKeepers

		//fill transactions
		block.Transactions = ds.context.Transactions

		hash := block.Hash()
		isExist, err := ledger.DefLedger.IsContainBlock(hash)
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
	//signatureRedeemScript, err := contract.CreateSignatureRedeemScript(ds.context.Owner)
	//if err != nil {
	//	return nil
	//}
	//signatureRedeemScriptHashToCodeHash := ToCodeHash(signatureRedeemScript)
	//if err != nil {
	//	return nil
	//}
	//outputs := []*utxo.TxOutput{}
	//if fee > 0 {
	//	feeOutput := &utxo.TxOutput{
	//		AssetID:     genesis.ONGTokenID,
	//		Value:       fee,
	//		Address: signatureRedeemScriptHashToCodeHash,
	//	}
	//	outputs = append(outputs, feeOutput)
	//}
	return &types.Transaction{
		TxType: types.BookKeeping,
		//PayloadVersion: payload.BookKeepingPayloadVersion,
		Payload:    bookKeepingPayload,
		Attributes: []*types.TxAttribute{},
	}
}

func (ds *DbftService) ChangeViewReceived(payload *p2pmsg.ConsensusPayload, message *ChangeView) {
	log.Debug()
	log.Info(fmt.Sprintf("Change View Received: height=%d View=%d index=%d nv=%d", payload.Height, message.ViewNumber(), payload.BookKeeperIndex, message.NewViewNumber))

	if message.NewViewNumber <= ds.context.ExpectedView[payload.BookKeeperIndex] {
		return
	}

	ds.context.ExpectedView[payload.BookKeeperIndex] = message.NewViewNumber

	ds.CheckExpectedView(message.NewViewNumber)
}

func (ds *DbftService) halt() error {
	log.Info("DBFT Stop")
	if ds.timer != nil {
		ds.timer.Stop()
	}

	if ds.started {
		ds.sub.Unsubscribe(message.TopicSaveBlockComplete)
	}
	return nil
}

func (ds *DbftService) InitializeConsensus(viewNum byte) error {
	log.Debug("[InitializeConsensus] Start InitializeConsensus.")
	log.Debug("[InitializeConsensus] viewNum: ", viewNum)

	if viewNum == 0 {
		ds.context.Reset(ds.Account)
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
			payload, ret := inventory.(*p2pmsg.ConsensusPayload)
			if ret == true {
				ds.NewConsensusPayload(payload)
			}
		}
	}
}

func (ds *DbftService) NewConsensusPayload(payload *p2pmsg.ConsensusPayload) {
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

func (ds *DbftService) PrepareRequestReceived(payload *p2pmsg.ConsensusPayload, message *PrepareRequest) {
	log.Info(fmt.Sprintf("Prepare Request Received: height=%d View=%d index=%d tx=%d", payload.Height, message.ViewNumber(), payload.BookKeeperIndex, len(message.Transactions)))

	if !ds.context.State.HasFlag(Backup) || ds.context.State.HasFlag(RequestReceived) {
		return
	}

	if uint32(payload.BookKeeperIndex) != ds.context.PrimaryIndex {
		return
	}

	header, err := ledger.DefLedger.GetHeaderByHash(ds.context.PrevHash)
	if err != nil {
		log.Info("PrepareRequestReceived GetHeader failed with ds.context.PrevHash", ds.context.PrevHash)
	}

	//TODO Add Error Catch
	prevBlockTimestamp := header.Timestamp
	if payload.Timestamp <= prevBlockTimestamp || payload.Timestamp > uint32(time.Now().Add(time.Minute*10).Unix()) {
		log.Info(fmt.Sprintf("Prepare Reques tReceived: Timestamp incorrect: %d", payload.Timestamp))
		return
	}

	if len(message.Transactions) == 0 || message.Transactions[0].TxType != types.BookKeeping {
		log.Error("PrepareRequestReceived first transaction type is not bookking")
		ds.RequestChangeView()
		return
	}

	backupContext := ds.context

	ds.context.State |= RequestReceived
	ds.context.Timestamp = payload.Timestamp
	ds.context.Nonce = message.Nonce
	ds.context.NextBookKeeper = message.NextBookKeeper
	ds.context.Transactions = message.Transactions
	ds.context.header = nil

	blockHash := ds.context.MakeHeader().Hash()
	err = crypto.Verify(*ds.context.BookKeepers[payload.BookKeeperIndex], blockHash[:], message.Signature)
	if err != nil {
		log.Warn("PrepareRequestReceived VerifySignature failed.", err)
		ds.context = backupContext
		ds.RequestChangeView()
		return
	}

	ds.context.Signatures = make([][]byte, len(ds.context.BookKeepers))
	ds.context.Signatures[payload.BookKeeperIndex] = message.Signature

	for _, tx := range ds.context.Transactions[1:] {
		if tx.TxType == types.BookKeeping {
			log.Error("PrepareRequestReceived non-first transaction type is bookking")
			ds.context = backupContext
			ds.RequestChangeView()
			return
		}
	}

	if len(ds.context.Transactions) > 1 {
		if err := ds.poolActor.VerifyBlock(ds.context.Transactions[1:], ds.context.Height-1); err != nil {
			log.Error("PrepareRequestReceived new transaction verification failed, will not sent Prepare Response", err)
			ds.context = backupContext
			ds.RequestChangeView()

			return
		}
	}

	ds.context.NextBookKeepers, err = vote.GetValidators(ds.context.Transactions)
	if err != nil {
		ds.context = backupContext
		log.Error("[PrepareRequestReceived] GetValidators failed")
		return
	}
	ds.context.NextBookKeeper, err = types.AddressFromBookKeepers(ds.context.NextBookKeepers)
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

	sign, err := crypto.Sign(ds.Account.PrivKey(), blockHash[:])
	if err != nil {
		log.Error("[DbftService] SignBySigner failed")
		return
	}
	ds.context.Signatures[ds.context.BookKeeperIndex] = sign

	payload = ds.context.MakePrepareResponse(ds.context.Signatures[ds.context.BookKeeperIndex])
	ds.SignAndRelay(payload)

	log.Info("Prepare Request finished")
}

func (ds *DbftService) PrepareResponseReceived(payload *p2pmsg.ConsensusPayload, message *PrepareResponse) {
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
	blockHash := header.Hash()
	err := crypto.Verify(*ds.context.BookKeepers[payload.BookKeeperIndex], blockHash[:], message.Signature)
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

func (ds *DbftService) BlockSignaturesReceived(payload *p2pmsg.ConsensusPayload, message *BlockSignatures) {
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

	blockHash := header.Hash()

	for i := 0; i < len(message.Signatures); i++ {
		sigdata := message.Signatures[i]

		if ds.context.Signatures[sigdata.Index] != nil {
			continue
		}

		err := crypto.Verify(*ds.context.BookKeepers[sigdata.Index], blockHash[:], sigdata.Signature)
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

func (ds *DbftService) SignAndRelay(payload *p2pmsg.ConsensusPayload) {
	buf := new(bytes.Buffer)
	payload.SerializeUnsigned(buf)
	payload.Signature, _ = crypto.Sign(ds.Account.PrivKey(), buf.Bytes())

	ds.p2p.Xmit(payload)
}

func (ds *DbftService) start() {
	log.Debug()
	ds.started = true

	if config.Parameters.GenBlockTime > config.MINGENBLOCKTIME {
		genesis.GenBlockTime = time.Duration(config.Parameters.GenBlockTime) * time.Second
	} else {
		log.Warn("The Generate block time should be longer than 2 seconds, so set it to be default 6 seconds.")
	}

	ds.sub.Subscribe(message.TopicSaveBlockComplete)

	ds.InitializeConsensus(0)
}

func (ds *DbftService) Timeout() {
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
			header, err := ledger.DefLedger.GetHeaderByHash(ds.context.PrevHash)
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
			ds.context.NextBookKeeper, err = types.AddressFromBookKeepers(ds.context.NextBookKeepers)
			if err != nil {
				log.Error("[Timeout] GetBookKeeperAddress failed")
				return
			}
			ds.context.header = nil
			//build block and sign
			block := ds.context.MakeHeader()
			blockHash := block.Hash()
			ds.context.Signatures[ds.context.BookKeeperIndex], _ = crypto.Sign(ds.Account.PrivKey(), blockHash[:])
		}
		payload := ds.context.MakePrepareRequest()
		ds.SignAndRelay(payload)
		ds.timer.Stop()
		ds.timer.Reset(genesis.GenBlockTime << (ds.timeView + 1))
	} else if (ds.context.State.HasFlag(Primary) && ds.context.State.HasFlag(RequestSent)) || ds.context.State.HasFlag(Backup) {
		ds.RequestChangeView()
	}
}
