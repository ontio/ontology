package common

//node info const
const (
	VERIFYNODE      = 1
	SERVICENODE     = 2
	VERIFYNODENAME  = "verify"
	SERVICENODENAME = "service"
	PROTOCOLVERSION = 0
)

//msg const
const (
	MSGCMDLEN         = 12
	CMDOFFSET         = 4
	CHECKSUMLEN       = 4
	HASHLEN           = 32 // hash length in byte
	MSGHDRLEN         = 24
	NETMAGIC          = 0x74746e41
	DIVHASHLEN        = 5
	MAXREQBLKONCE     = 16
	TIMESOFUPDATETIME = 2
)

//info update const
const (
	HELLOTIMEOUT     = 3 // Seconds
	MAXHELLORETYR    = 3
	MAXBUFLEN        = 1024 * 16 // Fixme The maximum buffer to receive message
	MAXCHANBUF       = 512
	MAXBLKHDRCNT     = 500
	MAXINVHDRCNT     = 500
	PERIODUPDATETIME = 3 // Time to update and sync information with other nodes
	HEARTBEAT        = 2
	KEEPALIVETIMEOUT = 3
	DIALTIMEOUT      = 6
	CONNMONITOR      = 6
	CONNMAXBACK      = 4000
	MAXRETRYCOUNT    = 3
)

// The peer state
const (
	INIT       = 0
	HAND       = 1
	HANDSHAKE  = 2
	HANDSHAKED = 3
	ESTABLISH  = 4
	INACTIVITY = 5
)

var ReceiveDuplicateBlockCnt uint64 //an index to detecting networking status

type PeerAddr struct {
	Time          int64
	Services      uint64
	IpAddr        [16]byte
	Port          uint16
	ConsensusPort uint16
	ID            uint64 // Unique ID
}

type MsgPayload struct {
	Id      uint64
	Payload []byte
	Len     int
}

const (
	CONNECT      = 0
	DISCONNECT   = 1
	P2PMSG       = 2
	VERSION_TYPE = "version"
)
