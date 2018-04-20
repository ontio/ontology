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

package node

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/events"
	"github.com/ontio/ontology/net/actor"
	msg "github.com/ontio/ontology/net/message"
	"github.com/ontio/ontology/net/protocol"
)

type node struct {
	state                    uint32   // node state
	id                       uint64   // The nodes's id
	cap                      [32]byte // The node capability set
	version                  uint32   // The network protocol the node used
	services                 uint64   // The services the node supplied
	relay                    bool     // The relay capability of the node (merge into capbility flag)
	height                   uint64   // The node latest block height
	txnCnt                   uint64   // The transactions be transmit by this node
	rxTxnCnt                 uint64   // The transaction received by this node
	publicKey                keypair.PublicKey
	chF                      chan func() error // Channel used to operate the node without lock
	link                                       // The link status and information
	local                    *node             // The pointer to local node
	nbrNodes                                   // The neighbor node connect with currently node except itself
	eventQueue                                 // The event queue to notice notice other modules
	idCache                                    // The buffer to store the id of the items which already be processed
	flightHeights            []uint32
	lastContact              time.Time
	nodeDisconnectSubscriber events.Subscriber
	tryTimes                 uint32
	ConnectingNodes
	RetryConnAddrs
	blockSync  *BlockSyncMgr
}

type RetryConnAddrs struct {
	sync.RWMutex
	RetryAddrs map[string]int
}

type ConnectingNodes struct {
	sync.RWMutex
	ConnectingAddrs []string
}

func (node *node) DumpInfo() {
	log.Info("Node info:")
	log.Info("\t state = ", node.state)
	log.Info(fmt.Sprintf("\t id = 0x%x", node.id))
	log.Info("\t addr = ", node.addr)
	log.Info("\t conn = ", node.conn)
	log.Info("\t cap = ", node.cap)
	log.Info("\t version = ", node.version)
	log.Info("\t services = ", node.services)
	log.Info("\t port = ", node.port)
	log.Info("\t relay = ", node.relay)
	log.Info("\t height = ", node.height)
	log.Info("\t conn cnt = ", node.link.connCnt)
}

func (node *node) IsAddrInNbrList(addr string) bool {
	node.nbrNodes.RLock()
	defer node.nbrNodes.RUnlock()
	for _, n := range node.nbrNodes.List {
		if n.GetState() == protocol.HAND || n.GetState() == protocol.HAND_SHAKE || n.GetState() == protocol.ESTABLISH {
			addrNew := n.GetAddr()
			port := n.GetPort()
			na := addrNew + ":" + strconv.Itoa(int(port))
			if strings.Compare(na, addr) == 0 {
				return true
			}
		}
	}
	return false
}

func (node *node) SetAddrInConnectingList(addr string) (added bool) {
	node.ConnectingNodes.Lock()
	defer node.ConnectingNodes.Unlock()
	for _, a := range node.ConnectingAddrs {
		if strings.Compare(a, addr) == 0 {
			return false
		}
	}
	node.ConnectingAddrs = append(node.ConnectingAddrs, addr)
	return true
}

func (node *node) RemoveAddrInConnectingList(addr string) {
	node.ConnectingNodes.Lock()
	defer node.ConnectingNodes.Unlock()
	addrs := []string{}
	for i, a := range node.ConnectingAddrs {
		if strings.Compare(a, addr) == 0 {
			addrs = append(node.ConnectingAddrs[:i], node.ConnectingAddrs[i+1:]...)
		}
	}
	node.ConnectingAddrs = addrs
}

func (node *node) UpdateInfo(t time.Time, version uint32, services uint64,
	port uint16, nonce uint64, relay uint8, height uint64) {

	node.UpdateRXTime(t)
	node.id = nonce
	node.version = version
	node.services = services
	node.port = port
	if relay == 0 {
		node.relay = false
	} else {
		node.relay = true
	}
	node.height = uint64(height)
}

func NewNode() *node {
	n := node{
		state: protocol.INIT,
		chF:   make(chan func() error),
	}
	runtime.SetFinalizer(&n, rmNode)
	go n.backend()
	return &n
}

func InitNode(pubKey keypair.PublicKey) protocol.Noder {
	n := NewNode()
	n.version = protocol.PROTOCOL_VERSION
	if config.Parameters.NodeType == protocol.SERVICE_NODE_NAME {
		n.services = uint64(protocol.SERVICE_NODE)
	} else if config.Parameters.NodeType == protocol.VERIFY_NODE_NAME {
		n.services = uint64(protocol.VERIFY_NODE)
	}

	n.link.port = uint16(config.Parameters.NodePort)
	n.relay = true
	rand.Seed(time.Now().UTC().UnixNano())

	key := keypair.SerializePublicKey(pubKey)
	err := binary.Read(bytes.NewBuffer(key[:8]), binary.LittleEndian, &(n.id))
	if err != nil {
		log.Error(err)
	}
	log.Info(fmt.Sprintf("Init node ID to 0x%x", n.id))
	n.nbrNodes.init()
	n.local = n
	n.publicKey = pubKey
	n.eventQueue.init()
	n.nodeDisconnectSubscriber = n.eventQueue.GetEvent("disconnect").Subscribe(events.EventNodeDisconnect, n.NodeDisconnect)
	n.blockSync = NewBlockSyncMgr(n)
	go n.initConnection()
	go n.updateConnection()
	go n.updateNodeInfo()
	go n.blockSync.Start()
	return n
}

func (n *node) NodeDisconnect(v interface{}) {
	if node, ok := v.(*node); ok {
		node.SetState(protocol.INACTIVITY)
		msg.NotifyPeerState(node.GetPubKey(), false)
		conn := node.getConn()
		conn.Close()
	}
}

func rmNode(node *node) {
	log.Debug(fmt.Sprintf("Remove unused/deuplicate node: 0x%0x", node.id))
}

func (node *node) backend() {
	for f := range node.chF {
		f()
	}
}

func (node *node) GetID() uint64 {
	return node.id
}

func (node *node) GetState() uint32 {
	return atomic.LoadUint32(&(node.state))
}

func (node *node) getConn() net.Conn {
	return node.conn
}

func (node *node) GetPort() uint16 {
	return node.port
}

func (node *node) GetHttpInfoPort() int {
	return int(node.httpInfoPort)
}

func (node *node) SetHttpInfoPort(nodeInfoPort uint16) {
	node.httpInfoPort = nodeInfoPort
}

func (node *node) GetHttpInfoState() bool {
	if node.cap[msg.HTTP_INFO_FLAG] == 0x01 {
		return true
	} else {
		return false
	}
}

func (node *node) SetHttpInfoState(nodeInfo bool) {
	if nodeInfo {
		node.cap[msg.HTTP_INFO_FLAG] = 0x01
	} else {
		node.cap[msg.HTTP_INFO_FLAG] = 0x00
	}
}

func (node *node) GetRelay() bool {
	return node.relay
}

func (node *node) Version() uint32 {
	return node.version
}

func (node *node) Services() uint64 {
	return node.services
}

func (node *node) IncRxTxnCnt() {
	node.rxTxnCnt++
}

func (node *node) GetTxnCnt() uint64 {
	return node.txnCnt
}

func (node *node) GetRxTxnCnt() uint64 {
	return node.rxTxnCnt
}

func (node *node) SetState(state uint32) {
	atomic.StoreUint32(&(node.state), state)
}

func (node *node) GetPubKey() keypair.PublicKey {
	return node.publicKey
}

func (node *node) CompareAndSetState(old, new uint32) bool {
	return atomic.CompareAndSwapUint32(&(node.state), old, new)
}

func (node *node) LocalNode() protocol.Noder {
	return node.local
}

func (node *node) GetHeight() uint64 {
	return node.height
}

func (node *node) SetHeight(height uint64) {
	//TODO read/write lock
	node.height = height
}

func (node *node) UpdateRXTime(t time.Time) {
	node.time = t
}

func (node *node) Xmit(message interface{}) error {
	log.Debug()
	var buffer []byte
	var err error
	switch message.(type) {
	case *types.Transaction:
		log.Debug("TX transaction message")
		txn := message.(*types.Transaction)
		buffer, err = msg.NewTxn(txn)
		if err != nil {
			log.Error("Error New Tx message: ", err)
			return err
		}
		node.txnCnt++
	case *types.Block:
		log.Debug("TX block message")
		block := message.(*types.Block)
		buffer, err = msg.NewBlock(block)
		if err != nil {
			log.Error("Error New Block message: ", err)
			return err
		}
	case *msg.ConsensusPayload:
		log.Debug("TX consensus message")
		consensusPayload := message.(*msg.ConsensusPayload)
		buffer, err = msg.NewConsensus(consensusPayload)
		if err != nil {
			log.Error("Error New consensus message: ", err)
			return err
		}
	case common.Uint256:
		log.Debug("TX block hash message")
		hash := message.(common.Uint256)
		buf := bytes.NewBuffer([]byte{})
		hash.Serialize(buf)
		// construct inv message
		invPayload := msg.NewInvPayload(common.BLOCK, 1, buf.Bytes())
		buffer, err = msg.NewInv(invPayload)
		if err != nil {
			log.Error("Error New inv message")
			return err
		}
	default:
		log.Warnf("Unknown Xmit message %v , type %v", message, reflect.TypeOf(message))
		return errors.New("Unknown Xmit message type")
	}
	node.nbrNodes.Broadcast(buffer)

	return nil
}

func (node *node) GetAddr() string {
	return node.addr
}

func (node *node) GetAddr16() ([16]byte, error) {
	var result [16]byte
	ip := net.ParseIP(node.addr).To16()
	if ip == nil {
		log.Error("Parse IP address error\n")
		return result, errors.New("Parse IP address error")
	}

	copy(result[:], ip[:16])
	return result, nil
}

func (node *node) GetTime() int64 {
	t := time.Now()
	return t.UnixNano()
}

func (node *node) GetBookkeeperAddr() keypair.PublicKey {
	return node.publicKey
}

func (node *node) GetBookkeepersAddrs() ([]keypair.PublicKey, uint64) {
	pks := make([]keypair.PublicKey, 1)
	pks[0] = node.publicKey
	var i uint64
	i = 1
	//TODO read lock
	for _, n := range node.nbrNodes.List {
		if n.GetState() == protocol.ESTABLISH && n.services != protocol.SERVICE_NODE {
			pkTmp := n.GetBookkeeperAddr()
			pks = append(pks, pkTmp)
			i++
		}
	}
	return pks, i
}

func (node *node) SetBookkeeperAddr(pk keypair.PublicKey) {
	node.publicKey = pk
}

func compareHeight(blockHeight uint64, heights []uint64) bool {
	for _, height := range heights {
		if blockHeight < height {
			return false
		}
	}
	return true
}

func (node *node) SyncNodeHeight() {
	for {
		heights, _ := node.GetNeighborHeights()
		height, _ := actor.GetCurrentBlockHeight()
		if compareHeight(uint64(height), heights) {
			break
		}
		<-time.After(5 * time.Second)
	}
}

func (node *node) isBlockSyncFinish() bool {
	noders := node.local.GetNeighborNoder()
	if len(noders) == 0 {
		return true
	}

	blockHeight, err := actor.GetCurrentBlockHeight()
	if err != nil {
		log.Errorf("node isBlockSyncFinish GetCurrentBlockHeight error:%s", err)
		return false
	}

	for _, v := range noders {
		if blockHeight < uint32(v.GetHeight()) {
			return false
		}
	}
	return true
}

func (node *node) WaitForSyncBlkFinish() {
	for {
		headerHeight, _ := actor.GetCurrentHeaderHeight()
		currentBlkHeight, _ := actor.GetCurrentBlockHeight()
		log.Info("WaitForSyncBlkFinish... current block height is ", currentBlkHeight, " ,current header height is ", headerHeight)

		if node.isBlockSyncFinish() {
			break
		}
		<-time.After(protocol.PERIOD_UPDATE_TIME * time.Second)
	}
}

func (node *node) WaitForPeersStart() {
	for {
		log.Debug("WaitForPeersStart...")
		if node.IsUptoMinNodeCount() {
			break
		}
		<-time.After(protocol.PERIOD_UPDATE_TIME * time.Second)
	}
}

func (node *node) StoreFlightHeight(height uint32) {
	node.flightHeights = append(node.flightHeights, height)
}

func (node *node) GetFlightHeightCnt() int {
	return len(node.flightHeights)
}
func (node *node) GetFlightHeights() []uint32 {
	return node.flightHeights
}

func (node *node) RemoveFlightHeightLessThan(h uint32) {
	heights := node.flightHeights
	p := len(heights)
	i := 0

	for i < p {
		if heights[i] < h {
			p--
			heights[p], heights[i] = heights[i], heights[p]
		} else {
			i++
		}
	}
	node.flightHeights = heights[:p]
}

func (node *node) RemoveFlightHeight(height uint32) {
	log.Debug("height is ", height)
	for _, h := range node.flightHeights {
		log.Debug("flight height ", h)
	}
	node.flightHeights = sliceRemove(node.flightHeights, height)
	for _, h := range node.flightHeights {
		log.Debug("after flight height ", h)
	}
}

func (node *node) GetLastRXTime() time.Time {
	return node.time
}

func (node *node) AddInRetryList(addr string) {
	node.RetryConnAddrs.Lock()
	defer node.RetryConnAddrs.Unlock()
	if node.RetryAddrs == nil {
		node.RetryAddrs = make(map[string]int)
	}
	if _, ok := node.RetryAddrs[addr]; ok {
		delete(node.RetryAddrs, addr)
		log.Debug("remove exsit addr from retry list", addr)
	}
	//alway set retry to 0
	node.RetryAddrs[addr] = 0
	log.Debug("add addr to retry list", addr)
}

func (node *node) RemoveFromRetryList(addr string) {
	node.RetryConnAddrs.Lock()
	defer node.RetryConnAddrs.Unlock()
	if len(node.RetryAddrs) > 0 {
		if _, ok := node.RetryAddrs[addr]; ok {
			delete(node.RetryAddrs, addr)
			log.Debug("remove addr from retry list", addr)
		}
	}
}


func (node *node) OnAddNode(n protocol.Noder){
	node.AddNbrNode(n)
	node.blockSync.OnAddNode(n.GetID())
}
func (node *node) OnDelNode(nodeId uint64) (protocol.Noder, bool){
	node.blockSync.OnDelNode(nodeId)
	return node.DelNbrNode(nodeId)
}
func (node *node) OnHeaderReceive(headers []*types.Header) {
	node.blockSync.OnHeaderReceive(headers)
}
func (node *node) OnBlockReceive(block *types.Block) {
	node.blockSync.OnBlockReceive(block)
}

func sliceRemove(slice []uint32, h uint32) []uint32 {
	for i, v := range slice {
		if v == h {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}
