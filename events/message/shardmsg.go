package message

const (
	ShardGetGenesisBlockReq = iota
	ShardGetGenesisBlockRsp
	ShardGetPeerInfoReq
	ShardGetPeerInfoRsp
)

type ShardSystemEventMsg struct {
	MsgType    int    `json:"msg_type"`
	MsgPayload []byte `json:"msg_payload"`
}
