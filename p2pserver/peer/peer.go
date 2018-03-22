package peer

import (
	"fmt"
	"github.com/Ontology/common/config"
	"github.com/Ontology/common/log"
	//actor "github.com/Ontology/p2pserver/actor/req"
	//. "github.com/Ontology/p2pserver/message"
	//"github.com/Ontology/p2pserver/link"
	"github.com/Ontology/crypto"
	"github.com/Ontology/events"
	. "github.com/Ontology/p2pserver/protocol"
	//"math/rand"
	//"net"
	//"strconv"
	//"runtime"
	"strings"
	"time"
)

type peer struct {
	//link          *link
	//consensusLink *link
	id             uint32
	state          uint32
	consensusState uint32
	version        uint32
	cap            []uint32
	this           *peer  // The pointer to local peer
	height         uint64 // The peer latest block height
	publicKey      *crypto.PubKey
	chF            chan func() error // Channel used to operate the node without lock
	//nbrPeers                 nbrNodes          // The neighbor peer connect with currently node except itself
	eventQueue               // The event queue to notice notice other modules
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

func (p *peer) getHdrs() {
	// if !p.isUptoMinCount() {
	// 	return
	// }
	// peers := p.this.GetNeighborNoder()
	// if len(peers) == 0 {
	// 	return
	// }
	// plist := []peer{}
	// for _, v := range peers {
	// 	height, _ := actor.GetCurrentHeaderHeight()
	// 	if uint64(height) < v.GetHeight() {
	// 		plist = append(plist, v)
	// 	}
	// }
	// count := len(plist)
	// if count == 0 {
	// 	return
	// }
	// rand.Seed(time.Now().UnixNano())
	// n := plist[rand.Intn(count)]
	//sendSyncHeaders(n)
}
func NewPeer(pubKey *crypto.PubKey) *peer {

	return nil
}

func rmNode(peer *peer) {
	log.Debug(fmt.Sprintf("Remove unused peer: 0x%0x", peer.id))
}

func (p *peer) Start(bool, bool) error {
	return nil
}
func (p *peer) Stop() error {
	return nil
}
func (p *peer) GetVersion() uint32 {
	return 0
}
func (p *peer) GetConnectionCnt() uint64 {
	return 0
}
func (p *peer) GetPort() (uint16, uint16) {
	return 0, 0
}
func (p *peer) GetState() uint32 {
	return 0
}
func (p *peer) GetId() uint64 {
	return 0
}
func (p *peer) Services() uint64 {
	return 0
}
func (p *peer) GetConnectionState() uint32 {
	return 0
}
func (p *peer) GetTime() int64 {
	return 0
}
func (p *peer) GetNeighborAddrs() ([]PeerAddr, uint64) {
	return nil, 0
}
func (p *peer) Xmit(interface{}) error {
	return nil
}
func (p *peer) IsSyncing() bool {
	return false
}
func (p *peer) IsStarted() bool {
	return false
}
func (p *peer) EnableDual(bool) error {
	return nil
}
func (p *peer) isUptoMinCount() bool {
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
func (p *peer) GetNbrCnt() uint32 {
	return 0
}
func (p *peer) HandleMsg(buf []byte, len int) error {
	return nil
}
