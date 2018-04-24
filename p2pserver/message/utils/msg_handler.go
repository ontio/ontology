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
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	evtActor "github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/types"
	actor "github.com/ontio/ontology/p2pserver/actor/req"
	msgCommon "github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/message/msg_pack"
	msgTypes "github.com/ontio/ontology/p2pserver/message/types"
	"github.com/ontio/ontology/p2pserver/net/protocol"
)

// AddrReqHandle hadnles the neighbor address request from peer
func AddrReqHandle(data *msgCommon.MsgPayload, args ...interface{}) error {
	log.Debug("RX addr request message")
	p2p := args[0].(p2p.P2P)
	remotePeer := p2p.GetPeer(data.Id)
	if remotePeer == nil {
		return errors.New("remotePeer invalid in AddrReqHandle")
	}

	var addrStr []msgCommon.PeerAddr
	var count uint64
	addrStr, count = p2p.GetNeighborAddrs()
	buf, err := msgpack.NewAddrs(addrStr, count)
	if err != nil {
		return err
	}
	p2p.Send(remotePeer, buf, false)
	return nil
}

// HeaderReqHandle handles the header sync req from peer
func HeadersReqHandle(data *msgCommon.MsgPayload, args ...interface{}) error {
	log.Debug("RX headers request message")
	p2p := args[0].(p2p.P2P)
	length := len(data.Payload)

	var headersReq msgTypes.HeadersReq
	headersReq.Deserialization(data.Payload[:length])
	headersReq.Verify(data.Payload[msgCommon.MSG_HDR_LEN:length])

	var startHash [msgCommon.HASH_LEN]byte
	var stopHash [msgCommon.HASH_LEN]byte
	startHash = headersReq.P.HashStart
	stopHash = headersReq.P.HashEnd

	headers, cnt, err := actor.GetHeadersFromHash(startHash, stopHash)
	if err != nil {
		return err
	}
	buf, err := msgpack.NewHeaders(headers, cnt)
	if err != nil {
		return err
	}
	remotePeer := p2p.GetPeer(data.Id)
	p2p.Send(remotePeer, buf, false)
	return nil
}

//PingHandle handle ping msg from peer
func PingHandle(data *msgCommon.MsgPayload, args ...interface{}) error {
	log.Debug("RX ping message")
	p2p := args[0].(p2p.P2P)
	length := len(data.Payload)

	var ping msgTypes.Ping
	ping.Deserialization(data.Payload[:length])
	ping.Verify(data.Payload[msgCommon.MSG_HDR_LEN:length])

	remotePeer := p2p.GetPeer(data.Id)

	remotePeer.SetHeight(ping.Height)

	height, err := actor.GetCurrentBlockHeight()
	if err != nil {
		return err
	}
	p2p.SetHeight(uint64(height))
	buf, err := msgpack.NewPongMsg(uint64(height))

	if err != nil {
		log.Error("failed build a new pong message")
	} else {
		p2p.Send(remotePeer, buf, false)
	}
	return err
}

///PongHandle handle pong msg from peer
func PongHandle(data *msgCommon.MsgPayload, args ...interface{}) error {
	log.Debug("RX pong message")
	p2p := args[0].(p2p.P2P)
	length := len(data.Payload)

	var pong msgTypes.Pong
	pong.Deserialization(data.Payload[:length])
	pong.Verify(data.Payload[msgCommon.MSG_HDR_LEN:length])

	remotePeer := p2p.GetPeer(data.Id)
	remotePeer.SetHeight(pong.Height)
	return nil
}

// BlkHeaderHandle handles the sync headers from peer
func BlkHeaderHandle(data *msgCommon.MsgPayload, args ...interface{}) error {
	log.Debug("RX block header message")
	length := len(data.Payload)
	var blkHeader msgTypes.BlkHeader
	blkHeader.Deserialization(data.Payload[:length])
	blkHeader.Verify(data.Payload[msgCommon.MSG_HDR_LEN:length])

	var blkHdr []*types.Header
	var i uint32
	for i = 0; i < blkHeader.Cnt; i++ {
		blkHdr = append(blkHdr, &blkHeader.BlkHdr[i])
	}
	actor.AddHeaders(blkHdr)
	return nil
}

// BlockHandle handles the block message from peer
func BlockHandle(data *msgCommon.MsgPayload, args ...interface{}) error {
	log.Debug("RX block message from ", data.Id)
	var pid *evtActor.PID
	if args[1] != nil {
		pid = args[1].(*evtActor.PID)
	}
	length := len(data.Payload)

	var block msgTypes.Block
	block.Deserialization(data.Payload[:length])
	block.Verify(data.Payload[msgCommon.MSG_HDR_LEN:length])

	hash := block.Blk.Hash()
	if con, _ := actor.IsContainBlock(hash); con != true {
		actor.AddBlock(&block.Blk)
		if pid != nil {
			input := &msgCommon.RemoveFlightHeight{
				Id:     data.Id,
				Height: block.Blk.Header.Height,
			}
			pid.Tell(input)
		}
	} else {
		log.Debug("Receive duplicated block")
	}
	return nil
}

// ConsensusHandle handles the consensus message from peer
func ConsensusHandle(data *msgCommon.MsgPayload, args ...interface{}) error {
	log.Debug("RX consensus message")
	length := len(data.Payload)

	var consensus msgTypes.Consensus
	consensus.Deserialization(data.Payload[:length])
	consensus.Cons.Verify()

	if actor.ConsensusPid != nil {
		actor.ConsensusPid.Tell(&consensus.Cons)
	}
	return nil
}

// NotFoundHandle handles the not found message from peer
func NotFoundHandle(data *msgCommon.MsgPayload, args ...interface{}) error {
	length := len(data.Payload)

	var notFound msgTypes.NotFound
	notFound.Deserialization(data.Payload[:length])
	log.Debug("RX notFound message, hash is ", notFound.Hash)
	return nil
}

// TransactionHandle handles the transaction message from peer
func TransactionHandle(data *msgCommon.MsgPayload, args ...interface{}) error {
	log.Debug("RX transaction message")
	length := len(data.Payload)

	var trn msgTypes.Trn
	trn.Deserialization(data.Payload[:length])

	tx := &trn.Txn
	actor.AddTransaction(tx)
	log.Debug("RX Transaction message hash", tx.Hash())

	return nil
}

// VersionHandle handles version handshake protocol from peer
func VersionHandle(data *msgCommon.MsgPayload, args ...interface{}) error {
	log.Debug("RX version message")
	p2p := args[0].(p2p.P2P)
	length := len(data.Payload)

	if length == 0 {
		log.Error(fmt.Sprintf("nil message for %s", msgCommon.VERSION_TYPE))
		return errors.New("nil message")
	}

	version := msgTypes.Version{}
	version.Deserialization(data.Payload[:length])
	version.Verify(data.Payload[msgCommon.MSG_HDR_LEN:length])

	if version.P.IsConsensus == true {
		if config.Parameters.DualPortSurpport == false {
			log.Warn("consensus port not surpport")
			return errors.New("consensus port not surpport")
		}
		remotePeer := p2p.GetPeerFromAddr(data.Addr)

		if remotePeer == nil {
			log.Warn(" peer is not exist")
			return errors.New("peer is not exist ")
		}
		p := p2p.GetPeer(version.P.Nonce)

		if p == nil {
			log.Warn("sync link is not exist")
			p2p.RemovePeerConsAddress(data.Addr)
			return errors.New("sync link is not exist")
		} else {
			//p synclink must exist,merged
			p.ConsLink = remotePeer.ConsLink
			p.ConsLink.SetID(version.P.Nonce)
			p.SetConsState(remotePeer.GetConsState())
			remotePeer = p

		}
		if version.P.Nonce == p2p.GetID() {
			log.Warn("The node handshake with itself")
			remotePeer.CloseCons()
			p2p.RemovePeerConsAddress(data.Addr)
			return errors.New("The node handshake with itself ")
		}

		s := remotePeer.GetConsState()
		if s != msgCommon.INIT && s != msgCommon.HAND {
			log.Warn("Unknown status to received version", s)
			return errors.New("Unknown status to received version ")
		}

		// Todo: change the method of input parameters
		remotePeer.UpdateInfo(time.Now(), version.P.Version,
			version.P.Services, version.P.SyncPort,
			version.P.ConsPort, version.P.Nonce,
			version.P.Relay, version.P.StartHeight)

		var buf []byte
		if s == msgCommon.INIT {
			remotePeer.SetConsState(msgCommon.HAND_SHAKE)
			vpl := msgpack.NewVersionPayload(p2p, true)
			buf, _ = msgpack.NewVersion(vpl, p2p.GetPubKey())
		} else if s == msgCommon.HAND {
			remotePeer.SetConsState(msgCommon.HAND_SHAKED)
			buf, _ = msgpack.NewVerAck(true)
		}
		p2p.Send(remotePeer, buf, true)
		return nil
	} else {
		remotePeer := p2p.GetPeerFromAddr(data.Addr)

		if remotePeer == nil {
			log.Warn("peer is not exist")
			p2p.RemovePeerSyncAddress(data.Addr)
			return errors.New("peer is not exist ")
		}
		if version.P.Nonce == p2p.GetID() {
			log.Warn("The node handshake with itself")
			remotePeer.CloseSync()
			p2p.RemovePeerSyncAddress(data.Addr)
			return errors.New("The node handshake with itself ")
		}

		s := remotePeer.GetSyncState()
		if s != msgCommon.INIT && s != msgCommon.HAND {
			log.Warn("Unknown status to received version", s)
			return errors.New("Unknown status to received version")
		}

		// Obsolete node
		n, ret := p2p.DelNbrNode(version.P.Nonce)
		if ret == true {
			log.Info(fmt.Sprintf("Peer reconnect 0x%x", version.P.Nonce))
			// Close the connection and release the node source
			n.SetSyncState(msgCommon.INACTIVITY)
			n.CloseSync()
			n.SetConsState(msgCommon.INACTIVITY)
			n.CloseCons()
			p2p.RemovePeerSyncAddress(data.Addr)
			p2p.RemovePeerConsAddress(data.Addr)
		}

		log.Debug("handle version version.pk is ", version.PK)
		if version.P.Cap[msgCommon.HTTP_INFO_FLAG] == 0x01 {
			remotePeer.SetHttpInfoState(true)
		} else {
			remotePeer.SetHttpInfoState(false)
		}
		remotePeer.SetHttpInfoPort(version.P.HttpInfoPort)
		remotePeer.SetBookKeeperAddr(version.PK)

		// if  version.P.Port == version.P.ConsensusPort don't updateInfo
		remotePeer.UpdateInfo(time.Now(), version.P.Version,
			version.P.Services, version.P.SyncPort,
			version.P.ConsPort, version.P.Nonce,
			version.P.Relay, version.P.StartHeight)
		remotePeer.SyncLink.SetID(version.P.Nonce)
		p2p.AddNbrNode(remotePeer)

		var buf []byte
		if s == msgCommon.INIT {
			remotePeer.SetSyncState(msgCommon.HAND_SHAKE)
			vpl := msgpack.NewVersionPayload(p2p, false)
			buf, _ = msgpack.NewVersion(vpl, p2p.GetPubKey())
		} else if s == msgCommon.HAND {
			remotePeer.SetSyncState(msgCommon.HAND_SHAKED)
			buf, _ = msgpack.NewVerAck(false)
		}
		p2p.Send(remotePeer, buf, false)
		return nil
	}
	return nil
}

// VerAckHandle handles the version ack from peer
func VerAckHandle(data *msgCommon.MsgPayload, args ...interface{}) error {
	log.Debug("RX verAck message")
	p2p := args[0].(p2p.P2P)
	length := len(data.Payload)
	p2p.RemoveFromConnectingList(data.Addr)
	if length == 0 {
		log.Error(fmt.Sprintf("nil message for %s", msgCommon.VERACK_TYPE))
		return errors.New("nil message")
	}

	verAck := msgTypes.VerACK{}
	verAck.Deserialization(data.Payload[:length])

	remotePeer := p2p.GetPeer(data.Id)

	if remotePeer == nil {
		log.Warn("nbr node is not exist")
		return errors.New("nbr node is not exist ")
	}

	if verAck.IsConsensus == true {
		if config.Parameters.DualPortSurpport == false {
			log.Warn("consensus port not surpport")
			return errors.New("consensus port not surpport")
		}
		s := remotePeer.GetConsState()
		if s != msgCommon.HAND_SHAKE && s != msgCommon.HAND_SHAKED {
			log.Warn("Unknown status to received verAck", s)
			return errors.New("Unknown status to received verAck")
		}

		remotePeer.SetConsState(msgCommon.ESTABLISH)
		remotePeer.SetConsConn(remotePeer.GetConsConn())

		if s == msgCommon.HAND_SHAKE {
			buf, _ := msgpack.NewVerAck(true)
			p2p.Send(remotePeer, buf, true)
		}
		addr := remotePeer.ConsLink.GetAddr()
		p2p.RemoveFromConnectingList(addr)
		return nil
	} else {
		s := remotePeer.GetSyncState()
		if s != msgCommon.HAND_SHAKE && s != msgCommon.HAND_SHAKED {
			log.Warn("Unknown status to received verAck", s)
			return errors.New("Unknown status to received verAck ")
		}

		remotePeer.SetSyncState(msgCommon.ESTABLISH)

		if s == msgCommon.HAND_SHAKE {
			buf, _ := msgpack.NewVerAck(false)
			p2p.Send(remotePeer, buf, false)
		}

		remotePeer.DumpInfo()

		buf, _ := msgpack.NewAddrReq()
		go p2p.Send(remotePeer, buf, false)

		addr := remotePeer.SyncLink.GetAddr()
		p2p.RemoveFromConnectingList(addr)
		//consensus port connect
		if config.Parameters.DualPortSurpport == false {
			return nil
		}

		i := strings.Index(addr, ":")
		if i < 0 {
			log.Warn("Split IP address error", addr)
			return nil
		}
		nodeConsensusAddr := addr[:i] + ":" +
			strconv.Itoa(int(remotePeer.GetConsPort()))
		go p2p.Connect(nodeConsensusAddr, true)

		return nil
	}
	return nil
}

// AddrHandle handles the neighbor address response message from peer
func AddrHandle(data *msgCommon.MsgPayload, args ...interface{}) error {
	log.Debug("Handle addr message")
	p2p := args[0].(p2p.P2P)
	length := len(data.Payload)

	var msg msgTypes.Addr
	msg.Deserialization(data.Payload[:length])
	msg.Verify(data.Payload[msgCommon.MSG_HDR_LEN:length])

	for _, v := range msg.NodeAddrs {
		var ip net.IP
		ip = v.IpAddr[:]
		address := ip.To16().String() + ":" + strconv.Itoa(int(v.Port))
		log.Info(fmt.Sprintf("The ip address is %s id is 0x%x", address, v.ID))

		if v.ID == p2p.GetID() {
			continue
		}

		if p2p.NodeEstablished(v.ID) {
			continue
		}

		if v.Port == 0 {
			continue
		}
		log.Info("Connect ipaddr ：", address)
		go p2p.Connect(address, false)
	}
	return nil
}

// DataReqHandle handles the data req(block/Transaction) from peer
func DataReqHandle(data *msgCommon.MsgPayload, args ...interface{}) error {
	log.Debug("RX data req message")
	p2p := args[0].(p2p.P2P)
	length := len(data.Payload)

	var dataReq msgTypes.DataReq
	dataReq.Deserialization(data.Payload[:length])

	//localPeer := p2p.Self
	remotePeer := p2p.GetPeer(data.Id)

	reqType := common.InventoryType(dataReq.DataType)
	hash := dataReq.Hash
	switch reqType {
	case common.BLOCK:
		block, err := actor.GetBlockByHash(hash)
		if err != nil {
			log.Debug("Can't get block by hash: ", hash,
				" ,send not found message")
			b, err := msgpack.NewNotFound(hash)
			p2p.Send(remotePeer, b, false)
			return err
		}
		log.Debug("block height is ", block.Header.Height,
			" ,hash is ", hash)
		buf, err := msgpack.NewBlock(block)
		if err != nil {
			return err
		}
		p2p.Send(remotePeer, buf, false)

	case common.TRANSACTION:
		txn, err := actor.GetTxnFromLedger(hash)
		if err != nil {
			log.Debug("Can't get transaction by hash: ",
				hash, " ,send not found message")
			b, err := msgpack.NewNotFound(hash)
			p2p.Send(remotePeer, b, false)
			return err
		}
		buf, err := msgpack.NewTxn(txn)
		if err != nil {
			return err
		}
		p2p.Send(remotePeer, buf, false)
	}
	return nil
}

// InvHandle handles the inventory message(block,
// transaction and consensus) from peer.
func InvHandle(data *msgCommon.MsgPayload, args ...interface{}) error {
	log.Debug("RX inv message")
	p2p := args[0].(p2p.P2P)
	length := len(data.Payload)
	var inv msgTypes.Inv
	inv.Deserialization(data.Payload[:length])
	inv.Verify(data.Payload[msgCommon.MSG_HDR_LEN:length])

	//localPeer := p2p.Self
	remotePeer := p2p.GetPeer(data.Id)

	var id common.Uint256
	str := hex.EncodeToString(inv.P.Blk)
	log.Debug(fmt.Sprintf("The inv type: 0x%x block len: %d, %s\n",
		inv.P.InvType, len(inv.P.Blk), str))

	invType := common.InventoryType(inv.P.InvType)
	switch invType {
	case common.TRANSACTION:
		log.Debug("RX TRX message")
		// TODO check the ID queue
		id.Deserialize(bytes.NewReader(inv.P.Blk[:32]))

		trn, err := actor.GetTransaction(id)
		if trn == nil || err != nil {
			txnDataReq, _ := msgpack.NewTxnDataReq(id)
			p2p.Send(remotePeer, txnDataReq, false)
		}
	case common.BLOCK:
		log.Debug("RX block message")
		var i uint32
		count := inv.P.Cnt
		log.Debug("RX inv-block message, hash is ", inv.P.Blk)
		for i = 0; i < count; i++ {
			id.Deserialize(bytes.NewReader(inv.P.Blk[msgCommon.HASH_LEN*i:]))
			// TODO check the ID queue
			isContainBlock, _ := actor.IsContainBlock(id)
			if !isContainBlock && msgTypes.LastInvHash != id {
				msgTypes.LastInvHash = id
				// send the block request
				log.Infof("inv request block hash: %x", id)
				blkDataReq, _ := msgpack.NewBlkDataReq(id)
				p2p.Send(remotePeer, blkDataReq, false)
			}
		}
	case common.CONSENSUS:
		log.Debug("RX consensus message")
		id.Deserialize(bytes.NewReader(inv.P.Blk[:32]))
		consDataReq, _ := msgpack.NewConsensusDataReq(id)
		p2p.Send(remotePeer, consDataReq, true)
	default:
		log.Warn("RX unknown inventory message")
	}
	return nil
}

// DisconnectHandle handles the disconnect events
func DisconnectHandle(data *msgCommon.MsgPayload, args ...interface{}) error {
	p2p := args[0].(p2p.P2P)
	remotePeer := p2p.GetPeer(data.Id)
	if remotePeer == nil {
		return nil
	}
	i := strings.Index(data.Addr, ":")
	if i < 0 {
		log.Error("link address format error", data.Addr)
		return errors.New("link address format error")
	}
	port, err := strconv.Atoi(data.Addr[i+1:])
	if err != nil {
		log.Error("Split port error", data.Addr[i+1:])
		return errors.New("Split port error")
	}
	p2p.RemoveFromConnectingList(data.Addr)
	p2p.RemovePeerSyncAddress(data.Addr)
	p2p.RemovePeerConsAddress(data.Addr)

	if remotePeer.SyncLink.GetPort() == uint16(port) {
		remotePeer.CloseSync()
		remotePeer.CloseCons()
	}
	if remotePeer.ConsLink.GetPort() == uint16(port) {
		remotePeer.CloseCons()
	}
	return nil
}
