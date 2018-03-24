package peer

import (
	"fmt"
	"github.com/Ontology/common/config"
	"github.com/Ontology/common/log"
	//actor "github.com/Ontology/p2pserver/actor/req"
	types "github.com/Ontology/p2pserver/common"
	//"github.com/Ontology/p2pserver/link"
	"github.com/Ontology/crypto"
	"github.com/Ontology/events"
	//"math/rand"
	//"net"
	//"strconv"
	//"runtime"
	"strings"
	"time"
)

type Peer struct {
	//link                     *link
	id                       uint32
	state                    uint32
	consensusState           uint32
	version                  uint32
	cap                      []uint32
	height                   uint64 // The peer latest block height
	publicKey                *crypto.PubKey
	chF                      chan func() error // Channel used to operate the node without lock
	eventQueue                                 // The event queue to notice notice other modules
	flightHeights            []uint32
	lastContact              time.Time
	nodeDisconnectSubscriber events.Subscriber
	tryCount                 uint32
	//connectingNodes
	//retryConnAddrs
	//actors
	// poolActor   *ns.TxPoolActor
	// ledgerActor *ns.LedgerActor
	// conActor    *ns.ConsensusActor
}

func NewPeer(pubKey *crypto.PubKey) *Peer {

	return nil
}

func rmNode(p *Peer) {
	log.Debug(fmt.Sprintf("Remove unused peer: 0x%0x", p.id))
}

func (p *Peer) Start(bool, bool) error {
	return nil
}
func (p *Peer) Stop() error {
	return nil
}
func (p *Peer) GetVersion() uint32 {
	return 0
}
func (p *Peer) GetConnectionCnt() uint {
	return 0
}
func (p *Peer) GetPort() (uint16, uint16) {
	return 0, 0
}
func (p *Peer) GetState() uint32 {
	return 0
}
func (p *Peer) GetId() uint64 {
	return 0
}
func (p *Peer) Services() uint64 {
	return 0
}
func (p *Peer) GetConnectionState() uint32 {
	return 0
}
func (p *Peer) GetTime() int64 {
	return 0
}
func (p *Peer) GetNeighborAddrs() ([]types.PeerAddr, uint64) {
	return nil, 0
}
func (p *Peer) Xmit(interface{}) error {
	return nil
}
func (p *Peer) IsSyncing() bool {
	return false
}
func (p *Peer) IsStarted() bool {
	return false
}
func (p *Peer) EnableDual(bool) error {
	return nil
}
func (p *Peer) isUptoMinCount() bool {
	consensusType := strings.ToLower(config.Parameters.ConsensusType)
	if consensusType == "" {
		consensusType = "dbft"
	}
	minCount := config.DBFTMINNODENUM
	switch consensusType {
	case "dbft":
	case "solo":
		minCount = config.SOLOMINNODENUM
	}
	return int(p.GetNbrCnt())+1 >= minCount
}
func (p *Peer) GetNbrCnt() uint32 {
	return 0
}
func (p *Peer) HandleMsg(buf []byte, len int) error {
	return nil
}
