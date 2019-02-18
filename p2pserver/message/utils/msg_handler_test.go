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
	"encoding/binary"
	"encoding/hex"
	"testing"
	"time"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/genesis"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/payload"
	ct "github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/events"
	msgCommon "github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/message/msg_pack"
	"github.com/ontio/ontology/p2pserver/message/types"
	"github.com/ontio/ontology/p2pserver/net/netserver"
	"github.com/ontio/ontology/p2pserver/net/protocol"
	"github.com/ontio/ontology/p2pserver/peer"
	"github.com/stretchr/testify/assert"
)

var (
	network p2p.P2P
)

func init() {
	log.Init(log.Stdout)
	// Start local network server and create message router
	network = netserver.NewNetServer()

	events.Init()
	// Initial a ledger
	var err error
	ledger.DefLedger, err = ledger.NewLedger(config.DEFAULT_DATA_DIR)
	if err != nil {
		log.Fatalf("NewLedger error %s", err)
	}

	bookKeepers, err := config.DefConfig.GetBookkeepers()
	if err != nil {
		log.Fatal("failed to get bookkeepers")
		return
	}
	genesisConfig := config.DefConfig.Genesis
	genesisBlock, err := genesis.BuildGenesisBlock(bookKeepers, genesisConfig)
	if err != nil {
		log.Fatal("failed to build genesis block", err)
		return

	}
	err = ledger.DefLedger.Init(bookKeepers, genesisBlock)
	if err != nil {
		log.Fatalf("DefLedger.Init error %s", err)
	}
}

// TestVersionHandle tests Function VersionHandle handling a version message
func TestVersionHandle(t *testing.T) {
	// Simulate a remote peer to connect to the local
	remotePeer := peer.NewPeer()
	assert.NotNil(t, remotePeer)

	network.AddPeerSyncAddress("127.0.0.1:50010", remotePeer)

	var testID uint64
	_, testPub, _ := keypair.GenerateKeyPair(keypair.PK_ECDSA, keypair.P256)
	key := keypair.SerializePublicKey(testPub)
	err := binary.Read(bytes.NewBuffer(key[:8]), binary.LittleEndian, &(testID))
	assert.Nil(t, err)

	// Construct a version packet
	buf := msgpack.NewVersion(network, false, 12345)
	version := buf.(*types.Version)
	version.P.Nonce = testID

	msg := &types.MsgPayload{
		Id:      testID,
		Addr:    "127.0.0.1:50010",
		Payload: buf,
	}

	// Invoke VersionHandle to handle the msg
	VersionHandle(msg, network, nil)

	// Get the remote peer from the neighbor peers by peer id
	tempPeer := network.GetPeer(testID)
	assert.NotNil(t, tempPeer)

	assert.Equal(t, tempPeer.GetID(), testID)
	assert.Equal(t, tempPeer.GetVersion(), network.GetVersion())
	assert.Equal(t, tempPeer.GetServices(), network.GetServices())
	assert.Equal(t, tempPeer.GetSyncPort(), network.GetSyncPort())
	assert.Equal(t, tempPeer.GetHttpInfoPort(), network.GetHttpInfoPort())
	assert.Equal(t, tempPeer.GetConsPort(), network.GetConsPort())
	assert.Equal(t, tempPeer.GetHeight(), uint64(12345))
	assert.Equal(t, tempPeer.GetSyncState(), uint32(msgCommon.HAND_SHAKE))

	network.DelNbrNode(testID)
}

// TestVerAckHandle tests Function VerAckHandle handling a version ack
func TestVerAckHandle(t *testing.T) {
	// Simulate a remote peer to be added to the neighbor peers
	var testID uint64
	_, testPub, _ := keypair.GenerateKeyPair(keypair.PK_ECDSA, keypair.P256)
	key := keypair.SerializePublicKey(testPub)
	err := binary.Read(bytes.NewBuffer(key[:8]), binary.LittleEndian, &(testID))
	assert.Nil(t, err)

	remotePeer := peer.NewPeer()
	assert.NotNil(t, remotePeer)

	remotePeer.SetHttpInfoPort(20335)
	remotePeer.UpdateInfo(time.Now(), 1, 12345678, 20336,
		20337, testID, 0, 12345)
	network.AddNbrNode(remotePeer)
	remotePeer.SetSyncState(msgCommon.HAND_SHAKE)

	// Construct a version ack packet
	buf := msgpack.NewVerAck(false)

	msg := &types.MsgPayload{
		Id:      testID,
		Addr:    "127.0.0.1:50010",
		Payload: buf,
	}

	// Invoke VerAckHandle to handle the msg
	VerAckHandle(msg, network, nil)

	// Get the remote peer from the neighbor peers by peer id
	tempPeer := network.GetPeer(testID)
	assert.NotNil(t, tempPeer)
	assert.Equal(t, tempPeer.GetSyncState(), uint32(msgCommon.ESTABLISH))

	network.DelNbrNode(testID)
}

// TestAddrReqHandle tests Function AddrReqHandle handling an address req
func TestAddrReqHandle(t *testing.T) {
	// Simulate a remote peer to be added to the neighbor peers
	var testID uint64
	_, testPub, _ := keypair.GenerateKeyPair(keypair.PK_ECDSA, keypair.P256)
	key := keypair.SerializePublicKey(testPub)
	err := binary.Read(bytes.NewBuffer(key[:8]), binary.LittleEndian, &(testID))
	assert.Nil(t, err)

	remotePeer := peer.NewPeer()
	assert.NotNil(t, remotePeer)

	remotePeer.UpdateInfo(time.Now(), 1, 12345678, 20336,
		20337, testID, 0, 12345)
	remotePeer.SyncLink.SetAddr("127.0.0.1:50010")

	network.AddNbrNode(remotePeer)

	// Construct an address request packet
	buf := msgpack.NewAddrReq()

	msg := &types.MsgPayload{
		Id:      testID,
		Addr:    "127.0.0.1:50010",
		Payload: buf,
	}

	// Invoke AddrReqHandle to handle the msg
	AddrReqHandle(msg, network, nil)

	network.DelNbrNode(testID)
}

// TestHeadersReqHandle tests Function HeadersReqHandle handling a header req
func TestHeadersReqHandle(t *testing.T) {
	// Simulate a remote peer to be added to the neighbor peers
	var testID uint64
	_, testPub, _ := keypair.GenerateKeyPair(keypair.PK_ECDSA, keypair.P256)
	key := keypair.SerializePublicKey(testPub)
	err := binary.Read(bytes.NewBuffer(key[:8]), binary.LittleEndian, &(testID))
	assert.Nil(t, err)

	remotePeer := peer.NewPeer()
	assert.NotNil(t, remotePeer)

	remotePeer.UpdateInfo(time.Now(), 1, 12345678, 20336,
		20337, testID, 0, 12345)
	remotePeer.SyncLink.SetAddr("127.0.0.1:50010")

	network.AddNbrNode(remotePeer)

	// Construct a headers request of packet
	headerHash := ledger.DefLedger.GetCurrentHeaderHash()
	buf := msgpack.NewHeadersReq(headerHash)

	msg := &types.MsgPayload{
		Id:      testID,
		Addr:    "127.0.0.1:50010",
		Payload: buf,
	}

	// Invoke HeadersReqhandle to handle the msg
	HeadersReqHandle(msg, network, nil)
	network.DelNbrNode(testID)
}

// TestPingHandle tests Function PingHandle handling a ping message
func TestPingHandle(t *testing.T) {
	// Simulate a remote peer to be added to the neighbor peers
	var testID uint64
	_, testPub, _ := keypair.GenerateKeyPair(keypair.PK_ECDSA, keypair.P256)
	key := keypair.SerializePublicKey(testPub)
	err := binary.Read(bytes.NewBuffer(key[:8]), binary.LittleEndian, &(testID))
	assert.Nil(t, err)

	remotePeer := peer.NewPeer()
	assert.NotNil(t, remotePeer)
	remotePeer.UpdateInfo(time.Now(), 1, 12345678, 20336,
		20337, testID, 0, 12345)
	remotePeer.SyncLink.SetAddr("127.0.0.1:50010")

	network.AddNbrNode(remotePeer)

	// Construct a ping packet
	height := ledger.DefLedger.GetCurrentBlockHeight()
	assert.Nil(t, err)

	buf := msgpack.NewPingMsg(uint64(height))

	msg := &types.MsgPayload{
		Id:      testID,
		Addr:    "127.0.0.1:50010",
		Payload: buf,
	}

	// Invoke PingHandle to handle the msg
	PingHandle(msg, network, nil)

	network.DelNbrNode(testID)
}

// TestPingHandle tests Function PingHandle handling a pong message
func TestPongHandle(t *testing.T) {
	// Simulate a remote peer to be added to the neighbor peers
	var testID uint64
	_, testPub, _ := keypair.GenerateKeyPair(keypair.PK_ECDSA, keypair.P256)
	key := keypair.SerializePublicKey(testPub)
	err := binary.Read(bytes.NewBuffer(key[:8]), binary.LittleEndian, &(testID))
	assert.Nil(t, err)

	remotePeer := peer.NewPeer()
	assert.NotNil(t, remotePeer)
	remotePeer.UpdateInfo(time.Now(), 1, 12345678, 20336,
		20337, testID, 0, 12345)
	remotePeer.SyncLink.SetAddr("127.0.0.1:50010")

	network.AddNbrNode(remotePeer)

	// Construct a pong packet
	height := ledger.DefLedger.GetCurrentBlockHeight()
	assert.Nil(t, err)

	buf := msgpack.NewPongMsg(uint64(height))

	msg := &types.MsgPayload{
		Id:      testID,
		Addr:    "127.0.0.1:50010",
		Payload: buf,
	}

	// Invoke PingHandle to handle the msg
	PongHandle(msg, network, nil)

	network.DelNbrNode(testID)
}

// TestBlkHeaderHandle tests Function BlkHeaderHandle handling a sync header msg
func TestBlkHeaderHandle(t *testing.T) {
	// Simulate a remote peer to be added to the neighbor peers
	var testID uint64
	_, testPub, _ := keypair.GenerateKeyPair(keypair.PK_ECDSA, keypair.P256)
	key := keypair.SerializePublicKey(testPub)
	err := binary.Read(bytes.NewBuffer(key[:8]), binary.LittleEndian, &(testID))
	assert.Nil(t, err)

	remotePeer := peer.NewPeer()
	assert.NotNil(t, remotePeer)
	remotePeer.UpdateInfo(time.Now(), 1, 12345678, 20336,
		20337, testID, 0, 12345)
	remotePeer.SyncLink.SetAddr("127.0.0.1:50010")

	network.AddNbrNode(remotePeer)

	// Construct a sync header packet
	hash := ledger.DefLedger.GetBlockHash(0)
	assert.NotEqual(t, hash, common.UINT256_EMPTY)

	headers, err := GetHeadersFromHash(hash, hash)
	assert.Nil(t, err)

	buf := msgpack.NewHeaders(headers)

	msg := &types.MsgPayload{
		Id:      testID,
		Addr:    "127.0.0.1:50010",
		Payload: buf,
	}

	// Invoke BlkHeaderHandle to handle the msg
	BlkHeaderHandle(msg, network, nil)

	network.DelNbrNode(testID)
}

// TestBlockHandle tests Function BlockHandle handling a block message
func TestBlockHandle(t *testing.T) {
	// Simulate a remote peer to be added to the neighbor peers
	var testID uint64
	_, testPub, _ := keypair.GenerateKeyPair(keypair.PK_ECDSA, keypair.P256)
	key := keypair.SerializePublicKey(testPub)
	err := binary.Read(bytes.NewBuffer(key[:8]), binary.LittleEndian, &(testID))
	assert.Nil(t, err)

	remotePeer := peer.NewPeer()
	assert.NotNil(t, remotePeer)
	remotePeer.UpdateInfo(time.Now(), 1, 12345678, 20336,
		20337, testID, 0, 12345)
	remotePeer.SyncLink.SetAddr("127.0.0.1:50010")

	network.AddNbrNode(remotePeer)

	// Construct a block packet
	hash := ledger.DefLedger.GetBlockHash(0)
	assert.NotEqual(t, hash, common.UINT256_EMPTY)

	block, err := ledger.DefLedger.GetBlockByHash(hash)
	assert.Nil(t, err)

	mr, err := common.Uint256FromHexString("1b8fa7f242d0eeb4395f89cbb59e4c29634047e33245c4914306e78a88e14ce5")
	assert.Nil(t, err)
	buf := msgpack.NewBlock(block, mr)

	msg := &types.MsgPayload{
		Id:      testID,
		Addr:    "127.0.0.1:50010",
		Payload: buf,
	}

	// Invoke BlockHandle to handle the msg
	BlockHandle(msg, network, nil)

	network.DelNbrNode(testID)
}

// TestConsensusHandle tests Function ConsensusHandle handling a consensus message
func TestConsensusHandle(t *testing.T) {
	var testID uint64
	_, testPub, _ := keypair.GenerateKeyPair(keypair.PK_ECDSA, keypair.P256)
	key := keypair.SerializePublicKey(testPub)
	err := binary.Read(bytes.NewBuffer(key[:8]), binary.LittleEndian, &(testID))
	assert.Nil(t, err)

	hash := ledger.DefLedger.GetBlockHash(0)
	assert.NotEqual(t, hash, common.UINT256_EMPTY)

	cpl := &types.ConsensusPayload{
		Version:         1,
		PrevHash:        hash,
		Height:          0,
		BookkeeperIndex: 0,
		Timestamp:       uint32(time.Now().UTC().UnixNano()),
		Data:            []byte{},
		Owner:           testPub,
		Signature:       []byte{},
	}

	buf := msgpack.NewConsensus(cpl)

	msg := &types.MsgPayload{
		Id:      testID,
		Addr:    "127.0.0.1:50010",
		Payload: buf,
	}

	ConsensusHandle(msg, network, nil)
}

// TestNotFoundHandle tests Function NotFoundHandle handling a not found message
func TestNotFoundHandle(t *testing.T) {
	tempStr := "3369930accc1ddd067245e8edadcd9bea207ba5e1753ac18a51df77a343bfe92"
	hex, _ := hex.DecodeString(tempStr)
	var hash common.Uint256
	hash.Deserialize(bytes.NewReader(hex))

	buf := msgpack.NewNotFound(hash)

	msg := &types.MsgPayload{
		Id:      0,
		Addr:    "127.0.0.1:50010",
		Payload: buf,
	}

	NotFoundHandle(msg, network, nil)
}

// TestTransactionHandle tests Function TransactionHandle handling a transaction message
func TestTransactionHandle(t *testing.T) {
	code := []byte("ont")
	invokeCodePayload := &payload.InvokeCode{
		Code: code,
	}
	tx := &ct.Transaction{
		Version: 0,
		TxType:  ct.Invoke,
		Payload: invokeCodePayload,
	}

	buf := msgpack.NewTxn(tx)

	msg := &types.MsgPayload{
		Id:      0,
		Addr:    "127.0.0.1:50010",
		Payload: buf,
	}

	TransactionHandle(msg, network, nil)
}

// TestAddrHandle tests Function AddrHandle handling a neighbor address response message
func TestAddrHandle(t *testing.T) {
	nodeAddrs := []msgCommon.PeerAddr{}
	buf := msgpack.NewAddrs(nodeAddrs)
	msg := &types.MsgPayload{
		Id:      0,
		Addr:    "127.0.0.1:50010",
		Payload: buf,
	}

	AddrHandle(msg, network, nil)
}

// TestDataReqHandle tests Function DataReqHandle handling a data req(block/Transaction)
func TestDataReqHandle(t *testing.T) {
	var testID uint64
	_, testPub, _ := keypair.GenerateKeyPair(keypair.PK_ECDSA, keypair.P256)
	key := keypair.SerializePublicKey(testPub)
	err := binary.Read(bytes.NewBuffer(key[:8]), binary.LittleEndian, &(testID))
	assert.Nil(t, err)

	remotePeer := peer.NewPeer()
	assert.NotNil(t, remotePeer)
	remotePeer.UpdateInfo(time.Now(), 1, 12345678, 20336,
		20337, testID, 0, 12345)
	remotePeer.SyncLink.SetAddr("127.0.0.1:50010")

	network.AddNbrNode(remotePeer)

	hash := ledger.DefLedger.GetBlockHash(0)
	assert.NotEqual(t, hash, common.UINT256_EMPTY)
	buf := msgpack.NewBlkDataReq(hash)

	msg := &types.MsgPayload{
		Id:      testID,
		Addr:    "127.0.0.1:50010",
		Payload: buf,
	}

	DataReqHandle(msg, network, nil)

	tempStr := "3369930accc1ddd067245e8edadcd9bea207ba5e1753ac18a51df77a343bfe92"
	hex, _ := hex.DecodeString(tempStr)
	var txHash common.Uint256
	txHash.Deserialize(bytes.NewReader(hex))
	buf = msgpack.NewTxnDataReq(txHash)
	msg = &types.MsgPayload{
		Id:      testID,
		Addr:    "127.0.0.1:50010",
		Payload: buf,
	}

	DataReqHandle(msg, network, nil)

	network.DelNbrNode(testID)
}

// TestInvHandle tests Function InvHandle handling an inventory message
func TestInvHandle(t *testing.T) {
	var testID uint64
	_, testPub, _ := keypair.GenerateKeyPair(keypair.PK_ECDSA, keypair.P256)
	key := keypair.SerializePublicKey(testPub)
	err := binary.Read(bytes.NewBuffer(key[:8]), binary.LittleEndian, &(testID))
	assert.Nil(t, err)

	remotePeer := peer.NewPeer()
	assert.NotNil(t, remotePeer)
	remotePeer.UpdateInfo(time.Now(), 1, 12345678, 20336,
		20337, testID, 0, 12345)
	remotePeer.SyncLink.SetAddr("127.0.0.1:50010")

	network.AddNbrNode(remotePeer)

	hash := ledger.DefLedger.GetBlockHash(0)
	assert.NotEqual(t, hash, common.UINT256_EMPTY)

	buf := bytes.NewBuffer([]byte{})
	hash.Serialize(buf)
	invPayload := msgpack.NewInvPayload(common.BLOCK, []common.Uint256{hash})
	buffer := msgpack.NewInv(invPayload)
	msg := &types.MsgPayload{
		Id:      testID,
		Addr:    "127.0.0.1:50010",
		Payload: buffer,
	}

	InvHandle(msg, network, nil)

	network.DelNbrNode(testID)
}

// TestDisconnectHandle tests Function DisconnectHandle handling a disconnect event
func TestDisconnectHandle(t *testing.T) {
	var testID uint64
	_, testPub, _ := keypair.GenerateKeyPair(keypair.PK_ECDSA, keypair.P256)
	key := keypair.SerializePublicKey(testPub)
	err := binary.Read(bytes.NewBuffer(key[:8]), binary.LittleEndian, &(testID))
	assert.Nil(t, err)

	remotePeer := peer.NewPeer()
	assert.NotNil(t, remotePeer)
	remotePeer.UpdateInfo(time.Now(), 1, 12345678, 20336,
		20337, testID, 0, 12345)
	remotePeer.SyncLink.SetAddr("127.0.0.1:50010")

	network.AddNbrNode(remotePeer)

	msgDisc, err := types.MakeEmptyMessage(msgCommon.DISCONNECT_TYPE)
	msg := &types.MsgPayload{
		Id:      testID,
		Addr:    "127.0.0.1:50010",
		Payload: msgDisc,
	}

	DisconnectHandle(msg, network, nil)

	network.DelNbrNode(testID)
}
