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
	"bytes"
	"encoding/hex"
	"errors"
	"net"
	"strconv"
	"strings"
	"time"

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

// AddrReqHandle hadnles the neighbor address request from peer
func AddrReqHandle(data *msgCommon.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Debug("receive addr request message", data.Addr, data.Id)
	remotePeer := p2p.GetPeer(data.Id)
	if remotePeer == nil {
		log.Error("remotePeer invalid in AddrReqHandle")
		return
	}

	var addrStr []msgCommon.PeerAddr
	addrStr = p2p.GetNeighborAddrs()
	buf, err := msgpack.NewAddrs(addrStr)
	if err != nil {
		log.Error(err)
		return
	}
	p2p.Send(remotePeer, buf, false)
}

// HeaderReqHandle handles the header sync req from peer
func HeadersReqHandle(data *msgCommon.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Debug("receive headers request message", data.Addr, data.Id)
	length := len(data.Payload)

	var headersReq msgTypes.HeadersReq
	headersReq.Deserialization(data.Payload[:length])
	if err := headersReq.Verify(data.Payload[msgCommon.MSG_HDR_LEN:length]); err != nil {
		log.Error(err)
		return
	}

	var startHash [msgCommon.HASH_LEN]byte
	var stopHash [msgCommon.HASH_LEN]byte
	startHash = headersReq.P.HashStart
	stopHash = headersReq.P.HashEnd

	headers, err := GetHeadersFromHash(startHash, stopHash)
	if err != nil {
		log.Error(err)
		return
	}
	buf, err := msgpack.NewHeaders(headers)
	if err != nil {
		log.Error(err)
		return
	}
	remotePeer := p2p.GetPeer(data.Id)
	if remotePeer == nil {
		log.Error("remotePeer invalid in HeadersReqHandle()")
		return
	}
	p2p.Send(remotePeer, buf, false)
}

//PingHandle handle ping msg from peer
func PingHandle(data *msgCommon.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Debug("receive ping message", data.Addr, data.Id)
	length := len(data.Payload)

	var ping msgTypes.Ping
	err := ping.Deserialization(data.Payload[:length])
	if err != nil {
		log.Error(err)
		return
	}
	if err = ping.Verify(data.Payload[msgCommon.MSG_HDR_LEN:length]); err != nil {
		log.Error(err)
		return
	}

	remotePeer := p2p.GetPeer(data.Id)
	if remotePeer == nil {
		log.Error("remotePeer invalid in PingHandle")
		return
	}
	remotePeer.SetHeight(ping.Height)

	height := ledger.DefLedger.GetCurrentBlockHeight()
	p2p.SetHeight(uint64(height))
	buf, err := msgpack.NewPongMsg(uint64(height))

	if err != nil {
		log.Error(err)
	} else {
		p2p.Send(remotePeer, buf, false)
	}
}

///PongHandle handle pong msg from peer
func PongHandle(data *msgCommon.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Debug("receive pong message", data.Addr, data.Id)
	length := len(data.Payload)

	var pong msgTypes.Pong
	err := pong.Deserialization(data.Payload[:length])
	if err != nil {
		log.Error(err)
		return
	}
	if err = pong.Verify(data.Payload[msgCommon.MSG_HDR_LEN:length]); err != nil {
		log.Error(err)
		return
	}

	remotePeer := p2p.GetPeer(data.Id)
	if remotePeer == nil {
		log.Error("remotePeer invalid in PongHandle")
		return
	}
	remotePeer.SetHeight(pong.Height)

}

// BlkHeaderHandle handles the sync headers from peer
func BlkHeaderHandle(data *msgCommon.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Debug("receive block header message", data.Addr, data.Id)
	length := len(data.Payload)
	var blkHeader msgTypes.BlkHeader
	err := blkHeader.Deserialization(data.Payload[:length])
	if err != nil {
		log.Error(err)
		return
	}
	if err = blkHeader.Verify(data.Payload[msgCommon.MSG_HDR_LEN:length]); err != nil {
		log.Error(err)
		return
	}

	var blkHdr []*types.Header
	var i uint32
	for i = 0; i < blkHeader.Cnt; i++ {
		blkHdr = append(blkHdr, &blkHeader.BlkHdr[i])
	}
	if pid != nil {
		input := &msgCommon.AppendHeaders{
			Headers: blkHdr,
		}
		pid.Tell(input)
	}
}

// BlockHandle handles the block message from peer
func BlockHandle(data *msgCommon.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Debug("receive block message from ", data.Addr, data.Id)
	length := len(data.Payload)

	var block msgTypes.Block
	err := block.Deserialization(data.Payload[:length])
	if err != nil {
		log.Error(err)
		return
	}
	if err = block.Verify(data.Payload[msgCommon.MSG_HDR_LEN:length]); err != nil {
		log.Error(err)
		return
	}

	if pid != nil {
		input := &msgCommon.AppendBlock{
			Block: &block.Blk,
		}
		pid.Tell(input)
	}
}

// ConsensusHandle handles the consensus message from peer
func ConsensusHandle(data *msgCommon.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Debug("receive consensus message", data.Addr, data.Id)
	length := len(data.Payload)

	var consensus msgTypes.Consensus
	err := consensus.Deserialization(data.Payload[:length])
	if err != nil {
		log.Error(err)
		return
	}
	if err = consensus.Cons.Verify(); err != nil {
		log.Error(err)
		return
	}

	if actor.ConsensusPid != nil {
		actor.ConsensusPid.Tell(&consensus.Cons)
	}
}

// NotFoundHandle handles the not found message from peer
func NotFoundHandle(data *msgCommon.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	length := len(data.Payload)

	var notFound msgTypes.NotFound
	err := notFound.Deserialization(data.Payload[:length])
	if err != nil {
		log.Error(err)
		return
	}
	log.Debug("receive notFound message, hash is ", notFound.Hash)
}

// TransactionHandle handles the transaction message from peer
func TransactionHandle(data *msgCommon.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Debug("receive transaction message", data.Addr, data.Id)
	length := len(data.Payload)

	var trn msgTypes.Trn
	err := trn.Deserialization(data.Payload[:length])
	if err != nil {
		log.Error(err)
		return
	}
	tx := &trn.Txn
	actor.AddTransaction(tx)
	log.Debug("receive Transaction message hash", tx.Hash())

}

// VersionHandle handles version handshake protocol from peer
func VersionHandle(data *msgCommon.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Debug("receive version message", data.Addr, data.Id)
	length := len(data.Payload)

	if length == 0 {
		log.Errorf("nil message for %s", msgCommon.VERSION_TYPE)
		return
	}

	version := msgTypes.Version{}
	err := version.Deserialization(data.Payload[:length])
	if err != nil {
		log.Error(err)
		return
	}
	if err = version.Verify(data.Payload[msgCommon.MSG_HDR_LEN:length]); err != nil {
		log.Error(err)
		return
	}

	remotePeer := p2p.GetPeerFromAddr(data.Addr)
	if remotePeer == nil {
		log.Warn("peer is not exist", data.Addr)
		//peer not exist,just remove list and return
		p2p.RemoveFromConnectingList(data.Addr)
		return
	}

	if version.P.IsConsensus == true {
		if config.DefConfig.P2PNode.DualPortSupport == false {
			log.Warn("consensus port not surpport")
			remotePeer.CloseCons()
			return
		}

		p := p2p.GetPeer(version.P.Nonce)

		if p == nil {
			log.Warn("sync link is not exist", version.P.Nonce)
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
			log.Warn("the node handshake with itself")
			remotePeer.CloseCons()
			return
		}

		s := remotePeer.GetConsState()
		if s != msgCommon.INIT && s != msgCommon.HAND {
			log.Warn("unknown status to received version", s)
			remotePeer.CloseCons()
			return
		}

		// Todo: change the method of input parameters
		remotePeer.UpdateInfo(time.Now(), version.P.Version,
			version.P.Services, version.P.SyncPort,
			version.P.ConsPort, version.P.Nonce,
			version.P.Relay, version.P.StartHeight)

		var buf []byte
		if s == msgCommon.INIT {
			remotePeer.SetConsState(msgCommon.HAND_SHAKE)
			vpl := msgpack.NewVersionPayload(p2p, true, ledger.DefLedger.GetCurrentBlockHeight())
			buf, err = msgpack.NewVersion(vpl, p2p.GetPubKey())
			if err != nil {
				log.Error(err)
				return
			}

		} else if s == msgCommon.HAND {
			remotePeer.SetConsState(msgCommon.HAND_SHAKED)
			buf, err = msgpack.NewVerAck(true)
			if err != nil {
				log.Error(err)
				return
			}

		}
		p2p.Send(remotePeer, buf, true)

	} else {

		if version.P.Nonce == p2p.GetID() {
			log.Warn("the node handshake with itself")
			remotePeer.CloseSync()
			return
		}

		s := remotePeer.GetSyncState()
		if s != msgCommon.INIT && s != msgCommon.HAND {
			log.Warn("unknown status to received version", s)
			remotePeer.CloseSync()
			return
		}

		// Obsolete node
		n, ret := p2p.DelNbrNode(version.P.Nonce)
		if ret == true {
			log.Infof("peer reconnect 0x%x", version.P.Nonce)
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

		log.Debug("handle version version.pk is ", version.PK)
		if version.P.Cap[msgCommon.HTTP_INFO_FLAG] == 0x01 {
			remotePeer.SetHttpInfoState(true)
		} else {
			remotePeer.SetHttpInfoState(false)
		}
		remotePeer.SetHttpInfoPort(version.P.HttpInfoPort)
		remotePeer.SetBookKeeperAddr(version.PK)

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

		var buf []byte
		if s == msgCommon.INIT {
			remotePeer.SetSyncState(msgCommon.HAND_SHAKE)
			vpl := msgpack.NewVersionPayload(p2p, false, ledger.DefLedger.GetCurrentBlockHeight())
			buf, err = msgpack.NewVersion(vpl, p2p.GetPubKey())
			if err != nil {
				log.Error(err)
				return
			}
		} else if s == msgCommon.HAND {
			remotePeer.SetSyncState(msgCommon.HAND_SHAKED)
			buf, err = msgpack.NewVerAck(false)
			if err != nil {
				log.Error(err)
				return
			}
		}
		p2p.Send(remotePeer, buf, false)
	}
}

// VerAckHandle handles the version ack from peer
func VerAckHandle(data *msgCommon.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Debug("receive verAck message from ", data.Addr, data.Id)
	length := len(data.Payload)
	p2p.RemoveFromConnectingList(data.Addr)
	if length == 0 {
		log.Errorf("nil message for %s", msgCommon.VERACK_TYPE)
		return
	}

	verAck := msgTypes.VerACK{}
	err := verAck.Deserialization(data.Payload[:length])
	if err != nil {
		log.Error(err)
		return
	}
	remotePeer := p2p.GetPeer(data.Id)

	if remotePeer == nil {
		log.Warn("nbr node is not exist", data.Id, data.Addr)
		return
	}

	if verAck.IsConsensus == true {
		if config.DefConfig.P2PNode.DualPortSupport == false {
			log.Warn("consensus port not surpport")
			return
		}
		s := remotePeer.GetConsState()
		if s != msgCommon.HAND_SHAKE && s != msgCommon.HAND_SHAKED {
			log.Warn("unknown status to received verAck", s)
			return
		}

		remotePeer.SetConsState(msgCommon.ESTABLISH)
		remotePeer.SetConsConn(remotePeer.GetConsConn())

		if s == msgCommon.HAND_SHAKE {
			buf, _ := msgpack.NewVerAck(true)
			p2p.Send(remotePeer, buf, true)
		}
	} else {
		s := remotePeer.GetSyncState()
		if s != msgCommon.HAND_SHAKE && s != msgCommon.HAND_SHAKED {
			log.Warn("unknown status to received verAck", s)
			return
		}

		remotePeer.SetSyncState(msgCommon.ESTABLISH)

		remotePeer.DumpInfo()

		addr := remotePeer.SyncLink.GetAddr()

		if s == msgCommon.HAND_SHAKE {
			buf, _ := msgpack.NewVerAck(false)
			p2p.Send(remotePeer, buf, false)

		} else {
			//consensus port connect
			if config.DefConfig.P2PNode.DualPortSupport && remotePeer.GetConsPort() > 0 {
				i := strings.Index(addr, ":")
				if i < 0 {
					log.Warn("split IP address error", addr)
					return
				}
				nodeConsensusAddr := addr[:i] + ":" +
					strconv.Itoa(int(remotePeer.GetConsPort()))
				go p2p.Connect(nodeConsensusAddr, true)
			}
		}

		buf, err := msgpack.NewAddrReq()
		if err != nil {
			log.Error(err)
			return
		}
		go p2p.Send(remotePeer, buf, false)
	}

}

// AddrHandle handles the neighbor address response message from peer
func AddrHandle(data *msgCommon.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Debug("handle addr message", data.Addr, data.Id)
	length := len(data.Payload)

	var msg msgTypes.Addr
	err := msg.Deserialization(data.Payload[:length])
	if err != nil {
		log.Error(err)
		return
	}
	if err = msg.Verify(data.Payload[msgCommon.MSG_HDR_LEN:length]); err != nil {
		log.Error(err)
		return
	}

	for _, v := range msg.NodeAddrs {
		var ip net.IP
		ip = v.IpAddr[:]
		address := ip.To16().String() + ":" + strconv.Itoa(int(v.Port))
		log.Infof("the ip address is %s id is 0x%x", address, v.ID)

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
		log.Info("connect ipaddr ：", address)
		go p2p.Connect(address, false)
	}
}

// DataReqHandle handles the data req(block/Transaction) from peer
func DataReqHandle(data *msgCommon.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Debug("receive data req message", data.Addr, data.Id)
	length := len(data.Payload)

	var dataReq msgTypes.DataReq
	err := dataReq.Deserialization(data.Payload[:length])
	if err != nil {
		log.Error(err)
		return
	}
	remotePeer := p2p.GetPeer(data.Id)
	if remotePeer == nil {
		log.Error("remotePeer invalid in DataReqHandle")
		return
	}
	reqType := common.InventoryType(dataReq.DataType)
	hash := dataReq.Hash
	switch reqType {
	case common.BLOCK:
		block, err := ledger.DefLedger.GetBlockByHash(hash)
		if err != nil || block == nil || block.Header == nil {
			log.Debug("can't get block by hash: ", hash,
				" ,send not found message")
			b, err := msgpack.NewNotFound(hash)
			if err != nil {
				log.Error(err)
				return
			}
			p2p.Send(remotePeer, b, false)
			return
		}
		log.Debug("block height is ", block.Header.Height,
			" ,hash is ", hash)
		buf, err := msgpack.NewBlock(block)
		if err != nil {
			log.Error(err)
			return
		}
		p2p.Send(remotePeer, buf, false)

	case common.TRANSACTION:
		txn, err := ledger.DefLedger.GetTransaction(hash)
		if err != nil {
			log.Debug("Can't get transaction by hash: ",
				hash, " ,send not found message")
			b, err := msgpack.NewNotFound(hash)
			if err != nil {
				log.Error(err)
				return
			}
			err = p2p.Send(remotePeer, b, false)
			if err != nil {
				log.Error(err)
				return
			}
		}
		buf, err := msgpack.NewTxn(txn)
		if err != nil {
			log.Error(err)
			return
		}
		err = p2p.Send(remotePeer, buf, false)
		if err != nil {
			log.Error(err)
			return
		}
	}
}

// InvHandle handles the inventory message(block,
// transaction and consensus) from peer.
func InvHandle(data *msgCommon.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Debug("receive inv message", data.Addr, data.Id)
	length := len(data.Payload)
	var inv msgTypes.Inv
	err := inv.Deserialization(data.Payload[:length])
	if err != nil {
		log.Error(err)
		return
	}
	if err = inv.Verify(data.Payload[msgCommon.MSG_HDR_LEN:length]); err != nil {
		log.Error(err)
		return
	}

	remotePeer := p2p.GetPeer(data.Id)
	if remotePeer == nil {
		log.Error("remotePeer invalid in InvHandle")
		return
	}
	var id common.Uint256
	str := hex.EncodeToString(inv.P.Blk)
	log.Debugf("the inv type: 0x%x block len: %d, %s\n",
		inv.P.InvType, len(inv.P.Blk), str)

	invType := common.InventoryType(inv.P.InvType)
	switch invType {
	case common.TRANSACTION:
		log.Debug("receive transaction message")
		// TODO check the ID queue
		err := id.Deserialize(bytes.NewReader(inv.P.Blk[:32]))
		if err != nil {
			log.Error(err)
			return
		}
		trn, err := ledger.DefLedger.GetTransaction(id)
		if trn == nil || err != nil {
			txnDataReq, err := msgpack.NewTxnDataReq(id)
			if err != nil {
				log.Error(err)
				return
			}
			err = p2p.Send(remotePeer, txnDataReq, false)
			if err != nil {
				log.Error(err)
				return
			}
		}
	case common.BLOCK:
		log.Debug("receive block message")
		var i uint32
		count := inv.P.Cnt
		log.Debug("receive inv-block message, hash is ", inv.P.Blk)
		for i = 0; i < count; i++ {
			id.Deserialize(bytes.NewReader(inv.P.Blk[msgCommon.HASH_LEN*i:]))
			// TODO check the ID queue
			isContainBlock, err := ledger.DefLedger.IsContainBlock(id)
			if err != nil {
				log.Error(err)
				return
			}
			if !isContainBlock && msgTypes.LastInvHash != id {
				msgTypes.LastInvHash = id
				// send the block request
				log.Infof("inv request block hash: %x", id)
				blkDataReq, err := msgpack.NewBlkDataReq(id)
				if err != nil {
					log.Error(err)
					return
				}
				err = p2p.Send(remotePeer, blkDataReq, false)
				if err != nil {
					log.Error(err)
					return
				}
			}
		}
	case common.CONSENSUS:
		log.Debug("receive consensus message")
		id.Deserialize(bytes.NewReader(inv.P.Blk[:32]))
		consDataReq, err := msgpack.NewConsensusDataReq(id)
		if err != nil {
			log.Error(err)
			return
		}
		err = p2p.Send(remotePeer, consDataReq, true)
		if err != nil {
			log.Error(err)
			return
		}
	default:
		log.Warn("receive unknown inventory message")
	}

}

// DisconnectHandle handles the disconnect events
func DisconnectHandle(data *msgCommon.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	remotePeer := p2p.GetPeer(data.Id)
	if remotePeer == nil {
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
func GetHeadersFromHash(startHash common.Uint256, stopHash common.Uint256) ([]types.Header, error) {
	var count uint32 = 0
	var empty [msgCommon.HASH_LEN]byte
	headers := []types.Header{}
	var startHeight uint32
	var stopHeight uint32
	curHeight := ledger.DefLedger.GetCurrentHeaderHeight()
	if startHash == empty {
		if stopHash == empty {
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
		if stopHash != empty {
			bkStop, err := ledger.DefLedger.GetHeaderByHash(stopHash)
			if err != nil || bkStop == nil {
				return nil, err
			}
			stopHeight = bkStop.Height

			// avoid unsigned integer underflow
			if startHeight < stopHeight {
				return nil, errors.New("do not have header to send")
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
			log.Errorf("net_server GetBlockWithHeight failed with err=%s, hash=%x,height=%d\n", err.Error(), hash, stopHeight+i)
			return nil, err
		}
		headers = append(headers, *hd)
	}

	return headers, nil
}
