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

//Store txHash, using for rejecting duplicate tx
// thread safe
var txCache, _ = lru.NewARC(msgCommon.MAX_TX_CACHE_SIZE)

// AddrReqHandle handles the neighbor address request from peer
func AddrReqHandle(data *msgTypes.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Trace("[p2p]receive addr request message", data.Addr, data.Id)
	remotePeer := p2p.GetPeer(data.Id)
	if remotePeer == nil {
		log.Debug("[p2p]remotePeer invalid in AddrReqHandle")
		return
	}

	addrStr := p2p.GetNeighborAddrs()
	//check mask peers
	mskPeers := config.DefConfig.P2PNode.ReservedCfg.MaskPeers
	if config.DefConfig.P2PNode.ReservedPeersOnly && len(mskPeers) > 0 {
		mskPeerMap := make(map[string]bool)
		for _, mskAddr := range mskPeers {
			mskPeerMap[mskAddr] = true
		}

		// get remote peer IP
		// if get remotePeerAddr failed, do masking anyway
		remoteAddr, _ := remotePeer.GetAddr16()
		var remoteIp net.IP = remoteAddr[:]

		// remove msk peers from neigh-addr-list
		// if remotePeer is in msk-list, skip masking
		if _, isMskPeer := mskPeerMap[remoteIp.String()]; !isMskPeer {
			mskAddrList := make([]msgCommon.PeerAddr, 0)
			for _, addr := range addrStr {
				var ip net.IP
				ip = addr.IpAddr[:]
				address := ip.To16().String()
				if _, present := mskPeerMap[address]; !present {
					mskAddrList = append(mskAddrList, addr)
				}
			}
			// replace with mskAddrList
			addrStr = mskAddrList
		}
	}

	msg := msgpack.NewAddrs(addrStr)
	err := p2p.Send(remotePeer, msg)
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
	err = p2p.Send(remotePeer, msg)
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

	err := p2p.Send(remotePeer, msg)
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
		stateHashHeight := config.GetStateHashCheckHeight(config.DefConfig.P2PNode.NetworkId)
		if block.Blk.Header.Height >= stateHashHeight && block.MerkleRoot == common.UINT256_EMPTY {
			log.Info("received block msg with empty merkle root")
			remotePeer := p2p.GetPeer(data.Id)
			if remotePeer != nil {
				remotePeer.Close()
			}

			return
		}

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

	if !txCache.Contains(trn.Txn.Hash()) {
		txCache.Add(trn.Txn.Hash(), nil)
		actor.AddTransaction(trn.Txn)
	} else {
		log.Tracef("[p2p]receive duplicate Transaction message, txHash: %x\n", trn.Txn.Hash())
	}
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
			remotePeer.Close()
			log.Debug("[p2p]peer not in reserved list,close", data.Addr)
			return
		}

	}

	if version.P.Nonce == p2p.GetID() {
		p2p.RemoveFromInConnRecord(remotePeer.GetAddr())
		p2p.RemoveFromOutConnRecord(remotePeer.GetAddr())
		log.Warn("[p2p]the node handshake with itself", remotePeer.GetAddr())
		p2p.SetOwnAddress(nodeAddr)
		remotePeer.Close()
		return
	}

	s := remotePeer.GetState()
	if s != msgCommon.INIT && s != msgCommon.HAND {
		log.Warnf("[p2p]unknown status to received version,%d,%s\n", s, remotePeer.GetAddr())
		remotePeer.Close()
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
			remotePeer.Close()
			log.Warn("[p2p]connecting peer %d ip format is wrong %s, close", version.P.Nonce, data.Addr)
			return
		}
		if ipNew == ipOld {
			//same id and same ip
			n, delOK := p2p.DelNbrNode(version.P.Nonce)
			if delOK {
				log.Infof("[p2p]peer reconnect %d", version.P.Nonce, data.Addr)
				// Close the connection and release the node source
				n.Close()
				if pid != nil {
					input := &msgCommon.RemovePeerID{
						ID: version.P.Nonce,
					}
					pid.Tell(input)
				}
			}
		} else {
			log.Warnf("[p2p]same peer id from different addr: %s, %s close latest one", ipOld, ipNew)
			remotePeer.Close()
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
		version.P.Services, version.P.SyncPort, version.P.Nonce,
		version.P.Relay, version.P.StartHeight, version.P.SoftVersion)
	remotePeer.Link.SetID(version.P.Nonce)
	p2p.AddNbrNode(remotePeer)

	if pid != nil {
		input := &msgCommon.AppendPeerID{
			ID: version.P.Nonce,
		}
		pid.Tell(input)
	}

	var msg msgTypes.Message
	if s == msgCommon.INIT {
		remotePeer.SetState(msgCommon.HAND_SHAKE)
		msg = msgpack.NewVersion(p2p, ledger.DefLedger.GetCurrentBlockHeight())
	} else if s == msgCommon.HAND {
		remotePeer.SetState(msgCommon.HAND_SHAKED)
		msg = msgpack.NewVerAck()
	}
	err = p2p.Send(remotePeer, msg)
	if err != nil {
		log.Warn(err)
		return
	}
}

// VerAckHandle handles the version ack from peer
func VerAckHandle(data *msgTypes.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Trace("[p2p]receive verAck message from ", data.Addr, data.Id)

	//verAck := data.Payload.(*msgTypes.VerACK)
	remotePeer := p2p.GetPeer(data.Id)

	if remotePeer == nil {
		log.Warn("[p2p]nbr node is not exist", data.Id, data.Addr)
		return
	}

	s := remotePeer.GetState()
	if s != msgCommon.HAND_SHAKE && s != msgCommon.HAND_SHAKED {
		log.Warnf("[p2p]unknown status to received verAck,state:%d,%s\n", s, data.Addr)
		return
	}

	remotePeer.SetState(msgCommon.ESTABLISH)
	p2p.RemoveFromConnectingList(data.Addr)
	remotePeer.DumpInfo()

	if s == msgCommon.HAND_SHAKE {
		msg := msgpack.NewVerAck()
		p2p.Send(remotePeer, msg)
	}

	msg := msgpack.NewAddrReq()
	go p2p.Send(remotePeer, msg)

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
		go p2p.Connect(address)
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
				err := p2p.Send(remotePeer, msg)
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
				err := p2p.Send(remotePeer, msg)
				if err != nil {
					log.Warn(err)
					return
				}
				return
			}
			msg = msgpack.NewBlock(block, merkleRoot)
			saveRespCache(reqID, msg)
		}
		err := p2p.Send(remotePeer, msg)
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
			err = p2p.Send(remotePeer, msg)
			if err != nil {
				log.Warn(err)
				return
			}
		}
		msg := msgpack.NewTxn(txn)
		err = p2p.Send(remotePeer, msg)
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
			err = p2p.Send(remotePeer, msg)
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
				err = p2p.Send(remotePeer, msg)
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
		err := p2p.Send(remotePeer, msg)
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

	if remotePeer.Link.GetAddr() == data.Addr {
		p2p.RemovePeerAddress(data.Addr)
		remotePeer.Close()
	}
}

//get blk hdrs from starthash to stophash
func GetHeadersFromHash(startHash common.Uint256, stopHash common.Uint256) ([]*types.RawHeader, error) {
	var count uint32 = 0
	headers := []*types.RawHeader{}
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
			bkStop, err := ledger.DefLedger.GetRawHeaderByHash(stopHash)
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
		bkStart, err := ledger.DefLedger.GetRawHeaderByHash(startHash)
		if err != nil || bkStart == nil {
			return nil, err
		}
		startHeight = bkStart.Height
		if stopHash != common.UINT256_EMPTY {
			bkStop, err := ledger.DefLedger.GetRawHeaderByHash(stopHash)
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
		header, err := ledger.DefLedger.GetHeaderByHash(hash)
		if err != nil {
			log.Debugf("[p2p]net_server GetBlockWithHeight failed with err=%s, hash=%x,height=%d\n", err.Error(), hash, stopHeight+i)
			return nil, err
		}

		sink := common.NewZeroCopySink(nil)
		header.Serialization(sink)

		hd := &types.RawHeader{
			Height:  header.Height,
			Payload: sink.Bytes(),
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
