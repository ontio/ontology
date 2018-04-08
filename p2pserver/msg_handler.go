package p2pserver

import (
	"net"
	"bytes"
	"strconv"
	"encoding/hex"
	"fmt"
	"time"
	"errors"

	"github.com/Ontology/p2pserver/peer"
	"github.com/Ontology/common"
	"github.com/Ontology/common/log"
	"github.com/Ontology/core/types"
	actor "github.com/Ontology/p2pserver/actor/req"
	msgCommon "github.com/Ontology/p2pserver/common"
	msg "github.com/Ontology/p2pserver/message"
)

func MsgHdrHandle(hdr msg.MsgHdr, peer peer.Peer, p2p P2PServer) error {
	log.Debug("RX MsgHdr message")
	return nil
}

func AddrReqHandle(addrReq msg.AddrReq, peer peer.Peer, p2p P2PServer) error {
	log.Debug("RX addr request message")
	var addrStr []msgCommon.PeerAddr
	var count uint64
	addrStr, count = p2p.GetNeighborAddrs()
	buf, err := NewAddrs(addrStr, count)
	if err != nil {
		return err
	}
	go peer.SendToSync(buf)
	return nil
}

func HeadersReqHandle(headReq msg.HeadersReq, peer peer.Peer, p2p P2PServer) error {
	log.Debug("RX headers request message")

	//Fix me:
	//node.LocalNode().AcqSyncReqSem()
	//defer node.LocalNode().RelSyncReqSem()

	var startHash [msgCommon.HASH_LEN]byte
	var stopHash [msgCommon.HASH_LEN]byte
	startHash = headReq.P.HashStart
	stopHash = headReq.P.HashEnd
	//FIXME if HeaderHashCount > 1
	headers, cnt, err := actor.GetHeadersFromHash(startHash, stopHash)
	if err != nil {
		return err
	}
	buf, err := NewHeaders(headers, cnt)
	if err != nil {
		return err
	}
	go peer.SendToSync(buf)
	return nil
}

func BlocksReqHandle(blocksReq msg.BlocksReq, peer peer.Peer, p2p P2PServer) error {
	log.Debug("RX blocks request message")
	var startHash common.Uint256
	var stopHash common.Uint256
	startHash = blocksReq.P.HashStart
	stopHash = blocksReq.P.HashStop

	//FIXME if HeaderHashCount > 1
	inv, err := actor.GetInvFromBlockHash(startHash, stopHash)
	if err != nil {
		return err
	}
	buf, err := NewInv(inv)
	if err != nil {
		return err
	}
	go peer.SendToSync(buf)
	return nil
}

func PingHandle(ping msg.Ping, peer peer.Peer, p2p P2PServer) error {
	log.Debug("RX ping message")
	peer.SetHeight(ping.Height)
	buf, err := NewPongMsg(p2p.Self.GetHeight())
	if err != nil {
		log.Error("failed build a new pong message")
	} else {
		go peer.SendToSync(buf)
	}
	return err
}

func PongHandle(pong msg.Pong, peer peer.Peer, p2p P2PServer) error {
	log.Debug("RX pong message")
	peer.SetHeight(pong.Height)
	return nil
}

func BlkHeaderHandle(blkHeader msg.BlkHeader, peer peer.Peer, p2p P2PServer) error {
	log.Debug("RX block header message")
	var blkHdr []*types.Header
	var i uint32
	for i = 0; i < blkHeader.Cnt; i++ {
		blkHdr = append(blkHdr, &blkHeader.BlkHdr[i])
	}
	actor.AddHeaders(blkHdr)
	return nil
}

func BlockHandle(block msg.Block, peer peer.Peer, p2p P2PServer) error {
	log.Debug("RX block message")
	hash := block.Blk.Hash()
	if con, _ := actor.IsContainBlock(hash); con != true {
		actor.AddBlock(&block.Blk)
	} else {
		log.Debug("Receive duplicated block")
	}
	return nil
}

func ConsensusHandle(cons msg.Consensus, peer peer.Peer, p2p P2PServer) error {
	log.Debug("RX consensus message")
	if actor.ConsensusPid != nil {
		actor.ConsensusPid.Tell(&cons.Cons)
	}
	return nil
}

func NotFoundHandle(notFound msg.NotFound, peer peer.Peer, p2p P2PServer) error {
	log.Debug("RX notFound message, hash is ", notFound.Hash)
	return nil
}

func TransactionHandle(trn msg.Trn, peer peer.Peer, p2p P2PServer) error {
	log.Debug("RX transaction message")
	tx := &trn.Txn
	if _, err := actor.GetTransaction(tx.Hash()); err == nil {
		actor.AddTransaction(tx)
		log.Debug("RX Transaction message hash", tx.Hash())
	}
	return nil
}

func VersionHandle(version msg.Version, peer peer.Peer, p2p P2PServer) error {
	log.Debug("RX version message")
	// Exclude the node itself
	if version.P.Nonce == p2p.Self.GetID() {
		if version.P.IsConsensus == false {
			log.Warn("The node handshake with itself")
			peer.CloseSync()
			return errors.New("The node handshake with itself ")
		}
		if version.P.IsConsensus == true {
			log.Warn("The node handshake with itself")
			peer.CloseCons()
			return errors.New("The node handshake with itself ")
		}
	}

	if version.P.IsConsensus == true {
		s := peer.GetConsState()
		if s != msgCommon.INIT && s != msgCommon.HAND {
			log.Warn("Unknown status to received version")
			return errors.New("Unknown status to received version ")
		}

		peer.UpdateInfo(time.Now(), version.P.Version, version.P.Services,
			version.P.Port, version.P.Nonce, version.P.Relay, version.P.StartHeight)
		peer.SetConsPort(version.P.ConsensusPort)

		var buf []byte
		if s == msgCommon.INIT {
			peer.SetConsState(msgCommon.HANDSHAKE)
			height, _ := actor.GetCurrentBlockHeight()
			vpl := NewVersionPayload(p2p.Self.GetVersion(), p2p.Self.GetServices(),p2p.Self.GetSyncPort(),
				p2p.Self.GetConsPort(), p2p.Self.GetID(), uint64(height), peer.GetRelay(), true)
			buf, _ = NewVersion(vpl, p2p.Self.GetPubKey())
		} else if s == msgCommon.HAND {
			peer.SetConsState(msgCommon.HANDSHAKED)
			buf, _ = NewVerAck(true)
		}
		peer.SendToCons(buf)
		return nil
	}

	s := peer.GetSyncState()
	if s != msgCommon.INIT && s != msgCommon.HAND {
		log.Warn("Unknown status to received version")
		return errors.New("Unknown status to received version")
	}

	// Obsolete node
	n, ret := p2p.Self.Np.DelNbrNode(version.P.Nonce)
	if ret == true {
		log.Info(fmt.Sprintf("Node reconnect 0x%x", version.P.Nonce))
		// Close the connection and release the node source
		n.SetSyncState(msgCommon.INACTIVITY)
		n.CloseSync()
	}

	log.Debug("handle version version.pk is ", version.PK)
	if version.P.Cap[msg.HTTP_INFO_FLAG] == 0x01 {
		peer.SetHttpInfoState(true)
	} else {
		peer.SetHttpInfoState(false)
	}
	peer.SetHttpInfoPort(version.P.HttpInfoPort)
	peer.SetConsPort(version.P.ConsensusPort)
	peer.SetBookKeeperAddr(version.PK)

	// if  version.P.Port == version.P.ConsensusPort don't updateInfo
	peer.UpdateInfo(time.Now(), version.P.Version, version.P.Services,
		version.P.Port, version.P.Nonce, version.P.Relay, version.P.StartHeight)

	p2p.Self.Np.AddNbrNode(&peer)

	var buf []byte
	if s == msgCommon.INIT {
		peer.SetSyncState(msgCommon.HANDSHAKE)
		height, _ := actor.GetCurrentBlockHeight()
		vpl := NewVersionPayload(p2p.Self.GetVersion(), p2p.Self.GetServices(),p2p.Self.GetSyncPort(),
			p2p.Self.GetConsPort(), p2p.Self.GetID(), uint64(height), peer.GetRelay(), false)
		buf, _ = NewVersion(vpl, p2p.Self.GetPubKey())
	} else if s == msgCommon.HAND {
		peer.SetSyncState(msgCommon.HANDSHAKED)
		buf, _ = NewVerAck(false)
	}
	peer.SendToSync(buf)

	return nil
}

func VerAckHandle(verAck msg.VerACK, peer peer.Peer, p2p P2PServer) error {
	log.Debug()

	if verAck.IsConsensus == true {
		s := p2p.Self.GetConsState()
		if s != msgCommon.HANDSHAKE && s != msgCommon.HANDSHAKED {
			log.Warn("Unknown status to received verAck")
			return errors.New("Unknown status to received verAck")
		}

		n := p2p.Self.Np.GetPeer(peer.GetID())
		if n == nil {
			log.Warn("nbr node is not exist")
			return errors.New("nbr node is not exist ")
		}

		peer.SetConsState(msgCommon.ESTABLISH)
		n.SetConsState(peer.GetConsState())
		n.SetConsConn(peer.GetConsConn())

		if s == msgCommon.HANDSHAKE {
			buf, _ := NewVerAck(true)
			peer.SendToCons(buf)
		}
		return nil
	}
	s := peer.GetSyncState()
	if s != msgCommon.HANDSHAKE && s != msgCommon.HANDSHAKED {
		log.Warn("Unknown status to received verAck")
		return errors.New("Unknown status to received verAck ")
	}

	peer.SetSyncState(msgCommon.ESTABLISH)

	if s == msgCommon.HANDSHAKE {
		buf, _ := NewVerAck(false)
		peer.SendToSync(buf)
	}

	peer.DumpInfo()
	// Fixme, there is a race condition here,
	// but it doesn't matter to access the invalid
	// node which will trigger a warning
	//TODO JQ: only master p2p port request neighbor list
	buf, _ := NewAddrReq()
	go peer.SendToSync(buf)

	addr := peer.GetAddr()
	port := peer.GetSyncPort()
	nodeAddr := addr + ":" + strconv.Itoa(int(port))
	//TODO JQ： only master p2p port remove the list
	p2p.Self.SyncLink.RemoveAddrInConnectingList(nodeAddr)

	//connect consensus port
	if s == msgCommon.HANDSHAKED {
		consensusPort := peer.GetConsPort()
		nodeConsensusAddr := addr + ":" + strconv.Itoa(int(consensusPort))
		//Fix me:
		go peer.ConsLink.Connect(nodeConsensusAddr, true)
	}
	return nil
}

func AddrHandle(addr msg.Addr, peer peer.Peer, p2p P2PServer) error {
	log.Debug()
	for _, v := range addr.NodeAddrs {
		var ip net.IP
		ip = v.IpAddr[:]
		address := ip.To16().String() + ":" + strconv.Itoa(int(v.Port))
		log.Info(fmt.Sprintf("The ip address is %s id is 0x%x", address, v.ID))

		if v.ID == p2p.Self.GetID() {
			continue
		}

		if p2p.Self.Np.NodeEstablished(v.ID) {
			continue
		}

		if v.Port == 0 {
			continue
		}

		go peer.SyncLink.Connect(address, false)
	}
	return nil
}

func DataReqHandle(dataReq msg.DataReq, peer peer.Peer, p2p P2PServer) error {
	log.Debug()
	reqType := common.InventoryType(dataReq.DataType)
	hash := dataReq.Hash
	switch reqType {
	case common.BLOCK:
		block, err := actor.GetBlockByHash(hash)
		if err != nil {
			log.Debug("Can't get block by hash: ", hash, " ,send not found message")
			b, err := NewNotFound(hash)
			peer.SendToSync(b)
			return err
		}
		log.Debug("block height is ", block.Header.Height, " ,hash is ", hash)
		buf, err := NewBlock(block)
		if err != nil {
			return err
		}
		peer.SendToSync(buf)

	case common.TRANSACTION:
		txn, err := actor.GetTxnFromLedger(hash)
		if err != nil {
			log.Debug("Can't get transaction by hash: ", hash, " ,send not found message")
			b, err := NewNotFound(hash)
			peer.SendToSync(b)
			return err
		}
		buf, err := NewTxn(txn)
		if err != nil {
			return err
		}
		go peer.SendToSync(buf)
	}
	return nil
}

func InvHandle(inv msg.Inv, peer peer.Peer, p2p P2PServer) error {
	log.Debug()
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
			txnDataReq, _ := NewTxnDataReq(id)
			peer.SendToSync(txnDataReq)
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
				blkDataReq, _ := NewBlkDataReq(id)
				peer.SendToSync(blkDataReq)
			}
		}
	case common.CONSENSUS:
		log.Debug("RX consensus message")
		id.Deserialize(bytes.NewReader(inv.P.Blk[:32]))
		consDataReq, _ := NewConsensusDataReq(id)
		peer.SendToCons(consDataReq)
	default:
		log.Warn("RX unknown inventory message")
	}
	return nil
}
