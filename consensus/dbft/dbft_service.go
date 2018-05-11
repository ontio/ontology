/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package dbft

import (
	"bytes"
	"fmt"
	"reflect"
	"time"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	actorTypes "github.com/ontio/ontology/consensus/actor"
	"github.com/ontio/ontology/core/genesis"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/signature"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/core/vote"
	"github.com/ontio/ontology/events"
	"github.com/ontio/ontology/events/message"
	p2pmsg "github.com/ontio/ontology/p2pserver/message/types"
	"github.com/ontio/ontology/validator/increment"
)

type DbftService struct {
	context           ConsensusContext
	Account           *account.Account
	timer             *time.Timer
	timerHeight       uint32
	timeView          byte
	blockReceivedTime time.Time
	started           bool
	ledger            *ledger.Ledger
	incrValidator     *increment.IncrementValidator
	poolActor         *actorTypes.TxPoolActor
	p2p               *actorTypes.P2PActor

	pid *actor.PID
	sub *events.ActorSubscriber
}

func NewDbftService(bkAccount *account.Account, txpool, p2p *actor.PID) (*DbftService, error) {

	service := &DbftService{
		Account:       bkAccount,
		timer:         time.NewTimer(time.Second * 15),
		started:       false,
		ledger:        ledger.DefLedger,
		incrValidator: increment.NewIncrementValidator(10),
		poolActor:     &actorTypes.TxPoolActor{Pool: txpool},
		p2p:           &actorTypes.P2PActor{P2P: p2p},
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
		this.incrValidator.Clean()
		this.halt()
	case *actorTypes.TimeOut:
		log.Info("dbft receive timeout")
		this.Timeout()
	case *message.SaveBlockCompleteMsg:
		log.Infof("dbft actor receives block complete event. block height=%d, numtx=%d",
			msg.Block.Header.Height, len(msg.Block.Transactions))
		this.incrValidator.AddBlock(msg.Block)
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
	self.p2p.Broadcast(block.Hash())

	self.blockReceivedTime = time.Now()

	self.InitializeConsensus(0)
}

func (ds *DbftService) BlockPersistCompleted(v interface{}) {
	if block, ok := v.(*types.Block); ok {
		log.Infof("persist block: %x", block.Hash())

		ds.p2p.Broadcast(block.Hash())
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
		for i, j := 0, 0; i < len(ds.context.Bookkeepers) && j < ds.context.M(); i++ {
			if ds.context.Signatures[i] != nil {
				sig := ds.context.Signatures[i]
				sigs[j].Index = uint16(i)
				sigs[j].Signature = sig

				block.Header.SigData = append(block.Header.SigData, sig)
				j++
			}
		}

		block.Header.Bookkeepers = ds.context.Bookkeepers

		//fill transactions
		block.Transactions = ds.context.Transactions

		hash := block.Hash()
		isExist, err := ds.ledger.IsContainBlock(hash)
		if err != nil {
			log.Errorf("DefLedger.IsContainBlock Hash:%x error:%s", hash, err)
			return err
		}
		if !isExist {
			// save block
			err := ds.ledger.AddBlock(block)
			if err != nil {
				return fmt.Errorf("CheckSignatures DefLedgerPid.RequestFuture Height:%d error:%s", block.Header.Height, err)
			}

			ds.context.State |= BlockGenerated
			payload := ds.context.MakeBlockSignatures(sigs)
			ds.SignAndRelay(payload)
		}
	}
	return nil
}

func (ds *DbftService) ChangeViewReceived(payload *p2pmsg.ConsensusPayload, message *ChangeView) {
	log.Debug()
	log.Info(fmt.Sprintf("Change View Received: height=%d View=%d index=%d nv=%d", payload.Height, message.ViewNumber(), payload.BookkeeperIndex, message.NewViewNumber))

	if message.NewViewNumber <= ds.context.ExpectedView[payload.BookkeeperIndex] {
		return
	}

	ds.context.ExpectedView[payload.BookkeeperIndex] = message.NewViewNumber

	ds.CheckExpectedView(message.NewViewNumber)
}

func (ds *DbftService) halt() error {
	log.Info("DBFT Stop")
	if ds.timer != nil {
		ds.timer.Stop()
	}

	if ds.started {
		ds.sub.Unsubscribe(message.TOPIC_SAVE_BLOCK_COMPLETE)
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

	if ds.context.BookkeeperIndex < 0 {
		log.Info("You aren't bookkeeper")
		return nil
	}

	if ds.context.BookkeeperIndex == int(ds.context.PrimaryIndex) {

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
	if inventory, ok := v.(common.Inventory); ok {
		if inventory.Type() == common.CONSENSUS {
			payload, ret := inventory.(*p2pmsg.ConsensusPayload)
			if ret == true {
				ds.NewConsensusPayload(payload)
			}
		}
	}
}

func (ds *DbftService) NewConsensusPayload(payload *p2pmsg.ConsensusPayload) {
	//if payload from current peer, ignore it
	if int(payload.BookkeeperIndex) == ds.context.BookkeeperIndex {
		return
	}

	//if payload is not same height with current contex, ignore it
	if payload.Version != ContextVersion || payload.PrevHash != ds.context.PrevHash || payload.Height != ds.context.Height {
		log.Debug("unmatched height")
		return
	}

	if ds.context.State.HasFlag(BlockGenerated) {
		log.Debug("has flag 'BlockGenerated'")
		return
	}

	if int(payload.BookkeeperIndex) >= len(ds.context.Bookkeepers) {
		log.Debug("bookkeeper index out of range")
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
	default:
		log.Warn("unknown consensus message type")
	}
}

func (ds *DbftService) PrepareRequestReceived(payload *p2pmsg.ConsensusPayload, message *PrepareRequest) {
	log.Info(fmt.Sprintf("Prepare Request Received: height=%d View=%d index=%d tx=%d", payload.Height, message.ViewNumber(), payload.BookkeeperIndex, len(message.Transactions)))

	if !ds.context.State.HasFlag(Backup) || ds.context.State.HasFlag(RequestReceived) {
		return
	}

	if uint32(payload.BookkeeperIndex) != ds.context.PrimaryIndex {
		return
	}

	header, err := ds.ledger.GetHeaderByHash(ds.context.PrevHash)
	if err != nil {
		log.Errorf("PrepareRequestReceived GetHeader failed with ds.context.PrevHash:%x", ds.context.PrevHash)
		return
	}
	if header == nil {
		log.Errorf("PrepareRequestReceived cannot GetHeaderByHash by PrevHash:%x", ds.context.PrevHash)
		return
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
	ds.context.NextBookkeeper = message.NextBookkeeper
	ds.context.Transactions = message.Transactions
	ds.context.header = nil

	blockHash := ds.context.MakeHeader().Hash()
	err = signature.Verify(ds.context.Bookkeepers[payload.BookkeeperIndex], blockHash[:], message.Signature)
	if err != nil {
		log.Warn("PrepareRequestReceived VerifySignature failed.", err)
		ds.context = backupContext
		ds.RequestChangeView()
		return
	}

	ds.context.Signatures = make([][]byte, len(ds.context.Bookkeepers))
	ds.context.Signatures[payload.BookkeeperIndex] = message.Signature

	if len(ds.context.Transactions) > 0 {
		height := ds.context.Height - 1
		start, end := ds.incrValidator.BlockRange()

		validHeight := height
		if height+1 == end {
			validHeight = start
		} else {
			ds.incrValidator.Clean()
			log.Infof("incr validator block height %v != ledger block height %v", int(end)-1, height)
		}

		if err := ds.poolActor.VerifyBlock(ds.context.Transactions, validHeight); err != nil {
			log.Error("PrepareRequestReceived new transaction verification failed, will not sent Prepare Response", err)
			ds.context = backupContext
			ds.RequestChangeView()

			return
		}

		for _, tx := range ds.context.Transactions {
			if err := ds.incrValidator.Verify(tx, validHeight); err != nil {
				log.Error("PrepareRequestReceived new transaction increment verification failed, will not sent Prepare Response", err)
				ds.context = backupContext
				ds.RequestChangeView()
				return
			}
		}
	}

	ds.context.NextBookkeepers, err = vote.GetValidators(ds.context.Transactions)
	if err != nil {
		ds.context = backupContext
		log.Error("[PrepareRequestReceived] GetValidators failed")
		return
	}
	ds.context.NextBookkeeper, err = types.AddressFromBookkeepers(ds.context.NextBookkeepers)
	if err != nil {
		ds.context = backupContext
		log.Error("[PrepareRequestReceived] GetBookkeeperAddress failed")
		return
	}

	if ds.context.NextBookkeeper != message.NextBookkeeper {
		ds.context = backupContext
		ds.RequestChangeView()
		log.Error("[PrepareRequestReceived] Unmatched NextBookkeeper")
		return
	}

	log.Info("send prepare response")
	ds.context.State |= SignatureSent

	if ds.context.BookkeeperIndex == -1 {
		log.Error("[DbftService] GetAccount failed")
		return
	}

	sig, err := signature.Sign(ds.Account, blockHash[:])
	if err != nil {
		log.Error("[DbftService] signing failed")
		return
	}
	ds.context.Signatures[ds.context.BookkeeperIndex] = sig

	payload = ds.context.MakePrepareResponse(ds.context.Signatures[ds.context.BookkeeperIndex])
	ds.SignAndRelay(payload)

	log.Info("Prepare Request finished")
}

func (ds *DbftService) PrepareResponseReceived(payload *p2pmsg.ConsensusPayload, message *PrepareResponse) {
	log.Info(fmt.Sprintf("Prepare Response Received: height=%d View=%d index=%d", payload.Height, message.ViewNumber(), payload.BookkeeperIndex))

	if ds.context.State.HasFlag(BlockGenerated) {
		return
	}

	//if the signature already exist, needn't handle again
	if ds.context.Signatures[payload.BookkeeperIndex] != nil {
		return
	}

	header := ds.context.MakeHeader()
	if header == nil {
		return
	}
	blockHash := header.Hash()
	err := signature.Verify(ds.context.Bookkeepers[payload.BookkeeperIndex], blockHash[:], message.Signature)
	if err != nil {
		return
	}

	ds.context.Signatures[payload.BookkeeperIndex] = message.Signature
	err = ds.CheckSignatures()
	if err != nil {
		log.Error("CheckSignatures failed", err)
		return
	}
	log.Info("Prepare Response finished")
}

func (ds *DbftService) BlockSignaturesReceived(payload *p2pmsg.ConsensusPayload, message *BlockSignatures) {
	log.Info(fmt.Sprintf("BlockSignatures Received: height=%d View=%d index=%d", payload.Height, message.ViewNumber(), payload.BookkeeperIndex))

	if ds.context.State.HasFlag(BlockGenerated) {
		return
	}

	//if the signature already exist, needn't handle again
	if ds.context.Signatures[payload.BookkeeperIndex] != nil {
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

		err := signature.Verify(ds.context.Bookkeepers[sigdata.Index], blockHash[:], sigdata.Signature)
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
}

func (ds *DbftService) RequestChangeView() {
	if ds.context.State.HasFlag(BlockGenerated) {
		return
	}
	// FIXME if there is no save block notifcation, when the timeout call this function it will crash
	if ds.context.ViewNumber > ds.context.ExpectedView[ds.context.BookkeeperIndex] {
		ds.context.ExpectedView[ds.context.BookkeeperIndex] = ds.context.ViewNumber + 1
	} else {
		ds.context.ExpectedView[ds.context.BookkeeperIndex] += 1
	}
	log.Info(fmt.Sprintf("Request change view: height=%d View=%d nv=%d state=%s", ds.context.Height,
		ds.context.ViewNumber, ds.context.ExpectedView[ds.context.BookkeeperIndex], ds.context.GetStateDetail()))

	ds.timer.Stop()
	ds.timer.Reset(genesis.GenBlockTime << (ds.context.ExpectedView[ds.context.BookkeeperIndex] + 1))

	ds.SignAndRelay(ds.context.MakeChangeView())
	ds.CheckExpectedView(ds.context.ExpectedView[ds.context.BookkeeperIndex])
}

func (ds *DbftService) SignAndRelay(payload *p2pmsg.ConsensusPayload) {
	buf := new(bytes.Buffer)
	payload.SerializeUnsigned(buf)
	payload.Signature, _ = signature.Sign(ds.Account, buf.Bytes())

	ds.p2p.Broadcast(payload)
}

func (ds *DbftService) start() {
	log.Debug()
	ds.started = true

	if config.DefConfig.Genesis.DBFT.GenBlockTime > config.MIN_GEN_BLOCK_TIME {
		genesis.GenBlockTime = time.Duration(config.DefConfig.Genesis.DBFT.GenBlockTime) * time.Second
	} else {
		log.Warn("The Generate block time should be longer than 2 seconds, so set it to be default 6 seconds.")
	}

	ds.sub.Subscribe(message.TOPIC_SAVE_BLOCK_COMPLETE)

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
			header, err := ds.ledger.GetHeaderByHash(ds.context.PrevHash)
			if err != nil {
				log.Errorf("[Timeout] GetHeader PrevHash:%x error:%s", ds.context.PrevHash, err)
				return
			}
			if header == nil {
				log.Errorf("[Timeout] cannot GetHeaderByHash by PrevHash:%x", ds.context.PrevHash)
				return
			}
			//set context Timestamp
			blockTime := header.Timestamp + 1
			if blockTime > now {
				ds.context.Timestamp = blockTime
			} else {
				ds.context.Timestamp = now
			}

			ds.context.Nonce = common.GetNonce()

			height := ds.context.Height - 1
			validHeight := height

			start, end := ds.incrValidator.BlockRange()

			if height+1 == end {
				validHeight = start
			} else {
				ds.incrValidator.Clean()
				log.Infof("incr validator block height %v != ledger block height %v", int(end)-1, height)
			}

			log.Infof("current block height %v, increment validator block cache range: [%d, %d)", height, start, end)
			txs := ds.poolActor.GetTxnPool(true, validHeight)

			transactions := make([]*types.Transaction, 0, len(txs))
			for _, txEntry := range txs {
				// TODO optimize to use height in txentry
				if err := ds.incrValidator.Verify(txEntry.Tx, validHeight); err == nil {
					transactions = append(transactions, txEntry.Tx)
				}
			}

			ds.context.Transactions = transactions

			ds.context.NextBookkeepers, err = vote.GetValidators(ds.context.Transactions)
			if err != nil {
				log.Error("[Timeout] GetValidators failed", err.Error())
				return
			}
			ds.context.NextBookkeeper, err = types.AddressFromBookkeepers(ds.context.NextBookkeepers)
			if err != nil {
				log.Error("[Timeout] GetBookkeeperAddress failed")
				return
			}
			ds.context.header = nil
			//build block and sign
			block := ds.context.MakeHeader()
			blockHash := block.Hash()
			ds.context.Signatures[ds.context.BookkeeperIndex], _ = signature.Sign(ds.Account, blockHash[:])
		}
		payload := ds.context.MakePrepareRequest()
		ds.SignAndRelay(payload)
		ds.timer.Stop()
		ds.timer.Reset(genesis.GenBlockTime << (ds.timeView + 1))
	} else if (ds.context.State.HasFlag(Primary) && ds.context.State.HasFlag(RequestSent)) || ds.context.State.HasFlag(Backup) {
		ds.RequestChangeView()
	}
}
