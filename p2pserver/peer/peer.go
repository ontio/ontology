package peer

import (
	"fmt"
	"github.com/Ontology/common/config"
	"github.com/Ontology/common/log"
	//actor "github.com/Ontology/p2pserver/actor/req"
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/Ontology/crypto"
	"github.com/Ontology/events"
	types "github.com/Ontology/p2pserver/common"
	conn "github.com/Ontology/p2pserver/link"
	"github.com/Ontology/p2pserver/message"
	"net"
	"runtime"
	"sync/atomic"
	"time"
)

type Peer struct {
	Conn                     *conn.Link
	ConsensusConn            *conn.Link
	id                       uint64
	state                    uint32
	version                  uint32
	cap                      [32]byte
	services                 uint64
	relay                    bool
	height                   uint64
	txnCnt                   uint64
	rxTxnCnt                 uint64
	httpInfoPort             uint16
	publicKey                *crypto.PubKey
	chF                      chan func() error // Channel used to operate the node without lock
	eventQueue                                 // The event queue to notice other modules
	lastContact              time.Time
	peerDisconnectSubscriber events.Subscriber
	notifyFunc               func(v interface{})
	tryCount                 uint32
	Np                       *NbrPeers
}

func (p *Peer) backend() {
	for f := range p.chF {
		f()
	}
}

func NewPeer(pubKey *crypto.PubKey) (*Peer, error) {
	p := &Peer{
		state: types.INIT,
		chF:   make(chan func() error),
	}
	runtime.SetFinalizer(&p, rmPeer)
	go p.backend()

	p.version = types.PROTOCOL_VERSION
	if config.Parameters.NodeType == types.SERVICE_NODE_NAME {
		p.services = uint64(types.SERVICE_NODE)
	} else if config.Parameters.NodeType == types.VERIFY_NODE_NAME {
		p.services = uint64(types.VERIFY_NODE)
	}
	p.Conn.SetPort(config.Parameters.NodePort)
	if config.Parameters.NodeConsensusPort != 0 &&
		config.Parameters.NodeConsensusPort != config.Parameters.NodePort {
		p.ConsensusConn.SetPort(config.Parameters.NodeConsensusPort)
	}

	p.relay = true

	key, err := pubKey.EncodePoint(true)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	err = binary.Read(bytes.NewBuffer(key[:8]), binary.LittleEndian, &(p.id))
	if err != nil {
		log.Error(err)
		return nil, err
	}
	log.Info(fmt.Sprintf("Init peer ID to 0x%x", p.id))
	p.publicKey = pubKey
	p.eventQueue.init()
	p.peerDisconnectSubscriber = p.eventQueue.GetEvent("disconnect").Subscribe(events.EventNodeDisconnect, p.notifyFunc)
	return p, nil
}

func rmPeer(p *Peer) {
	log.Debug(fmt.Sprintf("Remove unused peer: 0x%0x", p.id))
}
func (p *Peer) GetPubKey() *crypto.PubKey {
	return p.publicKey
}
func (p *Peer) GetVersion() uint32 {
	return p.version
}
func (p *Peer) GetHeight() uint64 {
	return p.height
}
func (p *Peer) SetHeight(height uint64) {
	p.height = height
}
func (p *Peer) GetState() uint32 {
	return p.state
}
func (p *Peer) SetState(state uint32) {
	atomic.StoreUint32(&(p.state), state)
}
func (p *Peer) GetID() uint64 {
	return p.id
}
func (p *Peer) GetRelay() bool {
	return p.relay
}
func (p *Peer) GetServices() uint64 {
	return p.services
}
func (p *Peer) GetConnectionState() uint32 {
	return p.state
}
func (p *Peer) GetTime() int64 {
	return p.lastContact.UnixNano()
}
func (p *Peer) GetPort() uint16 {
	return p.Conn.GetPort()
}
func (p *Peer) GetConsensusPort() uint16 {
	return p.ConsensusConn.GetPort()
}
func (p *Peer) GetAddr() string {
	return p.Conn.GetAddr()
}
func (p *Peer) GetAddr16() ([16]byte, error) {
	var result [16]byte
	ip := net.ParseIP(p.GetAddr()).To16()
	if ip == nil {
		log.Error("Parse IP address error\n")
		return result, errors.New("Parse IP address error")
	}

	copy(result[:], ip[:16])
	return result, nil
}
func (p *Peer) Close() {
	p.SetState(types.INACTIVITY)
	conn := p.Conn.GetConn()
	conn.Close()
}
func (p *Peer) AttachChan(msgchan chan types.MsgPayload) {
	p.Conn.SetChan(msgchan)
}

func (p *Peer) AttachEvent(fn func(v interface{})) {
	p.notifyFunc = fn
}

func (p *Peer) DelNbrNode(id uint64) (*Peer, bool) {
	return p.Np.DelNbrNode(id)
}

func (p *Peer) CloseConn() {
	p.Conn.CloseConn()
}

func (p *Peer) CloseConsensusConn() {
	p.ConsensusConn.CloseConn()
}

func (p *Peer) Send(buf []byte, isConsensus bool) {
	if isConsensus && p.ConsensusConn.Valid() {
		p.ConsensusConn.Tx(buf)
	}
	p.Conn.Tx(buf)
}

func (p *Peer) SetHttpInfoState(httpInfo bool) {
	if httpInfo {
		p.cap[message.HTTP_INFO_FLAG] = 0x01
	} else {
		p.cap[message.HTTP_INFO_FLAG] = 0x00
	}
}

func (p *Peer) GetHttpInfoPort() uint16 {
	return p.httpInfoPort
}

func (p *Peer) SetHttpInfoPort(port uint16) {
	p.httpInfoPort = port
}

//SetBookkeeperAddr set peer`s publickey
func (p *Peer) SetBookkeeperAddr(pk *crypto.PubKey) {
	p.publicKey = pk
}

//UpdateInfo update peer`s information
func (p *Peer) UpdateInfo(t time.Time, version uint32, services uint64,
	port uint16, nonce uint64, relay uint8, height uint64) {

	p.Conn.UpdateRXTime(t)
	p.id = nonce
	p.version = version
	p.services = services
	p.Conn.SetPort(port)
	if relay == 0 {
		p.relay = false
	} else {
		p.relay = true
	}
	p.height = uint64(height)
}

//AddNbrNode add peer to nbr peer list
func (p *Peer) AddNbrNode(remotePeer *Peer) {
	p.Np.AddNbrNode(remotePeer)
}

//StartListen init link layer for listenning
func (p *Peer) StartListen() {
	p.Conn.InitConnection()
	if p.ConsensusConn.GetPort() != 0 {
		p.ConsensusConn.InitConnection()
	}

}

//Dump print a peer`s information
func (p *Peer) Dump() {
	log.Info("Peer info:")
	log.Info("\t state = ", p.state)
	log.Info(fmt.Sprintf("\t id = 0x%x", p.id))
	log.Info("\t addr = ", p.Conn.GetAddr())
	log.Info("\t cap = ", p.cap)
	log.Info("\t version = ", p.version)
	log.Info("\t services = ", p.services)
	log.Info("\t port = ", p.Conn.GetPort())
	log.Info("\t consensus port = ", p.ConsensusConn.GetPort())
	log.Info("\t relay = ", p.relay)
	log.Info("\t height = ", p.height)
	log.Info("\t conn cnt = ", p.Np.GetConnectionCnt())
}
