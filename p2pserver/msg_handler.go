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
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/Ontology/common"
	"github.com/Ontology/common/log"
	"github.com/Ontology/core/types"
	actor "github.com/Ontology/p2pserver/actor/req"
	msgCommon "github.com/Ontology/p2pserver/common"
	msg "github.com/Ontology/p2pserver/message"
	"github.com/Ontology/p2pserver/msg_pack"
	"github.com/Ontology/p2pserver/peer"
)

func MsgHdrHandle(hdr msg.MsgHdr, peer peer.Peer, p2p P2PServer) error {
	log.Debug("RX MsgHdr message")
	return nil
}

// AddrReqHandle hadnles the neighbor address request from peer
func AddrReqHandle(data msgCommon.MsgPayload, p2p *P2PServer) error {
	log.Debug("RX addr request message")
	remotePeer := p2p.Self.Np.GetPeer(data.Id)

	var addrStr []msgCommon.PeerAddr
	var count uint64
	addrStr, count = p2p.GetNeighborAddrs()
	buf, err := msgpack.NewAddrs(addrStr, count)
	if err != nil {
		return err
	}
	remotePeer.SendToSync(buf)
	return nil
}

// HeaderReqHandle handles the header sync req from peer
func HeadersReqHandle(data msgCommon.MsgPayload, p2p *P2PServer) error {
	log.Debug("RX headers request message")

	length := len(data.Payload)

	var headersReq msg.HeadersReq
	headersReq.Deserialization(data.Payload[:length])
	headersReq.Verify(data.Payload[msgCommon.MSG_HDR_LEN:length])
	//Fix me:
	//node.LocalNode().AcqSyncReqSem()
	//defer node.LocalNode().RelSyncReqSem()

	var startHash [msgCommon.HASH_LEN]byte
	var stopHash [msgCommon.HASH_LEN]byte
	startHash = headersReq.P.HashStart
	stopHash = headersReq.P.HashEnd
	//FIXME if HeaderHashCount > 1
	headers, cnt, err := actor.GetHeadersFromHash(startHash, stopHash)
	if err != nil {
		return err
	}
	buf, err := msgpack.NewHeaders(headers, cnt)
	if err != nil {
		return err
	}
	remotePeer := p2p.Self.Np.GetPeer(data.Id)
	remotePeer.SendToSync(buf)
	return nil
}

// BlocksReqHandle handles the block sync req from peer
func BlocksReqHandle(data msgCommon.MsgPayload, p2p *P2PServer) error {
	log.Debug("RX blocks request message")

	length := len(data.Payload)

	var blocksReq msg.BlocksReq
	blocksReq.Deserialization(data.Payload[:length])
	blocksReq.Verify(data.Payload[msgCommon.MSG_HDR_LEN:length])

	var startHash common.Uint256
	var stopHash common.Uint256
	startHash = blocksReq.P.HashStart
	stopHash = blocksReq.P.HashStop

	//FIXME if HeaderHashCount > 1
	inv, err := actor.GetInvFromBlockHash(startHash, stopHash)
	if err != nil {
		return err
	}
	buf, err := msgpack.NewInv(inv)
	if err != nil {
		return err
	}
	remotePeer := p2p.Self.Np.GetPeer(data.Id)
	remotePeer.SendToSync(buf)
	return nil
}

func PingHandle(data msgCommon.MsgPayload, p2p *P2PServer) error {
	log.Debug("RX ping message")
	length := len(data.Payload)

	var ping msg.Ping
	ping.Deserialization(data.Payload[:length])
	ping.Verify(data.Payload[msgCommon.MSG_HDR_LEN:length])

	localPeer := p2p.Self
	remotePeer := p2p.Self.Np.GetPeer(data.Id)

	remotePeer.SetHeight(ping.Height)
	buf, err := msgpack.NewPongMsg(localPeer.GetHeight())

	if err != nil {
		log.Error("failed build a new pong message")
	} else {
		remotePeer.SendToSync(buf)
	}
	return err
}

func PongHandle(data msgCommon.MsgPayload, p2p *P2PServer) error {
	log.Debug("RX pong message")
	length := len(data.Payload)

	var pong msg.Pong
	pong.Deserialization(data.Payload[:length])
	pong.Verify(data.Payload[msgCommon.MSG_HDR_LEN:length])

	remotePeer := p2p.Self.Np.GetPeer(data.Id)
	remotePeer.SetHeight(pong.Height)
	return nil
}

// BlkHeaderHandle handles the sync headers from peer
func BlkHeaderHandle(data msgCommon.MsgPayload, p2p *P2PServer) error {
	log.Debug("RX block header message")
	length := len(data.Payload)
	var blkHeader msg.BlkHeader
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
func BlockHandle(data msgCommon.MsgPayload, p2p *P2PServer) error {
	log.Debug("RX block message")
	length := len(data.Payload)

	var block msg.Block
	block.Deserialization(data.Payload[:length])
	block.Verify(data.Payload[msgCommon.MSG_HDR_LEN:length])

	hash := block.Blk.Hash()
	if con, _ := actor.IsContainBlock(hash); con != true {
		actor.AddBlock(&block.Blk)
	} else {
		log.Debug("Receive duplicated block")
	}
	return nil
}

// ConsensusHandle handles the consensus message from peer
func ConsensusHandle(data msgCommon.MsgPayload, p2p *P2PServer) error {
	log.Debug("RX consensus message")
	length := len(data.Payload)

	var consensus msg.Consensus
	consensus.Deserialization(data.Payload[:length])
	consensus.Cons.Verify()

	if actor.ConsensusPid != nil {
		actor.ConsensusPid.Tell(&consensus.Cons)
	}
	return nil
}

// NotFoundHandle handles the not found message from peer
func NotFoundHandle(data msgCommon.MsgPayload, p2p *P2PServer) error {
	length := len(data.Payload)

	var notFound msg.NotFound
	notFound.Deserialization(data.Payload[:length])
	log.Debug("RX notFound message, hash is ", notFound.Hash)
	return nil
}

// TransactionHandle handles the transaction message from peer
func TransactionHandle(data msgCommon.MsgPayload, p2p *P2PServer) error {
	log.Debug("RX transaction message")
	length := len(data.Payload)

	var trn msg.Trn
	trn.Deserialization(data.Payload[:length])

	tx := &trn.Txn
	if _, err := actor.GetTransaction(tx.Hash()); err == nil {
		actor.AddTransaction(tx)
		log.Debug("RX Transaction message hash", tx.Hash())
	}
	return nil
}

// VersionHandle handles version handshake protocol from peer
func VersionHandle(data msgCommon.MsgPayload, p2p *P2PServer) error {
	log.Debug("RX version message")
	length := len(data.Payload)

	if length == 0 {
		log.Error(fmt.Sprintf("nil message for %s", msgCommon.VERSION_TYPE))
		return errors.New("nil message")
	}

	version := msg.Version{}
	copy(version.Hdr.CMD[0:len(msgCommon.VERSION_TYPE)], msgCommon.VERSION_TYPE)

	version.Deserialization(data.Payload[:length])
	version.Verify(data.Payload[msgCommon.MSG_HDR_LEN:length])

	localPeer := p2p.Self

	if version.P.IsConsensus == true {
		remotePeer := p2p.network.GetPeerFromAddr(data.Addr)

		if remotePeer == nil {
			log.Warn(" peer is not exist")
			return errors.New("peer is not exist ")
		}
		p := p2p.Self.Np.GetPeer(version.P.Nonce)

		if p == nil {
			log.Warn("sync link is not exist")
			return errors.New("sync link is not exist")
		} else {
			//p synclink must exist,merged
			p.ConsLink = remotePeer.ConsLink
			p.SetConsState(remotePeer.GetConsState())
			remotePeer = p

		}
		if version.P.Nonce == p2p.Self.GetID() {
			log.Warn("The node handshake with itself")
			remotePeer.CloseCons()
			return errors.New("The node handshake with itself ")
		}

		s := remotePeer.GetConsState()
		if s != msgCommon.INIT && s != msgCommon.HAND {
			log.Warn("Unknown status to received version")
			return errors.New("Unknown status to received version ")
		}

		remotePeer.UpdateInfo(time.Now(), version.P.Version, version.P.Services,
			version.P.SyncPort, version.P.ConsPort, version.P.Nonce,
			version.P.Relay, version.P.StartHeight)

		var buf []byte
		if s == msgCommon.INIT {
			remotePeer.SetConsState(msgCommon.HANDSHAKE)
			vpl := msgpack.NewVersionPayload(localPeer, true)
			buf, _ = msgpack.NewVersion(vpl, p2p.Self.GetPubKey())
		} else if s == msgCommon.HAND {
			remotePeer.SetConsState(msgCommon.HANDSHAKED)
			buf, _ = msgpack.NewVerAck(true)
		}
		remotePeer.SendToCons(buf)
		return nil
	} else {
		remotePeer := p2p.network.GetPeerFromAddr(data.Addr)

		if remotePeer == nil {
			log.Warn("peer is not exist")
			return errors.New("peer is not exist ")
		}
		if version.P.Nonce == p2p.Self.GetID() {
			log.Warn("The node handshake with itself")
			remotePeer.CloseSync()
			return errors.New("The node handshake with itself ")
		}

		s := remotePeer.GetSyncState()
		if s != msgCommon.INIT && s != msgCommon.HAND {
			log.Warn("Unknown status to received version")
			return errors.New("Unknown status to received version")
		}

		// Obsolete node
		n, ret := p2p.Self.Np.DelNbrNode(version.P.Nonce)
		if ret == true {
			log.Info(fmt.Sprintf("Peer reconnect 0x%x", version.P.Nonce))
			// Close the connection and release the node source
			n.SetSyncState(msgCommon.INACTIVITY)
			n.CloseSync()
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
		remotePeer.UpdateInfo(time.Now(), version.P.Version, version.P.Services,
			version.P.SyncPort, version.P.ConsPort, version.P.Nonce,
			version.P.Relay, version.P.StartHeight)

		p2p.Self.Np.AddNbrNode(remotePeer)

		var buf []byte
		if s == msgCommon.INIT {
			remotePeer.SetSyncState(msgCommon.HANDSHAKE)
			vpl := msgpack.NewVersionPayload(localPeer, false)
			buf, _ = msgpack.NewVersion(vpl, p2p.Self.GetPubKey())
		} else if s == msgCommon.HAND {
			remotePeer.SetSyncState(msgCommon.HANDSHAKED)
			buf, _ = msgpack.NewVerAck(false)
		}
		remotePeer.SendToSync(buf)
		return nil
	}
	return nil
}

// VerAckHandle handles the version ack from peer
func VerAckHandle(data msgCommon.MsgPayload, p2p *P2PServer) error {
	log.Debug("RX verAck message")

	length := len(data.Payload)

	if length == 0 {
		log.Error(fmt.Sprintf("nil message for %s", msgCommon.VERACK_TYPE))
		return errors.New("nil message")
	}

	verAck := msg.VerACK{}
	verAck.Deserialization(data.Payload[:length])

	localPeer := p2p.Self
	remotePeer := localPeer.Np.GetPeer(data.Id)

	if remotePeer == nil {
		log.Warn("nbr node is not exist")
		return errors.New("nbr node is not exist ")
	}

	if verAck.IsConsensus == true {
		s := remotePeer.GetConsState()
		if s != msgCommon.HANDSHAKE && s != msgCommon.HANDSHAKED {
			log.Warn("Unknown status to received verAck")
			return errors.New("Unknown status to received verAck")
		}

		remotePeer.SetConsState(msgCommon.ESTABLISH)
		remotePeer.SetConsConn(remotePeer.GetConsConn())

		if s == msgCommon.HANDSHAKE {
			buf, _ := msgpack.NewVerAck(true)
			remotePeer.SendToCons(buf)
		}
		return nil
	} else {
		s := remotePeer.GetSyncState()
		if s != msgCommon.HANDSHAKE && s != msgCommon.HANDSHAKED {
			log.Warn("Unknown status to received verAck")
			return errors.New("Unknown status to received verAck ")
		}

		remotePeer.SetSyncState(msgCommon.ESTABLISH)

		if s == msgCommon.HANDSHAKE {
			buf, _ := msgpack.NewVerAck(false)
			remotePeer.SendToSync(buf)
		}

		remotePeer.DumpInfo()

		buf, _ := msgpack.NewAddrReq()
		go remotePeer.SendToSync(buf)

		addr := remotePeer.GetAddr()
		// port := remotePeer.GetSyncPort()
		// nodeAddr := addr + ":" + strconv.Itoa(int(port))

		// p2p.Self.SyncLink.RemoveAddrInConnectingList(nodeAddr)

		//connect consensus port
		if s == msgCommon.HANDSHAKED {
			consensusPort := remotePeer.GetConsPort()
			nodeConsensusAddr := addr + ":" + strconv.Itoa(int(consensusPort))
			//Fix me:
			go p2p.network.Connect(nodeConsensusAddr, true)
		}
		return nil
	}
	return nil
}

// AddrHandle handles the neighbor address response message from peer
func AddrHandle(data msgCommon.MsgPayload, p2p *P2PServer) error {
	log.Debug("Handle addr message")
	localPeer := p2p.Self
	//remotePeer := p2p.Self.Np.GetPeer(data.Id)
	length := len(data.Payload)

	var msg msg.Addr
	msg.Deserialization(data.Payload[:length])
	msg.Verify(data.Payload[msgCommon.MSG_HDR_LEN:length])

	for _, v := range msg.NodeAddrs {
		var ip net.IP
		ip = v.IpAddr[:]
		address := ip.To16().String() + ":" + strconv.Itoa(int(v.Port))
		log.Info(fmt.Sprintf("The ip address is %s id is 0x%x", address, v.ID))

		if v.ID == localPeer.GetID() {
			continue
		}

		if localPeer.Np.NodeEstablished(v.ID) {
			continue
		}

		if v.Port == 0 {
			continue
		}

		go p2p.network.Connect(address, false)
	}
	return nil
}

// DataReqHandle handles the data req(block/Transaction) from peer
func DataReqHandle(data msgCommon.MsgPayload, p2p *P2PServer) error {
	log.Debug("RX data req message")
	length := len(data.Payload)

	var dataReq msg.DataReq
	dataReq.Deserialization(data.Payload[:length])

	//localPeer := p2p.Self
	remotePeer := p2p.Self.Np.GetPeer(data.Id)

	reqType := common.InventoryType(dataReq.DataType)
	hash := dataReq.Hash
	switch reqType {
	case common.BLOCK:
		block, err := actor.GetBlockByHash(hash)
		if err != nil {
			log.Debug("Can't get block by hash: ", hash, " ,send not found message")
			b, err := msgpack.NewNotFound(hash)
			remotePeer.SendToSync(b)
			return err
		}
		log.Debug("block height is ", block.Header.Height, " ,hash is ", hash)
		buf, err := msgpack.NewBlock(block)
		if err != nil {
			return err
		}
		remotePeer.SendToSync(buf)

	case common.TRANSACTION:
		txn, err := actor.GetTxnFromLedger(hash)
		if err != nil {
			log.Debug("Can't get transaction by hash: ", hash, " ,send not found message")
			b, err := msgpack.NewNotFound(hash)
			remotePeer.SendToSync(b)
			return err
		}
		buf, err := msgpack.NewTxn(txn)
		if err != nil {
			return err
		}
		remotePeer.SendToSync(buf)
	}
	return nil
}

// InvHandle handles the inventory message(block, transaction and consensus) from peer.
func InvHandle(data msgCommon.MsgPayload, p2p *P2PServer) error {
	log.Debug("RX inv message")
	length := len(data.Payload)
	var inv msg.Inv
	inv.Deserialization(data.Payload[:length])
	inv.Verify(data.Payload[msgCommon.MSG_HDR_LEN:length])

	//localPeer := p2p.Self
	remotePeer := p2p.Self.Np.GetPeer(data.Id)

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
			remotePeer.SendToSync(txnDataReq)
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
			if !isContainBlock && msg.LastInvHash != id {
				msg.LastInvHash = id
				// send the block request
				log.Infof("inv request block hash: %x", id)
				blkDataReq, _ := msgpack.NewBlkDataReq(id)
				remotePeer.SendToSync(blkDataReq)
			}
		}
	case common.CONSENSUS:
		log.Debug("RX consensus message")
		id.Deserialize(bytes.NewReader(inv.P.Blk[:32]))
		consDataReq, _ := msgpack.NewConsensusDataReq(id)
		remotePeer.SendToCons(consDataReq)
	default:
		log.Warn("RX unknown inventory message")
	}
	return nil
}

//
func DisconnectHandle(data msgCommon.MsgPayload, p2p *P2PServer) error {
	remotePeer := p2p.Self.Np.GetPeer(data.Id)
	i := strings.Index(data.Addr, ":")
	if i < 0 {
		log.Error("link address format error", data.Addr)
	}
	port, err := strconv.Atoi(data.Addr[i:])
	if err != nil {
		log.Error("Split port error", data.Addr[i:])
	}
	if remotePeer.SyncLink.GetPort() == uint16(port) {
		remotePeer.CloseSync()
		remotePeer.CloseCons()
	}
	if remotePeer.ConsLink.GetPort() == uint16(port) {
		remotePeer.CloseCons()
	}
	return nil
}
