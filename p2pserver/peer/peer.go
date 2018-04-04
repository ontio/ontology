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
	LinkConn                 *conn.Link
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

	p.version = types.PROTOCOLVERSION
	if config.Parameters.NodeType == types.SERVICENODENAME {
		p.services = uint64(types.SERVICENODE)
	} else if config.Parameters.NodeType == types.VERIFYNODENAME {
		p.services = uint64(types.VERIFYNODE)
	}
	p.LinkConn.SetPort(uint16(config.Parameters.NodePort))
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
	return 0
}
func (p *Peer) GetTime() int64 {
	return p.lastContact.UnixNano()
}
func (p *Peer) GetPort() uint16 {
	return p.LinkConn.GetPort()
}
func (p *Peer) GetConsensusPort() uint16 {
	return p.LinkConn.GetConsensusPort()
}
func (p *Peer) Send(buf []byte) {
	p.LinkConn.Tx(buf)
}
func (p *Peer) GetAddr() string {
	return p.LinkConn.GetAddr()
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
	conn := p.LinkConn.GetConn()
	conn.Close()
}
func (p *Peer) AttachChan(msgchan chan types.MsgPayload) {
	p.LinkConn.SetChan(msgchan)
}

func (p *Peer) AttachEvent(fn func(v interface{})) {
	p.notifyFunc = fn
}

func (p *Peer) DelNbrNode(id uint64) (*Peer, bool) {
	return p.Np.DelNbrNode(id)
}

func (p *Peer) CloseConn() {
	p.LinkConn.CloseConn()
}

func (p *Peer) Tx(buf []byte) {
	p.LinkConn.Tx(buf)
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

func (p *Peer) SetBookkeeperAddr(pk *crypto.PubKey) {
	p.publicKey = pk
}

func (p *Peer) UpdateInfo(t time.Time, version uint32, services uint64,
	port uint16, nonce uint64, relay uint8, height uint64) {

	p.LinkConn.UpdateRXTime(t)
	p.id = nonce
	p.version = version
	p.services = services
	p.LinkConn.SetPort(port)
	if relay == 0 {
		p.relay = false
	} else {
		p.relay = true
	}
	p.height = uint64(height)
}

func (p *Peer) AddNbrNode(remotePeer *Peer) {
	p.Np.AddNbrNode(remotePeer)
}
