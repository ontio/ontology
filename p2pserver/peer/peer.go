package peer

import (
	"fmt"
	"github.com/Ontology/common/config"
	"github.com/Ontology/common/log"
	//actor "github.com/Ontology/p2pserver/actor/req"
	"github.com/Ontology/crypto"
	"github.com/Ontology/events"
	types "github.com/Ontology/p2pserver/common"
	conn "github.com/Ontology/p2pserver/link"
	//"math/rand"
	"bytes"
	"encoding/binary"
	"errors"
	"net"
	"runtime"
	//"strings"
	"time"
)

type Peer struct {
	link                     *conn.Link
	id                       uint64
	state                    uint32
	version                  uint32
	cap                      [32]byte
	services                 uint64
	relay                    bool
	height                   uint64
	txnCnt                   uint64
	rxTxnCnt                 uint64
	ReceiveChan              chan types.MsgPayload
	publicKey                *crypto.PubKey
	chF                      chan func() error // Channel used to operate the node without lock
	eventQueue                                 // The event queue to notice other modules
	flightHeights            []uint32
	lastContact              time.Time
	nodeDisconnectSubscriber events.Subscriber
	tryCount                 uint32
	//connectingNodes
	//retryConnAddrs
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
	p.link.SetPort(uint16(config.Parameters.NodePort))
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
	p.nodeDisconnectSubscriber = p.eventQueue.GetEvent("disconnect").Subscribe(events.EventNodeDisconnect, p.NodeDisconnect)

	return p, nil
}

func rmPeer(p *Peer) {
	log.Debug(fmt.Sprintf("Remove unused peer: 0x%0x", p.id))
}

func (p *Peer) Version() uint32 {
	return p.version
}
func (p *Peer) GetHeight() uint64 {
	return p.height
}
func (p *Peer) GetState() uint32 {
	return p.state
}
func (p *Peer) GetID() uint64 {
	return p.id
}
func (p *Peer) GetRelay() bool {
	return p.relay
}
func (p *Peer) Services() uint64 {
	return p.services
}
func (p *Peer) GetConnectionState() uint32 {
	return 0
}
func (p *Peer) GetTime() int64 {
	return p.lastContact.UnixNano()
}
func (p *Peer) GetPort() uint16 {
	return p.link.GetPort()
}
func (p *Peer) Send(buf []byte) {
	p.link.Tx(buf)
}
func (p *Peer) GetAddr() string {
	return p.link.GetAddr()
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

func (p *Peer) NodeDisconnect(v interface{}) {
}

func (p *Peer) HandleMsg(buf []byte, len int) error {
	return nil
}
