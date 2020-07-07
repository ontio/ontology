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

	lru "github.com/hashicorp/golang-lru"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/types"
	actor "github.com/ontio/ontology/p2pserver/actor/req"
	msgCommon "github.com/ontio/ontology/p2pserver/common"
	msgpack "github.com/ontio/ontology/p2pserver/message/msg_pack"
	msgTypes "github.com/ontio/ontology/p2pserver/message/types"
	p2p "github.com/ontio/ontology/p2pserver/net/protocol"
	"github.com/ontio/ontology/p2pserver/protocols/block_sync"
	"github.com/ontio/ontology/p2pserver/protocols/bootstrap"
	"github.com/ontio/ontology/p2pserver/protocols/discovery"
	"github.com/ontio/ontology/p2pserver/protocols/heatbeat"
	"github.com/ontio/ontology/p2pserver/protocols/recent_peers"
	"github.com/ontio/ontology/p2pserver/protocols/reconnect"
	"github.com/ontio/ontology/p2pserver/protocols/subnet"
	"github.com/ontio/ontology/p2pserver/protocols/utils"
)

//respCache cache for some response data
var respCache, _ = lru.NewARC(msgCommon.MAX_RESP_CACHE_SIZE)

//Store txHash, using for rejecting duplicate tx
// thread safe
var txCache, _ = lru.NewARC(msgCommon.MAX_TX_CACHE_SIZE)

type MsgHandler struct {
	seeds                    *utils.HostsResolver
	blockSync                *block_sync.BlockSyncMgr
	reconnect                *reconnect.ReconnectService
	discovery                *discovery.Discovery
	heatBeat                 *heatbeat.HeartBeat
	bootstrap                *bootstrap.BootstrapService
	persistRecentPeerService *recent_peers.PersistRecentPeerService
	subnet                   *subnet.SubNet
	ledger                   *ledger.Ledger
	acct                     *account.Account // nil if conenesus is not enabled
	staticReserveFilter      p2p.AddressFilter
}

func NewMsgHandler(acct *account.Account, staticReserveFilter p2p.AddressFilter, ld *ledger.Ledger, logger msgCommon.Logger) *MsgHandler {
	gov := utils.NewGovNodeResolver(ld)
	seedsList := config.DefConfig.Genesis.SeedList
	seeds, invalid := utils.NewHostsResolver(seedsList)
	if invalid != nil {
		panic(fmt.Errorf("invalid seed list； %v", invalid))
	}
	subNet := subnet.NewSubNet(acct, seeds, gov, logger)
	return &MsgHandler{ledger: ld, seeds: seeds, subnet: subNet, acct: acct, staticReserveFilter: staticReserveFilter}
}

func (self *MsgHandler) GetReservedAddrFilter(staticFilterEnabled bool) p2p.AddressFilter {
	return self.subnet.GetReservedAddrFilter(staticFilterEnabled)
}

func (self *MsgHandler) GetMaskAddrFilter() p2p.AddressFilter {
	return self.subnet.GetMaskAddrFilter()
}

func (self *MsgHandler) GetSubnetMembersInfo() []msgCommon.SubnetMemberInfo {
	return self.subnet.GetMembersInfo()
}

func (self *MsgHandler) start(net p2p.P2P) {
	self.blockSync = block_sync.NewBlockSyncMgr(net, self.ledger)
	self.reconnect = reconnect.NewReconectService(net, self.staticReserveFilter)
	maskFilter := self.subnet.GetMaskAddrFilter()
	self.discovery = discovery.NewDiscovery(net, config.DefConfig.P2PNode.ReservedCfg.MaskPeers, maskFilter, 0)
	self.bootstrap = bootstrap.NewBootstrapService(net, self.seeds)
	self.heatBeat = heatbeat.NewHeartBeat(net, self.ledger)
	self.persistRecentPeerService = recent_peers.NewPersistRecentPeerService(net)
	go self.persistRecentPeerService.Start()
	go self.blockSync.Start()
	go self.reconnect.Start()
	go self.discovery.Start()
	go self.heatBeat.Start()
	go self.bootstrap.Start()
	go self.subnet.Start(net)
}

func (self *MsgHandler) stop() {
	self.blockSync.Stop()
	self.reconnect.Stop()
	self.discovery.Stop()
	self.persistRecentPeerService.Stop()
	self.heatBeat.Stop()
	self.bootstrap.Stop()
	self.subnet.Stop()
}

func (self *MsgHandler) HandleSystemMessage(net p2p.P2P, msg p2p.SystemMessage) {
	switch m := msg.(type) {
	case p2p.NetworkStart:
		self.start(net)
	case p2p.PeerConnected:
		self.blockSync.OnAddNode(m.Info.Id)
		self.reconnect.OnAddPeer(m.Info)
		self.discovery.OnAddPeer(m.Info)
		self.bootstrap.OnAddPeer(m.Info)
		self.persistRecentPeerService.AddNodeAddr(m.Info.RemoteListenAddress())
		self.subnet.OnAddPeer(net, m.Info)
	case p2p.PeerDisConnected:
		self.blockSync.OnDelNode(m.Info.Id)
		self.reconnect.OnDelPeer(m.Info)
		self.discovery.OnDelPeer(m.Info)
		self.bootstrap.OnDelPeer(m.Info)
		self.subnet.OnDelPeer(m.Info)
		self.persistRecentPeerService.DelNodeAddr(m.Info.RemoteListenAddress())
	case p2p.NetworkStop:
		self.stop()
	case p2p.HostAddrDetected:
		self.subnet.OnHostAddrDetected(m.ListenAddr)
	}
}

func (self *MsgHandler) HandlePeerMessage(ctx *p2p.Context, msg msgTypes.Message) {
	log.Trace("[p2p]receive message", ctx.Sender().GetAddr(), ctx.Sender().GetID())
	switch m := msg.(type) {
	case *msgTypes.AddrReq:
		self.discovery.AddrReqHandle(ctx)
	case *msgTypes.FindNodeResp:
		self.discovery.FindNodeResponseHandle(ctx, m)
	case *msgTypes.FindNodeReq:
		self.discovery.FindNodeHandle(ctx, m)
	case *msgTypes.HeadersReq:
		HeadersReqHandle(ctx, m)
	case *msgTypes.Ping:
		self.heatBeat.PingHandle(ctx, m)
	case *msgTypes.Pong:
		self.heatBeat.PongHandle(ctx, m)
	case *msgTypes.BlkHeader:
		self.blockSync.OnHeaderReceive(ctx.Sender().GetID(), m.BlkHdr)
	case *msgTypes.Block:
		self.blockHandle(ctx, m)
	case *msgTypes.Consensus:
		ConsensusHandle(ctx, m)
	case *msgTypes.Trn:
		TransactionHandle(ctx, m)
	case *msgTypes.Addr:
		self.discovery.AddrHandle(ctx, m)
	case *msgTypes.DataReq:
		DataReqHandle(ctx, m)
	case *msgTypes.Inv:
		InvHandle(ctx, m)
	case *msgTypes.SubnetMembersRequest:
		self.subnet.OnMembersRequest(ctx, m)
	case *msgTypes.SubnetMembers:
		self.subnet.OnMembersResponse(ctx, m)
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

// HeaderReqHandle handles the header sync req from peer
func HeadersReqHandle(ctx *p2p.Context, headersReq *msgTypes.HeadersReq) {
	startHash := headersReq.HashStart
	stopHash := headersReq.HashEnd

	headers, err := GetHeadersFromHash(startHash, stopHash)
	if err != nil {
		log.Debugf("HeadersReqHandle error: %s,startHash:%s,stopHash:%s", err.Error(), startHash.ToHexString(), stopHash.ToHexString())
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

// blockHandle handles the block message from peer
func (self *MsgHandler) blockHandle(ctx *p2p.Context, block *msgTypes.Block) {
	stateHashHeight := config.GetStateHashCheckHeight(config.DefConfig.P2PNode.NetworkId)
	if block.Blk.Header.Height >= stateHashHeight && block.MerkleRoot == common.UINT256_EMPTY {
		remotePeer := ctx.Sender()
		remotePeer.Close()
		return
	}

	self.blockSync.OnBlockReceive(ctx.Sender().GetID(), ctx.MsgSize, block.Blk, block.CCMsg, block.MerkleRoot)
}

// ConsensusHandle handles the consensus message from peer
func ConsensusHandle(ctx *p2p.Context, consensus *msgTypes.Consensus) {
	cpid := actor.GetConsensusPid()
	if cpid != nil {
		if err := consensus.Cons.Verify(); err != nil {
			log.Warn(err)
			return
		}
		consensus.Cons.PeerId = ctx.Sender().GetID()
		cpid.Tell(&consensus.Cons)
	}
}

// TransactionHandle handles the transaction message from peer
func TransactionHandle(ctx *p2p.Context, trn *msgTypes.Trn) {
	if !txCache.Contains(trn.Txn.Hash()) {
		txCache.Add(trn.Txn.Hash(), nil)
		actor.AddTransaction(trn.Txn)
	} else {
		log.Tracef("[p2p]receive duplicate Transaction message, txHash: %x\n", trn.Txn.Hash())
	}
}

// DataReqHandle handles the data req(block/Transaction) from peer
func DataReqHandle(ctx *p2p.Context, dataReq *msgTypes.DataReq) {
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
func InvHandle(ctx *p2p.Context, inv *msgTypes.Inv) {
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
	data, ok := respCache.Get(key)
	if ok {
		return data
	}
	return nil
}

//saveRespCache save response msg to cache
func saveRespCache(key string, value interface{}) {
	respCache.Add(key, value)
}

func (mh *MsgHandler) ReconnectService() *reconnect.ReconnectService {
	return mh.reconnect
}
