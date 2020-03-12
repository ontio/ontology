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

package protocols

import (
	"errors"
	"fmt"
	"net"
	"strconv"

	lru "github.com/hashicorp/golang-lru"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/types"
	actor "github.com/ontio/ontology/p2pserver/actor/req"
	msgCommon "github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/dht"
	msgpack "github.com/ontio/ontology/p2pserver/message/msg_pack"
	msgTypes "github.com/ontio/ontology/p2pserver/message/types"
	p2p "github.com/ontio/ontology/p2pserver/net/protocol"
	"github.com/ontio/ontology/p2pserver/protocols/block_sync"
	"github.com/ontio/ontology/p2pserver/protocols/reconnect"
)

//respCache cache for some response data
var respCache *lru.ARCCache

//Store txHash, using for rejecting duplicate tx
// thread safe
var txCache, _ = lru.NewARC(msgCommon.MAX_TX_CACHE_SIZE)

type MsgHandler struct {
	blockSync *block_sync.BlockSyncMgr
	reconnect *reconnect.ReconnectService
	ledger    *ledger.Ledger
}

func NewMsgHandler(ld *ledger.Ledger) *MsgHandler {
	return &MsgHandler{ledger: ld}
}

func (self *MsgHandler) start(net p2p.P2P) {
	self.blockSync = block_sync.NewBlockSyncMgr(net, self.ledger)
	self.reconnect = reconnect.NewReconectService(net)

	go self.blockSync.Start()
	go self.reconnect.Start()
}

func (self *MsgHandler) stop() {
	self.blockSync.Close()
	self.reconnect.Close()
}

func (self *MsgHandler) HandleSystemMessage(net p2p.P2P, msg SystemMessage) {
	switch m := msg.(type) {
	case NetworkStart:
		self.start(net)
	case PeerConnected:
		self.blockSync.OnAddNode(m.Info.Id.ToUint64())
		self.reconnect.OnAddPeer(m.Info)
	case PeerDisConnected:
		self.blockSync.OnDelNode(m.Info.Id.ToUint64())
		self.reconnect.OnDelPeer(m.Info)
	case NetworkStop:
		self.stop()
	}
}

func (self *MsgHandler) HandlePeerMessage(ctx *Context, msg msgTypes.Message) {
	log.Trace("[p2p]receive message", ctx.Sender().GetAddr(), ctx.Sender().GetID())
	switch m := msg.(type) {
	case *msgTypes.AddrReq:
		AddrReqHandle(ctx)
	case *msgTypes.FindNodeResp:
		FindNodeResponseHandle(ctx, m)
	case *msgTypes.FindNodeReq:
		FindNodeHandle(ctx, m)
	case *msgTypes.HeadersReq:
		HeadersReqHandle(ctx, m)
	case *msgTypes.Ping:
		PingHandle(ctx, m)
	case *msgTypes.Pong:
		PongHandle(ctx, m)
	case *msgTypes.BlkHeader:
		self.blockSync.OnHeaderReceive(ctx.Sender().GetID(), m.BlkHdr)
	case *msgTypes.Block:
		self.blockHandle(ctx, m)
	case *msgTypes.Consensus:
		ConsensusHandle(ctx, m)
	case *msgTypes.Trn:
		TransactionHandle(ctx, m)
	case *msgTypes.Addr:
		AddrHandle(ctx, m)
	case *msgTypes.DataReq:
		DataReqHandle(ctx, m)
	case *msgTypes.Inv:
		InvHandle(ctx, m)
	case *msgTypes.Disconnected:
		DisconnectHandle(ctx)
	case *msgTypes.NotFound:
		log.Debug("[p2p]receive notFound message, hash is ", m.Hash)
	default:
		msgType := msg.CmdType()
		if msgType == msgCommon.VERACK_TYPE || msgType == msgCommon.VERSION_TYPE {
			log.Infof("receive message: %s from peer %s", msgType, ctx.Sender().GetAddr())
		} else {
			log.Warn("unknown message handler for the msg: ", msgType)
		}
	}
}

// AddrReqHandle handles the neighbor address request from peer
func AddrReqHandle(ctx *Context) {
	remotePeer := ctx.Sender()
	p2p := ctx.Network()

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
	err := remotePeer.Send(msg)

	if err != nil {
		log.Warn(err)
		return
	}
}

func FindNodeResponseHandle(ctx *Context, fresp *msgTypes.FindNodeResp) {
	if fresp.Success {
		log.Debugf("[p2p dht] %s", "find peer success, do nothing")
		return
	}
	p2p := ctx.Network()
	// we should connect to closer peer to ask them them where should we go
	for _, curpa := range fresp.CloserPeers {
		// already connected
		if p2p.GetPeer(curpa.PeerID) != nil {
			continue
		}
		// do nothing about
		if curpa.PeerID == p2p.GetID() {
			continue
		}
		log.Debugf("[dht] try to connect to another peer by dht: %d ==> %s", curpa.PeerID, curpa.Addr)
		go p2p.Connect(curpa.Addr)
	}
}

// FindNodeHandle handles the neighbor address request from peer
func FindNodeHandle(ctx *Context, freq *msgTypes.FindNodeReq) {
	// we recv message must from establised peer
	remotePeer := ctx.Sender()

	var fresp msgTypes.FindNodeResp
	// check the target is my self
	log.Debugf("[dht] find node for peerid: %d", freq.TargetID)
	p2p := ctx.Network()
	if freq.TargetID == p2p.GetKadKeyId().Id {
		fresp.Success = true
		fresp.TargetID = freq.TargetID
		// you've already connected with me so there's no need to give you my address
		// omit the address
		if err := remotePeer.Send(&fresp); err != nil {
			log.Warn(err)
		}
		return
	}
	// search dht
	closer := p2p.BetterPeers(freq.TargetID, dht.AlphaValue)

	paddrs := p2p.GetPeerStringAddr()
	for _, kid := range closer {
		pid := kid.ToUint64()
		if addr, ok := paddrs[pid]; ok {
			curAddr := msgTypes.PeerAddr{
				Addr:   addr,
				PeerID: pid,
			}
			fresp.CloserPeers = append(fresp.CloserPeers, curAddr)

		}
	}
	fresp.TargetID = freq.TargetID
	log.Debugf("[dht] find %d more closer peers:", len(fresp.CloserPeers))
	for _, curpa := range fresp.CloserPeers {
		log.Debugf("    dht: pid: %d, addr: %s", curpa.PeerID, curpa.Addr)
	}

	if err := remotePeer.Send(&fresp); err != nil {
		log.Warn(err)
	}
}

// HeaderReqHandle handles the header sync req from peer
func HeadersReqHandle(ctx *Context, headersReq *msgTypes.HeadersReq) {
	startHash := headersReq.HashStart
	stopHash := headersReq.HashEnd

	headers, err := GetHeadersFromHash(startHash, stopHash)
	if err != nil {
		log.Warnf("HeadersReqHandle error: %s,startHash:%s,stopHash:%s", err.Error(), startHash.ToHexString(), stopHash.ToHexString())
		return
	}
	remotePeer := ctx.Sender()
	msg := msgpack.NewHeaders(headers)
	err = remotePeer.Send(msg)
	if err != nil {
		log.Warn(err)
		return
	}
}

//PingHandle handle ping msg from peer
func PingHandle(ctx *Context, ping *msgTypes.Ping) {
	remotePeer := ctx.Sender()
	remotePeer.SetHeight(ping.Height)
	p2p := ctx.Network()

	height := ledger.DefLedger.GetCurrentBlockHeight()
	p2p.SetHeight(uint64(height))
	msg := msgpack.NewPongMsg(uint64(height))

	err := remotePeer.Send(msg)
	if err != nil {
		log.Warn(err)
	}
}

///PongHandle handle pong msg from peer
func PongHandle(ctx *Context, pong *msgTypes.Pong) {
	remotePeer := ctx.Network()
	remotePeer.SetHeight(pong.Height)
}

// blockHandle handles the block message from peer
func (self *MsgHandler) blockHandle(ctx *Context, block *msgTypes.Block) {
	stateHashHeight := config.GetStateHashCheckHeight(config.DefConfig.P2PNode.NetworkId)
	if block.Blk.Header.Height >= stateHashHeight && block.MerkleRoot == common.UINT256_EMPTY {
		remotePeer := ctx.Sender()
		remotePeer.Close()
		return
	}

	self.blockSync.OnBlockReceive(ctx.Sender().GetID(), ctx.msgSize, block.Blk, block.CCMsg, block.MerkleRoot)
}

// ConsensusHandle handles the consensus message from peer
func ConsensusHandle(ctx *Context, consensus *msgTypes.Consensus) {
	if actor.ConsensusPid != nil {
		if err := consensus.Cons.Verify(); err != nil {
			log.Warn(err)
			return
		}
		consensus.Cons.PeerId = ctx.Sender().GetID()
		actor.ConsensusPid.Tell(&consensus.Cons)
	}
}

// TransactionHandle handles the transaction message from peer
func TransactionHandle(ctx *Context, trn *msgTypes.Trn) {
	if !txCache.Contains(trn.Txn.Hash()) {
		txCache.Add(trn.Txn.Hash(), nil)
		actor.AddTransaction(trn.Txn)
	} else {
		log.Tracef("[p2p]receive duplicate Transaction message, txHash: %x\n", trn.Txn.Hash())
	}
}

// AddrHandle handles the neighbor address response message from peer
func AddrHandle(ctx *Context, msg *msgTypes.Addr) {
	p2p := ctx.Network()
	for _, v := range msg.NodeAddrs {
		if v.Port == 0 || v.ID == p2p.GetID() {
			continue
		}
		var ip net.IP
		ip = v.IpAddr[:]
		address := ip.To16().String() + ":" + strconv.Itoa(int(v.Port))

		if p2p.NodeEstablished(v.ID) {
			continue
		}

		log.Debug("[p2p]connect ip address:", address)
		go p2p.Connect(address)
	}
}

// DataReqHandle handles the data req(block/Transaction) from peer
func DataReqHandle(ctx *Context, dataReq *msgTypes.DataReq) {
	remotePeer := ctx.Sender()
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
				log.Debug("[p2p]can't get block by hash: ", hash, " ,send not found message")
				msg := msgpack.NewNotFound(hash)
				err := remotePeer.Send(msg)
				if err != nil {
					log.Warn(err)
					return
				}
				return
			}
			ccMsg, err := ledger.DefLedger.GetCrossChainMsg(block.Header.Height - 1)
			if err != nil {
				log.Debugf("[p2p]failed to get cross chain message at height %v, err %v",
					block.Header.Height-1, err)
				msg := msgpack.NewNotFound(hash)
				err := remotePeer.Send(msg)
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
				err := remotePeer.Send(msg)
				if err != nil {
					log.Warn(err)
					return
				}
				return
			}
			msg = msgpack.NewBlock(block, ccMsg, merkleRoot)
			saveRespCache(reqID, msg)
		}
		err := remotePeer.Send(msg)
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
			err = remotePeer.Send(msg)
			if err != nil {
				log.Warn(err)
				return
			}
		}
		msg := msgpack.NewTxn(txn)
		err = remotePeer.Send(msg)
		if err != nil {
			log.Warn(err)
			return
		}
	}
}

// InvHandle handles the inventory message(block,
// transaction and consensus) from peer.
func InvHandle(ctx *Context, inv *msgTypes.Inv) {
	remotePeer := ctx.Sender()
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
			err = remotePeer.Send(msg)
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
				err = remotePeer.Send(msg)
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
		err := remotePeer.Send(msg)
		if err != nil {
			log.Warn(err)
			return
		}
	default:
		log.Warn("[p2p]receive unknown inventory message")
	}

}

// DisconnectHandle handles the disconnect events
func DisconnectHandle(ctx *Context) {
	remotePeer := ctx.Sender()
	p2p := ctx.Network()
	if remotePeer == nil {
		log.Debug("[p2p]disconnect peer is nil")
		return
	}

	if remotePeer.Link.GetAddr() == remotePeer.GetAddr() {
		p2p.RemovePeerAddress(remotePeer.GetAddr())
		remotePeer.Close()
	}
	p2p.RemoveDHT(remotePeer.GetKId())
}

//get blk hdrs from starthash to stophash
func GetHeadersFromHash(startHash common.Uint256, stopHash common.Uint256) ([]*types.RawHeader, error) {
	var count uint32 = 0
	var headers []*types.RawHeader
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
