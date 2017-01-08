package net

import (
	"log"
	"time"
	"os"
	"bytes"
	"encoding/hex"
	"GoOnchain/common"
	"GoOnchain/config"
)

var cnt int = 0

const (
	HELLOTIMEOUT	 = 3	// Seconds
	MAXHELLORETYR	 = 3
	MAXBUFLEN	 = 1024 * 300 // Fixme The maximum buffer to receive message
	MAXCHANBUF	 = 512
	NETMAGIC	 = 0x74746e41 // Keep the same as antshares only for testing
	//NETMAGIC	 = 0x414d5446 // Keep the same as antshares only for testing
	PROTOCOLVERSION  = 0

	NODETESTPORT	 = 20333	// TODO get from config file
	PERIODUPDATETIME = 3		// Time to update and sync information with other nodes
)

// The unconfirmed transaction queue
var UnconfTrsCh = make(chan *Msg, MAXCHANBUF)
// Channel used to commnucate with ledger module
var NetToLedgerCh = make(chan *Msg, MAXCHANBUF)
// Channel used to communicate with Consensus module
var NetToConsensusCh = make(chan *Msg, MAXCHANBUF)
// Copntrol channel to send some module control command
var NetToLedgerCtlCh = make(chan *Msg, MAXCHANBUF)
// Channel used to commnucate with ledger module
var LedgerToNetCh = make(chan *Msg, MAXCHANBUF)
// Channel used to communicate with Consensus module
var ConsensusToNetCh = make(chan *Msg, MAXCHANBUF)
// Copntrol channel to send some module control command
var LedgerToNetCtlCh = make(chan string, MAXCHANBUF)

// TODO update to slice for better function calling speed
var funcMap = struct {
	handles map[string] func(*node, *Msg)
} {handles: map[string] func(*node, *Msg) {
	"version":	rxVersion,
	"verack":	rxVerack,
	"getaddr":	rxGetaddr,
	"addr":		rxAddr,
	"getheaders":	rxGetHeaders,
	"headers":	rxHeaders,
	"getblocks":	rxGetBlocks,
	"inv":		rxInv,
	"getdata":	rxGetData,
	"block":	rxBlock,
	"tx":		rxTransaction,
}}

func Init() {
}

func rxLedgerMsg(msg *Msg) {
	common.Trace()
}

func rxLedgerCtlMsg(msg string) {
	common.Trace()
}

func rxConsensusMsg(msg *Msg) {
	common.Trace()
}

func handleModuleMsg() {
	select {
	case ledgeMsg := <- LedgerToNetCh:
		rxLedgerMsg(ledgeMsg)
		break
	case consMsg := <- ConsensusToNetCh:
		rxConsensusMsg(consMsg)
		break
	case ledgerCtlMsg := <- LedgerToNetCtlCh:
		rxLedgerCtlMsg(ledgerCtlMsg)
		break
	default:
		log.Println("Unknown message received by net module")
	}
}

/*
 * The node state switch table after rx message, there is time limitation for each action
 * The Hanshark status will switch to INIT after TIMEOUT if not received the VerACK
 * in this time window
 *  _______________________________________________________________________
 * |          |    INIT         | HANDSHAKE |  ESTABLISH | INACTIVITY      |
 * |-----------------------------------------------------------------------|
 * | version  | HANDSHAKE(timer)|           |            | HANDSHAKE(timer)|
 * |          | if helloTime > 3| Tx verack | Depend on  | if helloTime > 3|
 * |          | Tx version      |           | node update| Tx version      |
 * |          | then Tx verack  |           |            | then Tx verack  |
 * |-----------------------------------------------------------------------|
 * | verack   |                 | ESTABLISH |            |                 |
 * |          |   No Action     |           | No Action  | No Action       |
 * |------------------------------------------------------------------------
 *
 * The node state switch table after TX message, there is time limitation for each action
 *  ____________________________________________________________
 * |          |    INIT   | HANDSHAKE  | ESTABLISH | INACTIVITY |
 * |------------------------------------------------------------|
 * | version  |           |  INIT      | None      |            |
 * |          | Update    |  Update    |           | Update     |
 * |          | helloTime |  helloTime |           | helloTime  |
 * |------------------------------------------------------------|
 */
// TODO The process should be adjusted based on above table
func rxVersion(node *node, msg *Msg) {
	common.Trace()

	m, err := msg.serialization()
	if (err != nil) {
		log.Println("Error Convert rx version msg ", err.Error())
	}

	str := hex.EncodeToString(m)
	log.Printf("The RX version message length is %d, %s", len(m), str)

	t := time.Now()
	// TODO check version compatible or not
	s := node.getState()
	if (s == HANDSHAKEING) {
		node.setState(HANDSHAKED)
		buf, _ := newVerack()
		log.Println("TX verack")
		go node.tx(buf)
	} else if (s != ESTABLISH) {
		node.setHandshakeTime(t)
		node.setState(HANDSHAKEING)
		buf, _ := newVersion()
		go node.tx(buf)
	}

	// TODO Update other node information

	log.Printf("Node %s state is %d", node.getID(), node.getState())
	node.updateTime(t)
}

func rxVerack(node *node, msg *Msg) {
	common.Trace()
	t := time.Now()
	// TODO we loading the state&time without consider race case
	th := node.getHandshakeTime()
	s := node.getState()

	m, _ := msg.serialization()
	str := hex.EncodeToString(m)
	log.Printf("The message rx verack length is %d, %s", len(m), str)


	// TODO take care about the time duration overflow
	tDelta := t.Sub(th)
	if (tDelta.Seconds() < HELLOTIMEOUT) {
		if (s == HANDSHAKEING) {
			node.setState(ESTABLISH)
			//buf, _ := newHeadersReq()
			buf, _ := newVerack()
			log.Println("Run to here 2")
			go node.tx(buf)
		} else if (s == HANDSHAKED) {
			node.setState(ESTABLISH)
		}
	}

	log.Printf("Node %s state is %d", node.getID(), node.getState())
	node.updateTime(t)
}

func rxGetHeaders (node *node, msg *Msg) {
	common.Trace()
	NetToLedgerCh <- msg
}

func rxHeaders(node *node, msg *Msg) {
	common.Trace()
	NetToLedgerCh <- msg
}

func rxGetaddr(node *node, msg *Msg) {
	common.Trace()
	NetToLedgerCh <- msg
}

func rxAddr(node *node, msg *Msg) {
	common.Trace()
	NetToLedgerCh <- msg
}

func rxConsensus(node *node, msg *Msg) {
	common.Trace()
	NetToConsensusCh <- msg
}

func rxFilteradd(node *node, msg *Msg) {
	common.Trace()
}

func rxFilterClear(node *node, msg *Msg) {
	common.Trace()
}

func rxFilterLoad(node *node, msg *Msg) {
	common.Trace()
}

func rxGetBlocks(node *node, msg *Msg) {
	common.Trace()
	NetToLedgerCh <- msg
}

func rxBlock(node *node, msg *Msg) {
	common.Trace()
	NetToLedgerCh <- msg
}

func rxGetData(node *node, msg *Msg) {
	common.Trace()
}

func rxInv(node *node, msg *Msg) {
	common.Trace()
}

func rxMemPool(node *node, msg *Msg) {
	common.Trace()
}

// Receive the transaction
func rxTransaction(node *node, msg *Msg) {
	common.Trace()
}

func rxAlert(node *node, msg *Msg) {
	common.Trace()
	// TODO Handle Alert
	log.Printf("Alert get from node %s", node.getID())
}

func rxMerkleBlock(node *node, msg *Msg) {
	common.Trace()
}

func rxNotFound(node *node, msg *Msg) {
	common.Trace()
}

func rxPing(node *node, msg *Msg) {
	// TODO
	common.Trace()
}

func rxPong(node *node, msg *Msg) {
	// TODO
	common.Trace()
}

func rxReject(node *node, msg *Msg) {
	// TODO
	common.Trace()
}

func handleNodeMsg(node *node, msg *Msg) {
	// TODO Init parse and check
	var cmd []byte = msg.CMD[:]
	n := bytes.IndexByte(cmd, 0)
	handle, ok := funcMap.handles[string(cmd[:n])]

	if ok == false {
		log.Printf("Unknow node message recevied %s", msg.CMD)
		return
	}
	handle(node, msg)
}

// Trigger handshake
func handshake(node *node) error {
	node.setHandshakeTime(time.Now())
	buf, _ := newVersion()
	go node.tx(buf)

	timer := time.NewTimer(time.Second * HELLOTIMEOUT)
	go func() {
		<-timer.C
		r := node.getHandshakeRetry()
		if ((r < MAXHELLORETYR) && (node.getState() != ESTABLISH)) {
			r++
			node.setHandshakeRetry(r)
			log.Println("Handshake with %s timeout", node.getID())
			handshake(node)
		}
	} ()

	// TODO Does the timer should be recollected?
	return nil
}

func txBlockHeadersReq(node *node) {
	// TODO Need Lock
	if (node.getState() != ESTABLISH) {
		log.Println("Incorrectly node state to send get Header message")
		return
	}

	buf, _ := newHeadersReq()
	go node.tx(buf)
}

func txInventory(node *node) {
	// TODO get transaction entity TX/Block/Consensus

}

func keepAlive(from *node, dst *node) {
	// Need move to node function or keep here?
}

func updateNodeInfo() {
	ticker := time.NewTicker(time.Second * PERIODUPDATETIME)
	quit := make(chan struct{})

	for {
		select {
		case <- ticker.C:
			common.Trace()
			for _, node := range nodes.list {
				h1 := node.getHeight()
				h2 := nodes.node.getHeight()
				if (node.getState() == ESTABLISH) && (h1 > h2) {
					//buf, _ := newMsg("version")
					buf, _ := newMsg("getheaders")
					//buf, _ := newMsg("getaddr")
					go node.tx(buf)
				}
			}
		case <- quit:
			ticker.Stop()
			return
		}
	}
	// TODO when to close the timer
	//close(quit)
}

func StartProtocol() {
	seedNodes, err := config.SeedNodes()

	if err != nil {
		log.Println("Access the config file failure")
		os.Exit(1)
		// TODO should we kick off a blind connection in this case
	}

	for _, nodeAddr := range seedNodes {
		nodes.node.connect(nodeAddr)
	}

	log.Println("Run after go through seed nodes")
	// TODO Housekeeping routine to keep node status update
	updateNodeInfo()
}
