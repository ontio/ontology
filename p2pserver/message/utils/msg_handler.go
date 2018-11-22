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

package utils

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	lru "github.com/hashicorp/golang-lru"
	evtActor "github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/types"
	actor "github.com/ontio/ontology/p2pserver/actor/req"
	msgCommon "github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/message/msg_pack"
	msgTypes "github.com/ontio/ontology/p2pserver/message/types"
	"github.com/ontio/ontology/p2pserver/net/protocol"
)

//respCache cache for some response data
var respCache *lru.ARCCache

// AddrReqHandle handles the neighbor address request from peer
func AddrReqHandle(data *msgTypes.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Trace("[p2p]receive addr request message", data.Addr, data.Id)
	remotePeer := p2p.GetPeer(data.Id)
	if remotePeer == nil {
		log.Debug("[p2p]remotePeer invalid in AddrReqHandle")
		return
	}

	var addrStr []msgCommon.PeerAddr
	addrStr = p2p.GetNeighborAddrs()
	//check mask peers
	mskPeers := config.DefConfig.P2PNode.ReservedCfg.MaskPeers
	if config.DefConfig.P2PNode.ReservedPeersOnly && len(mskPeers) > 0 {
		for i := 0; i < len(addrStr); i++ {
			var ip net.IP
			ip = addrStr[i].IpAddr[:]
			address := ip.To16().String()
			for j := 0; j < len(mskPeers); j++ {
				if address == mskPeers[j] {
					addrStr = append(addrStr[:i], addrStr[i+1:]...)
					i--
					break
				}
			}
		}

	}
	msg := msgpack.NewAddrs(addrStr)
	err := p2p.Send(remotePeer, msg, false)
	if err != nil {
		log.Warn(err)
		return
	}
}

// HeaderReqHandle handles the header sync req from peer
func HeadersReqHandle(data *msgTypes.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Trace("[p2p]receive headers request message", data.Addr, data.Id)

	headersReq := data.Payload.(*msgTypes.HeadersReq)

	startHash := headersReq.HashStart
	stopHash := headersReq.HashEnd

	headers, err := GetHeadersFromHash(startHash, stopHash)
	if err != nil {
		log.Warnf("get headers in HeadersReqHandle error: %s,startHash:%s,stopHash:%s", err.Error(), startHash.ToHexString(), stopHash.ToHexString())
		return
	}
	remotePeer := p2p.GetPeer(data.Id)
	if remotePeer == nil {
		log.Debugf("[p2p]remotePeer invalid in HeadersReqHandle, peer id: %d", data.Id)
		return
	}
	msg := msgpack.NewHeaders(headers)
	err = p2p.Send(remotePeer, msg, false)
	if err != nil {
		log.Warn(err)
		return
	}
}

//PingHandle handle ping msg from peer
func PingHandle(data *msgTypes.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Trace("[p2p]receive ping message", data.Addr, data.Id)

	ping := data.Payload.(*msgTypes.Ping)
	remotePeer := p2p.GetPeer(data.Id)
	if remotePeer == nil {
		log.Debug("[p2p]remotePeer invalid in PingHandle")
		return
	}
	remotePeer.SetHeight(ping.Height)

	height := ledger.DefLedger.GetCurrentBlockHeight()
	p2p.SetHeight(uint64(height))
	msg := msgpack.NewPongMsg(uint64(height))

	err := p2p.Send(remotePeer, msg, false)
	if err != nil {
		log.Warn(err)
	}
}

///PongHandle handle pong msg from peer
func PongHandle(data *msgTypes.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Trace("[p2p]receive pong message", data.Addr, data.Id)

	pong := data.Payload.(*msgTypes.Pong)

	remotePeer := p2p.GetPeer(data.Id)
	if remotePeer == nil {
		log.Debug("[p2p]remotePeer invalid in PongHandle")
		return
	}
	remotePeer.SetHeight(pong.Height)
}

// BlkHeaderHandle handles the sync headers from peer
func BlkHeaderHandle(data *msgTypes.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Trace("[p2p]receive block header message", data.Addr, data.Id)
	if pid != nil {
		var blkHeader = data.Payload.(*msgTypes.BlkHeader)
		input := &msgCommon.AppendHeaders{
			FromID:  data.Id,
			Headers: blkHeader.BlkHdr,
		}
		pid.Tell(input)
	}
}

// BlockHandle handles the block message from peer
func BlockHandle(data *msgTypes.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Trace("[p2p]receive block message from ", data.Addr, data.Id)

	if pid != nil {
		var block = data.Payload.(*msgTypes.Block)
		input := &msgCommon.AppendBlock{
			FromID:     data.Id,
			BlockSize:  data.PayloadSize,
			Block:      block.Blk,
			MerkleRoot: block.MerkleRoot,
		}
		pid.Tell(input)
	}
}

// ConsensusHandle handles the consensus message from peer
func ConsensusHandle(data *msgTypes.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Debugf("[p2p]receive consensus message:%v,%d", data.Addr, data.Id)

	if actor.ConsensusPid != nil {
		var consensus = data.Payload.(*msgTypes.Consensus)
		if err := consensus.Cons.Verify(); err != nil {
			log.Warn(err)
			return
		}
		consensus.Cons.PeerId = data.Id
		actor.ConsensusPid.Tell(&consensus.Cons)
	}
}

// NotFoundHandle handles the not found message from peer
func NotFoundHandle(data *msgTypes.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	var notFound = data.Payload.(*msgTypes.NotFound)
	log.Debug("[p2p]receive notFound message, hash is ", notFound.Hash)
}

// TransactionHandle handles the transaction message from peer
func TransactionHandle(data *msgTypes.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Trace("[p2p]receive transaction message", data.Addr, data.Id)

	var trn = data.Payload.(*msgTypes.Trn)
	actor.AddTransaction(trn.Txn)
	log.Trace("[p2p]receive Transaction message hash", trn.Txn.Hash())

}

// VersionHandle handles version handshake protocol from peer
func VersionHandle(data *msgTypes.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Trace("[p2p]receive version message", data.Addr, data.Id)

	version := data.Payload.(*msgTypes.Version)

	remotePeer := p2p.GetPeerFromAddr(data.Addr)
	if remotePeer == nil {
		log.Debug("[p2p]peer is not exist", data.Addr)
		//peer not exist,just remove list and return
		p2p.RemoveFromConnectingList(data.Addr)
		return
	}
	addrIp, err := msgCommon.ParseIPAddr(data.Addr)
	if err != nil {
		log.Warn(err)
		return
	}
	nodeAddr := addrIp + ":" +
		strconv.Itoa(int(version.P.SyncPort))
	if config.DefConfig.P2PNode.ReservedPeersOnly && len(config.DefConfig.P2PNode.ReservedCfg.ReservedPeers) > 0 {
		found := false
		for _, addr := range config.DefConfig.P2PNode.ReservedCfg.ReservedPeers {
			if strings.HasPrefix(data.Addr, addr) {
				log.Debug("[p2p]peer in reserved list", data.Addr)
				found = true
				break
			}
		}
		if !found {
			remotePeer.CloseSync()
			remotePeer.CloseCons()
			log.Debug("[p2p]peer not in reserved list,close", data.Addr)
			return
		}

	}

	if version.P.IsConsensus == true {
		if config.DefConfig.P2PNode.DualPortSupport == false {
			log.Warn("[p2p]consensus port not surpport", data.Addr)
			remotePeer.CloseCons()
			return
		}

		p := p2p.GetPeer(version.P.Nonce)

		if p == nil {
			log.Warn("[p2p]sync link is not exist", version.P.Nonce, data.Addr)
			remotePeer.CloseCons()
			remotePeer.CloseSync()
			return
		} else {
			//p synclink must exist,merged
			p.ConsLink = remotePeer.ConsLink
			p.ConsLink.SetID(version.P.Nonce)
			p.SetConsState(remotePeer.GetConsState())
			remotePeer = p

		}
		if version.P.Nonce == p2p.GetID() {
			log.Warn("[p2p]the node handshake with itself", data.Addr)
			p2p.SetOwnAddress(nodeAddr)
			p2p.RemoveFromInConnRecord(remotePeer.GetAddr())
			p2p.RemoveFromOutConnRecord(remotePeer.GetAddr())
			remotePeer.CloseCons()
			return
		}

		s := remotePeer.GetConsState()
		if s != msgCommon.INIT && s != msgCommon.HAND {
			log.Warnf("[p2p]unknown status to received version,%d,%s\n", s, data.Addr)
			remotePeer.CloseCons()
			return
		}

		// Todo: change the method of input parameters
		remotePeer.UpdateInfo(time.Now(), version.P.Version,
			version.P.Services, version.P.SyncPort,
			version.P.ConsPort, version.P.Nonce,
			version.P.Relay, version.P.StartHeight)

		var msg msgTypes.Message
		if s == msgCommon.INIT {
			remotePeer.SetConsState(msgCommon.HAND_SHAKE)
			msg = msgpack.NewVersion(p2p, true, ledger.DefLedger.GetCurrentBlockHeight())
		} else if s == msgCommon.HAND {
			remotePeer.SetConsState(msgCommon.HAND_SHAKED)
			msg = msgpack.NewVerAck(true)

		}
		err := p2p.Send(remotePeer, msg, true)
		if err != nil {
			log.Warn(err)
			return
		}
	} else {
		if version.P.Nonce == p2p.GetID() {
			p2p.RemoveFromInConnRecord(remotePeer.GetAddr())
			p2p.RemoveFromOutConnRecord(remotePeer.GetAddr())
			log.Warn("[p2p]the node handshake with itself", remotePeer.GetAddr())
			p2p.SetOwnAddress(nodeAddr)
			remotePeer.CloseSync()
			return
		}

		s := remotePeer.GetSyncState()
		if s != msgCommon.INIT && s != msgCommon.HAND {
			log.Warnf("[p2p]unknown status to received version,%d,%s\n", s, remotePeer.GetAddr())
			remotePeer.CloseSync()
			return
		}

		// Obsolete node
		p := p2p.GetPeer(version.P.Nonce)
		if p != nil {
			ipOld, err := msgCommon.ParseIPAddr(p.GetAddr())
			if err != nil {
				log.Warn("[p2p]exist peer %d ip format is wrong %s", version.P.Nonce, p.GetAddr())
				return
			}
			ipNew, err := msgCommon.ParseIPAddr(data.Addr)
			if err != nil {
				remotePeer.CloseSync()
				log.Warn("[p2p]connecting peer %d ip format is wrong %s, close", version.P.Nonce, data.Addr)
				return
			}
			if ipNew == ipOld {
				//same id and same ip
				n, ret := p2p.DelNbrNode(version.P.Nonce)
				if ret == true {
					log.Infof("[p2p]peer reconnect %d", version.P.Nonce, data.Addr)
					// Close the connection and release the node source
					n.CloseSync()
					n.CloseCons()
					if pid != nil {
						input := &msgCommon.RemovePeerID{
							ID: version.P.Nonce,
						}
						pid.Tell(input)
					}
				}
			} else {
				log.Warnf("[p2p]same peer id from different addr: %s, %s close latest one", ipOld, ipNew)
				remotePeer.CloseSync()
				return

			}
		}

		if version.P.Cap[msgCommon.HTTP_INFO_FLAG] == 0x01 {
			remotePeer.SetHttpInfoState(true)
		} else {
			remotePeer.SetHttpInfoState(false)
		}
		remotePeer.SetHttpInfoPort(version.P.HttpInfoPort)

		remotePeer.UpdateInfo(time.Now(), version.P.Version,
			version.P.Services, version.P.SyncPort,
			version.P.ConsPort, version.P.Nonce,
			version.P.Relay, version.P.StartHeight)
		remotePeer.SyncLink.SetID(version.P.Nonce)
		p2p.AddNbrNode(remotePeer)

		if pid != nil {
			input := &msgCommon.AppendPeerID{
				ID: version.P.Nonce,
			}
			pid.Tell(input)
		}

		var msg msgTypes.Message
		if s == msgCommon.INIT {
			remotePeer.SetSyncState(msgCommon.HAND_SHAKE)
			msg = msgpack.NewVersion(p2p, false, ledger.DefLedger.GetCurrentBlockHeight())
		} else if s == msgCommon.HAND {
			remotePeer.SetSyncState(msgCommon.HAND_SHAKED)
			msg = msgpack.NewVerAck(false)
		}
		err := p2p.Send(remotePeer, msg, false)
		if err != nil {
			log.Warn(err)
			return
		}
	}
}

// VerAckHandle handles the version ack from peer
func VerAckHandle(data *msgTypes.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Trace("[p2p]receive verAck message from ", data.Addr, data.Id)

	verAck := data.Payload.(*msgTypes.VerACK)
	remotePeer := p2p.GetPeer(data.Id)

	if remotePeer == nil {
		log.Warn("[p2p]nbr node is not exist", data.Id, data.Addr)
		return
	}

	if verAck.IsConsensus == true {
		if config.DefConfig.P2PNode.DualPortSupport == false {
			log.Warn("[p2p]consensus port not surpport")
			return
		}
		s := remotePeer.GetConsState()
		if s != msgCommon.HAND_SHAKE && s != msgCommon.HAND_SHAKED {
			log.Warnf("[p2p]unknown status to received verAck,state:%d,%s\n", s, data.Addr)
			return
		}

		remotePeer.SetConsState(msgCommon.ESTABLISH)
		p2p.RemoveFromConnectingList(data.Addr)
		remotePeer.SetConsConn(remotePeer.GetConsConn())

		if s == msgCommon.HAND_SHAKE {
			msg := msgpack.NewVerAck(true)
			p2p.Send(remotePeer, msg, true)
		}
	} else {
		s := remotePeer.GetSyncState()
		if s != msgCommon.HAND_SHAKE && s != msgCommon.HAND_SHAKED {
			log.Warnf("[p2p]unknown status to received verAck,state:%d,%s\n", s, data.Addr)
			return
		}

		remotePeer.SetSyncState(msgCommon.ESTABLISH)
		p2p.RemoveFromConnectingList(data.Addr)
		remotePeer.DumpInfo()

		addr := remotePeer.SyncLink.GetAddr()

		if s == msgCommon.HAND_SHAKE {
			msg := msgpack.NewVerAck(false)
			p2p.Send(remotePeer, msg, false)
		} else {
			//consensus port connect
			if config.DefConfig.P2PNode.DualPortSupport && remotePeer.GetConsPort() > 0 {
				addrIp, err := msgCommon.ParseIPAddr(addr)
				if err != nil {
					log.Warn(err)
					return
				}
				nodeConsensusAddr := addrIp + ":" +
					strconv.Itoa(int(remotePeer.GetConsPort()))
				go p2p.Connect(nodeConsensusAddr, true)
			}
		}

		msg := msgpack.NewAddrReq()
		go p2p.Send(remotePeer, msg, false)
	}

}

// AddrHandle handles the neighbor address response message from peer
func AddrHandle(data *msgTypes.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Trace("[p2p]handle addr message", data.Addr, data.Id)

	var msg = data.Payload.(*msgTypes.Addr)
	for _, v := range msg.NodeAddrs {
		var ip net.IP
		ip = v.IpAddr[:]
		address := ip.To16().String() + ":" + strconv.Itoa(int(v.Port))

		if v.ID == p2p.GetID() {
			continue
		}

		if p2p.NodeEstablished(v.ID) {
			continue
		}

		if ret := p2p.GetPeerFromAddr(address); ret != nil {
			continue
		}

		if v.Port == 0 {
			continue
		}
		if p2p.IsAddrFromConnecting(address) {
			continue
		}
		log.Debug("[p2p]connect ip address:", address)
		go p2p.Connect(address, false)
	}
}

// DataReqHandle handles the data req(block/Transaction) from peer
func DataReqHandle(data *msgTypes.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Trace("[p2p]receive data req message", data.Addr, data.Id)

	var dataReq = data.Payload.(*msgTypes.DataReq)

	remotePeer := p2p.GetPeer(data.Id)
	if remotePeer == nil {
		log.Debug("[p2p]remotePeer invalid in DataReqHandle")
		return
	}
	reqType := common.InventoryType(dataReq.DataType)
	hash := dataReq.Hash
	switch reqType {
	case common.BLOCK:
		reqID := fmt.Sprintf("%x%s", reqType, hash.ToHexString())
		data := getRespCacheValue(reqID)
		var msg msgTypes.Message
		if data != nil {
			switch data.(type) {
			case *msgTypes.Block:
				msg = data.(*msgTypes.Block)
			}
		}
		if msg == nil {
			var merkleRoot common.Uint256
			block, err := ledger.DefLedger.GetBlockByHash(hash)
			if err != nil || block == nil || block.Header == nil {
				log.Debug("[p2p]can't get block by hash: ", hash,
					" ,send not found message")
				msg := msgpack.NewNotFound(hash)
				err := p2p.Send(remotePeer, msg, false)
				if err != nil {
					log.Warn(err)
					return
				}
				return
			}
			merkleRoot, err = ledger.DefLedger.GetStateMerkleRoot(block.Header.Height)
			if err != nil {
				log.Debugf("[p2p]failed to get state merkel root at height %v, err %v",
					block.Header.Height, err)
				msg := msgpack.NewNotFound(hash)
				err := p2p.Send(remotePeer, msg, false)
				if err != nil {
					log.Warn(err)
					return
				}
				return
			}
			msg = msgpack.NewBlock(block, merkleRoot)
			saveRespCache(reqID, msg)
		}
		err := p2p.Send(remotePeer, msg, false)
		if err != nil {
			log.Warn(err)
			return
		}

	case common.TRANSACTION:
		txn, err := ledger.DefLedger.GetTransaction(hash)
		if err != nil {
			log.Debug("[p2p]Can't get transaction by hash: ",
				hash, " ,send not found message")
			msg := msgpack.NewNotFound(hash)
			err = p2p.Send(remotePeer, msg, false)
			if err != nil {
				log.Warn(err)
				return
			}
		}
		msg := msgpack.NewTxn(txn)
		err = p2p.Send(remotePeer, msg, false)
		if err != nil {
			log.Warn(err)
			return
		}
	}
}

// InvHandle handles the inventory message(block,
// transaction and consensus) from peer.
func InvHandle(data *msgTypes.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Trace("[p2p]receive inv message", data.Addr, data.Id)
	var inv = data.Payload.(*msgTypes.Inv)

	remotePeer := p2p.GetPeer(data.Id)
	if remotePeer == nil {
		log.Debug("[p2p]remotePeer invalid in InvHandle")
		return
	}
	if len(inv.P.Blk) == 0 {
		log.Debug("[p2p]empty inv payload in InvHandle")
		return
	}
	var id common.Uint256
	str := inv.P.Blk[0].ToHexString()
	log.Debugf("[p2p]the inv type: 0x%x block len: %d, %s\n",
		inv.P.InvType, len(inv.P.Blk), str)

	invType := common.InventoryType(inv.P.InvType)
	switch invType {
	case common.TRANSACTION:
		log.Debug("[p2p]receive transaction message", id)
		// TODO check the ID queue
		id = inv.P.Blk[0]
		trn, err := ledger.DefLedger.GetTransaction(id)
		if trn == nil || err != nil {
			msg := msgpack.NewTxnDataReq(id)
			err = p2p.Send(remotePeer, msg, false)
			if err != nil {
				log.Warn(err)
				return
			}
		}
	case common.BLOCK:
		log.Debug("[p2p]receive block message")
		for _, id = range inv.P.Blk {
			log.Debug("[p2p]receive inv-block message, hash is ", id)
			// TODO check the ID queue
			isContainBlock, err := ledger.DefLedger.IsContainBlock(id)
			if err != nil {
				log.Warn(err)
				return
			}
			if !isContainBlock && msgTypes.LastInvHash != id {
				msgTypes.LastInvHash = id
				// send the block request
				log.Infof("[p2p]inv request block hash: %x", id)
				msg := msgpack.NewBlkDataReq(id)
				err = p2p.Send(remotePeer, msg, false)
				if err != nil {
					log.Warn(err)
					return
				}
			}
		}
	case common.CONSENSUS:
		log.Debug("[p2p]receive consensus message")
		id = inv.P.Blk[0]
		msg := msgpack.NewConsensusDataReq(id)
		err := p2p.Send(remotePeer, msg, true)
		if err != nil {
			log.Warn(err)
			return
		}
	default:
		log.Warn("[p2p]receive unknown inventory message")
	}

}

// DisconnectHandle handles the disconnect events
func DisconnectHandle(data *msgTypes.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Debug("[p2p]receive disconnect message", data.Addr, data.Id)
	p2p.RemoveFromInConnRecord(data.Addr)
	p2p.RemoveFromOutConnRecord(data.Addr)
	remotePeer := p2p.GetPeer(data.Id)
	if remotePeer == nil {
		log.Debug("[p2p]disconnect peer is nil")
		return
	}
	p2p.RemoveFromConnectingList(data.Addr)

	if remotePeer.SyncLink.GetAddr() == data.Addr {
		p2p.RemovePeerSyncAddress(data.Addr)
		p2p.RemovePeerConsAddress(data.Addr)
		remotePeer.CloseSync()
		remotePeer.CloseCons()
	}
	if remotePeer.ConsLink.GetAddr() == data.Addr {
		p2p.RemovePeerConsAddress(data.Addr)
		remotePeer.CloseCons()
	}
}

//get blk hdrs from starthash to stophash
func GetHeadersFromHash(startHash common.Uint256, stopHash common.Uint256) ([]*types.Header, error) {
	var count uint32 = 0
	headers := []*types.Header{}
	var startHeight uint32
	var stopHeight uint32
	curHeight := ledger.DefLedger.GetCurrentHeaderHeight()
	if startHash == common.UINT256_EMPTY {
		if stopHash == common.UINT256_EMPTY {
			if curHeight > msgCommon.MAX_BLK_HDR_CNT {
				count = msgCommon.MAX_BLK_HDR_CNT
			} else {
				count = curHeight
			}
		} else {
			bkStop, err := ledger.DefLedger.GetHeaderByHash(stopHash)
			if err != nil || bkStop == nil {
				return nil, err
			}
			stopHeight = bkStop.Height
			count = curHeight - stopHeight
			if count > msgCommon.MAX_BLK_HDR_CNT {
				count = msgCommon.MAX_BLK_HDR_CNT
			}
		}
	} else {
		bkStart, err := ledger.DefLedger.GetHeaderByHash(startHash)
		if err != nil || bkStart == nil {
			return nil, err
		}
		startHeight = bkStart.Height
		if stopHash != common.UINT256_EMPTY {
			bkStop, err := ledger.DefLedger.GetHeaderByHash(stopHash)
			if err != nil || bkStop == nil {
				return nil, err
			}
			stopHeight = bkStop.Height

			// avoid unsigned integer underflow
			if startHeight < stopHeight {
				return nil, errors.New("[p2p]do not have header to send")
			}
			count = startHeight - stopHeight

			if count >= msgCommon.MAX_BLK_HDR_CNT {
				count = msgCommon.MAX_BLK_HDR_CNT
				stopHeight = startHeight - msgCommon.MAX_BLK_HDR_CNT
			}
		} else {

			if startHeight > msgCommon.MAX_BLK_HDR_CNT {
				count = msgCommon.MAX_BLK_HDR_CNT
			} else {
				count = startHeight
			}
		}
	}

	var i uint32
	for i = 1; i <= count; i++ {
		hash := ledger.DefLedger.GetBlockHash(stopHeight + i)
		hd, err := ledger.DefLedger.GetHeaderByHash(hash)
		if err != nil {
			log.Debugf("[p2p]net_server GetBlockWithHeight failed with err=%s, hash=%x,height=%d\n", err.Error(), hash, stopHeight+i)
			return nil, err
		}
		headers = append(headers, hd)
	}

	return headers, nil
}

//getRespCacheValue get response data from cache
func getRespCacheValue(key string) interface{} {
	if respCache == nil {
		return nil
	}
	data, ok := respCache.Get(key)
	if ok {
		return data
	}
	return nil
}

//saveRespCache save response msg to cache
func saveRespCache(key string, value interface{}) bool {
	if respCache == nil {
		var err error
		respCache, err = lru.NewARC(msgCommon.MAX_RESP_CACHE_SIZE)
		if err != nil {
			return false
		}
	}
	respCache.Add(key, value)
	return true
}
