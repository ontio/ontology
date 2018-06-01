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

package p2pserver

import (
	"bytes"
	"fmt"
	"time"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/common/serialization"
	vconfig "github.com/ontio/ontology/consensus/vbft/config"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/signature"
	"github.com/ontio/ontology/core/types"
	actor "github.com/ontio/ontology/p2pserver/actor/req"
	msgCom "github.com/ontio/ontology/p2pserver/common"
	mt "github.com/ontio/ontology/p2pserver/message/types"
)

const (
	MSG_CACHE = 1000
	TIME_OUT  = 60
)

type emergencyGov struct {
	account            *account.Account          // local account to sign data
	emergencyMsgCh     chan *msgCom.EmergencyMsg // The emergency governance message queue
	context            *emergencyGovContext      // Emergency goverance context
	server             *P2PServer                // Pointer to the local node
	blkSyncCh          chan struct{}             // The block sync mgr use it to notify
	stopCh             chan struct{}             // Stop emergency governance loop
	timerEvt           chan struct{}             // Time out event
	emgBlkCompletedEvt chan struct{}             // The emergency goverance block completed event
}

// NewEmergencyGov returns a new instance of emergency governance
func NewEmergencyGov(server *P2PServer, account *account.Account) *emergencyGov {
	emergencyGov := &emergencyGov{
		account: account,
		context: &emergencyGovContext{},
		server:  server,
	}
	emergencyGov.init()
	return emergencyGov
}

// init intializes an emergency governance
func (this *emergencyGov) init() {
	this.emergencyMsgCh = make(chan *msgCom.EmergencyMsg, MSG_CACHE)
	this.blkSyncCh = make(chan struct{}, 1)
	this.context.reset()
	this.stopCh = make(chan struct{})
	this.timerEvt = make(chan struct{}, 1)
	this.emgBlkCompletedEvt = make(chan struct{}, 1)
}

// Start starts an emergency governance loop
func (this *emergencyGov) Start() {

	for {
		select {
		case msg, ok := <-this.emergencyMsgCh:
			if ok {
				this.handleEmergencyMsg(msg)
			}
		case <-this.emgBlkCompletedEvt:
			this.handleEmergencyBlockCompletedEvt()
		case <-this.stopCh:
			return
		case <-this.timerEvt:
			this.context.reset()
		}
	}
}

// Stop stops an emergency governance
func (this *emergencyGov) Stop() {
	if this.stopCh != nil {
		this.stopCh <- struct{}{}
	}

	if this.emergencyMsgCh != nil {
		close(this.emergencyMsgCh)
	}

	if this.timerEvt != nil {
		close(this.timerEvt)
	}
	if this.emgBlkCompletedEvt != nil {
		close(this.emgBlkCompletedEvt)
	}
}

// handleEmergencyMsg dispatch the msg to the msg handler
func (this *emergencyGov) handleEmergencyMsg(msg *msgCom.EmergencyMsg) {
	switch msg.MsgType {
	case msgCom.EmergencyReq:
		emergencyRequest := msg.Content.(*mt.EmergencyActionRequest)
		this.EmergencyActionRequestReceived(emergencyRequest)
	case msgCom.EmergencyRsp:
		emergencyResponse := msg.Content.(*mt.EmergencyActionResponse)
		this.EmergencyActionResponseReceived(emergencyResponse)
	case msgCom.EmergencyAdminStart:
		emergencyRequest := msg.Content.(*mt.EmergencyActionRequest)
		this.startEmergencyGov(emergencyRequest)
	default:
		log.Infof("handleEmergencyMsg: unknown msg type %d", msg.MsgType)
	}
}

func (this *emergencyGov) handleEmergencyBlockCompletedEvt() {
	if this.context.getStatus() == EmergencyGovComplete {
		return
	}

	log.Tracef("handleEmergencyBlockCompletedEvt")
	this.context.setStatus(EmergencyGovComplete)

	// notify consensus and block sync mgr to recover
	cmd := &msgCom.EmergencyGovCmd{
		Pause:  false,
		Height: this.context.getEmergencyGovHeight(),
	}
	actor.NotifyEmergencyGovCmd(cmd)
	this.server.notifyEmergencyGovCmd(cmd)

	if !this.context.timer.Stop() {
		<-this.context.timer.C
	}
	this.context.done <- struct{}{}
}

// EmergencyActionResponseReceived handles an emergency governance response from network
func (this *emergencyGov) EmergencyActionResponseReceived(msg *mt.EmergencyActionResponse) {
	// Todo: Check whether local node  supports emergency governance policy
	log.Trace("EmergencyActionResponseReceived: receive emergency governance response")

	if this.context.getStatus() == EmergencyGovComplete {
		return
	}

	id, err := vconfig.PubkeyID(msg.PubKey)
	if err != nil {
		log.Errorf("failed to get id from public key: %v", msg.PubKey)
		return
	}

	if this.context.getSig(id) != nil {
		return
	}

	rspHash := msg.Hash()
	err = signature.Verify(msg.PubKey, rspHash[:], msg.RspSig)
	if err != nil {
		log.Errorf("failed to verify response signature %v. PubKey %v rspHash %x rspSig %x",
			err, msg.PubKey, rspHash, msg.RspSig)
		return
	}
	if this.context.EmergencyReqCache == nil {
		this.context.appendEmergencyRsp(id, msg)
		return
	}

	block := this.context.getEmergencyBlock()
	if block == nil {
		return
	}
	blockHash := block.Hash()
	err = signature.Verify(msg.PubKey, blockHash[:], msg.SigOnBlk)
	if err != nil {
		log.Errorf("failed to verify block hash signature %v. PubKey %v blockHash %x SigOnBlk %x",
			err, msg.PubKey, blockHash, msg.SigOnBlk)
		return
	}

	this.context.setSig(id, msg.SigOnBlk)

	this.checkSignatures()
}

// checkSignatures checks whether the signatures reaches the threshold 2/3
func (this *emergencyGov) checkSignatures() {
	if this.context.getSignatureCount() >= this.context.threshold() {
		block := this.context.getEmergencyBlock()
		if block == nil {
			return
		}

		for id, sig := range this.context.Signatures {
			pubkey, _ := id.Pubkey()
			block.Header.Bookkeepers = append(block.Header.Bookkeepers, pubkey)
			block.Header.SigData = append(block.Header.SigData, sig)
		}

		contained, err := ledger.DefLedger.IsContainBlock(block.Hash())
		if err != nil {
			log.Errorf("checkSignatures: hash %x, error %v", block.Hash(), err)
			return
		}

		if !contained {
			err := ledger.DefLedger.AddBlock(block)
			if err != nil {
				log.Tracef("DefLedger add block failed. err %v", err)
				return
			}
			this.server.Xmit(block.Hash())
		}

		this.context.setStatus(EmergencyGovComplete)

		// notify consensus and block sync mgr to recover
		cmd := &msgCom.EmergencyGovCmd{
			Pause:  false,
			Height: block.Header.Height,
		}
		actor.NotifyEmergencyGovCmd(cmd)
		this.server.notifyEmergencyGovCmd(cmd)

		if !this.context.timer.Stop() {
			<-this.context.timer.C
		}
		this.context.done <- struct{}{}
	}
}

// checkEvidence checks whether the evidence is valid
func (this *emergencyGov) checkEvidence(evd mt.EmergencyEvidence) bool {
	// Todo: check evidence
	return true
}

// checkBlock checks whether the block is valid
func (this *emergencyGov) checkBlock(block *types.Block) bool {
	// 1. Check payload
	payload := block.Header.ConsensusPayload
	if payload == nil {
		//Todo:
		return false
	}

	curHeight := ledger.DefLedger.GetCurrentHeaderHeight()

	if curHeight >= block.Header.Height {
		log.Errorf("emergency governance height %d is less than current height %d",
			block.Header.Height, curHeight)
		return false
	}

	if curHeight < block.Header.Height-1 {
		log.Tracef("Waiting for block sync mgr to sync block till emergency goverance height, curHeight %d",
			block.Header.Height, curHeight)
		<-this.blkSyncCh
	}

	log.Tracef("checkBlock: block height %d, prevBlockHash %x", block.Header.Height, block.Header.PrevBlockHash)
	tmpBlk, err := ledger.DefLedger.GetBlockByHash(block.Header.PrevBlockHash)
	if err != nil || tmpBlk == nil || tmpBlk.Header == nil {
		log.Trace("Can't get block by hash: ", block.Header.PrevBlockHash)
		return false
	}

	return true
}

// checkReqSignature checks whether the request signature valid
func (this *emergencyGov) checkReqSignature(msg *mt.EmergencyActionRequest) bool {
	buf := new(bytes.Buffer)
	buf.Write([]byte{byte(msg.Reason), byte(msg.Evidence)})
	serialization.WriteUint32(buf, msg.ProposalBlkNum)
	msg.ProposalBlk.Serialize(buf)
	serialization.WriteVarBytes(buf, keypair.SerializePublicKey(msg.ProposerPK))
	serialization.WriteVarBytes(buf, msg.ProposerSigOnBlk)

	for _, sig := range msg.AdminSigs {
		m := int(sig.M)
		kn := len(sig.PubKeys)
		sn := len(sig.SigData)

		if kn > 24 || sn < m || m > kn {
			log.Errorf("wrong emergency governance sig param length")
			return false
		}

		if kn == 1 {
			err := signature.Verify(sig.PubKeys[0], buf.Bytes(), sig.SigData[0])
			if err != nil {
				log.Errorf("signature verification failed. %v", err)
				return false
			}
		} else {
			if err := signature.VerifyMultiSignature(buf.Bytes(), sig.PubKeys, m, sig.SigData); err != nil {
				log.Errorf("multi-signature failed. %v", err)
				return false
			}
		}
	}
	return true
}

// validatePendingRspMsg validates the emergency governance responses in cache
func (this *emergencyGov) validatePendingRspMsg() {
	if len(this.context.EmergencyRspCache) == 0 {
		return
	}
	block := this.context.getEmergencyBlock()
	if block == nil {
		return
	}
	blockHash := block.Hash()
	for id, msg := range this.context.EmergencyRspCache {
		err := signature.Verify(msg.PubKey, blockHash[:], msg.SigOnBlk)
		if err != nil {
			continue
		}

		this.context.setSig(id, msg.SigOnBlk)

		this.checkSignatures()
	}
	this.context.clearEmergencyRspCache()
}

// EmergencyActionRequestReceived handles an emergency governance request from network
func (this *emergencyGov) EmergencyActionRequestReceived(msg *mt.EmergencyActionRequest) error {
	// Todo: check whether local node support emergency governance
	log.Tracef("EmergencyActionRequestReceived: receive emergency governance request at height %d", msg.ProposalBlkNum)

	if this.context != nil && this.context.getStatus() == EmergencyGovStart {
		return fmt.Errorf("EmergencyActionRequestReceived: emergency governacne started")
	}

	// 1. Validate evidence
	if !this.checkEvidence(msg.Evidence) {
		return fmt.Errorf("EmergencyActionRequestReceived: checkEvidence failed %v",
			msg.Evidence)
	}

	// 2. Validate block
	if !this.checkBlock(msg.ProposalBlk) {
		return fmt.Errorf("EmergencyActionRequestReceived: checkBlock failed")
	}

	// 3. Validate admin signature
	if !this.checkReqSignature(msg) {
		return fmt.Errorf("EmergencyActionRequestReceived: checkSignature failed")
	}

	// Todo: 4. Validate admin pubkey

	peers, err := getPeers()
	if err != nil {
		return fmt.Errorf("EmergencyActionRequestReceived: failed to get peers. %v", err)
	}
	this.context.setPeers(peers)
	this.context.setStatus(EmergencyGovStart)
	this.context.setEmergencyReqCache(msg)
	this.context.setEmergencyGovHeight(msg.ProposalBlkNum)

	// notify consensus and block sync mgr to pause
	cmd := &msgCom.EmergencyGovCmd{
		Pause:  true,
		Height: msg.ProposalBlkNum,
	}
	actor.NotifyEmergencyGovCmd(cmd)
	this.server.notifyEmergencyGovCmd(cmd)

	// Broadcast the response
	response, err := this.constructEmergencyActionResponse(msg.ProposalBlk)
	if err != nil {
		return err
	}

	pubkey := this.server.GetPubKey()
	id, _ := vconfig.PubkeyID(pubkey)
	this.context.setSig(id, response.SigOnBlk)

	log.Trace("EmergencyActionRequestReceived: broadcast response")
	this.server.Xmit(response)

	this.validatePendingRspMsg()

	this.context.timer = time.NewTimer(TIME_OUT * time.Second)
	go this.emergencyTimer()

	return nil
}

// signBlock signs the block
func (this *emergencyGov) signBlock(block *types.Block) []byte {
	blockHash := block.Hash()
	sigOnBlk, _ := signature.Sign(this.account, blockHash[:])
	return sigOnBlk
}

// constructEmergencyActionResponse constructs an emergency governance response with the block
func (this *emergencyGov) constructEmergencyActionResponse(block *types.Block) (*mt.EmergencyActionResponse, error) {
	rsp := &mt.EmergencyActionResponse{
		PubKey: this.account.PublicKey,
	}
	rsp.SigOnBlk = this.signBlock(block)

	hash := rsp.Hash()
	rsp.RspSig, _ = signature.Sign(this.account, hash[:])
	return rsp, nil
}

// startEmergencyGov starts an new emergency governance introduced by admin
func (this *emergencyGov) startEmergencyGov(msg *mt.EmergencyActionRequest) {
	log.Tracef("startEmergencyGov: receive emergency governance admin request at height %d", msg.ProposalBlkNum)
	if this.context != nil && this.context.getStatus() == EmergencyGovStart {
		log.Info("startEmergencyGov: local node is in emergency governance progress")
		return
	}
	// 1. Validate evidence
	if !this.checkEvidence(msg.Evidence) {
		log.Errorf("EmergencyActionRequestReceived: checkEvidence failed %v",
			msg.Evidence)
		return
	}

	// 2. Validate block
	if !this.checkBlock(msg.ProposalBlk) {
		log.Errorf("EmergencyActionRequestReceived: checkBlock failed")
		return
	}

	// 3. Validate admin signature
	if !this.checkReqSignature(msg) {
		log.Errorf("EmergencyActionRequestReceived: checkSignature failed")
		return
	}

	// Todo: 4. Validate admin pubkey

	// notify consensus and block sync mgr to pause
	cmd := &msgCom.EmergencyGovCmd{
		Pause:  true,
		Height: msg.ProposalBlkNum,
	}

	actor.NotifyEmergencyGovCmd(cmd)
	this.server.notifyEmergencyGovCmd(cmd)

	this.context.reset()
	this.context.setStatus(EmergencyGovStart)
	this.context.setEmergencyReqCache(msg)
	this.context.setEmergencyGovHeight(msg.ProposalBlkNum)

	peers, _ := getPeers()
	this.context.setPeers(peers)

	sig := this.signBlock(msg.ProposalBlk)
	pubkey := this.server.GetPubKey()
	id, _ := vconfig.PubkeyID(pubkey)
	this.context.setSig(id, sig)

	hash := msg.Hash()
	reqSig, _ := signature.Sign(this.account, hash[:])
	msg.ReqSig = reqSig
	msg.ReqPK = pubkey

	this.server.Xmit(msg)
	this.context.timer = time.NewTimer(TIME_OUT * time.Second)
	go this.emergencyTimer()
}

func (this *emergencyGov) emergencyTimer() {
	select {
	case <-this.context.timer.C:
		log.Error("emergencyTimer: emergency governance timeout")
		this.timerEvt <- struct{}{}
		return
	case <-this.context.done:
		return
	}
}
