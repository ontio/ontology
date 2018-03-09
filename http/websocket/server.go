package websocket

import (
	"bytes"
	. "github.com/Ontology/common"
	. "github.com/Ontology/common/config"
	"github.com/Ontology/core/ledger"
	"github.com/Ontology/core/types"
	. "github.com/Ontology/http/base/actor"
	. "github.com/Ontology/http/base/rest"
	Err "github.com/Ontology/http/base/error"
	"github.com/Ontology/http/websocket/websocket"
	sc "github.com/Ontology/smartcontract/common"
	"github.com/Ontology/smartcontract/event"
	"github.com/Ontology/events/message"
)

var ws *websocket.WsServer
var (
	pushBlockFlag    bool = false
	pushRawBlockFlag bool = false
	pushBlockTxsFlag bool = false
)

func StartServer() {
	SubscribeEvent(message.TopicSaveBlockComplete,SendBlock2WSclient)
	SubscribeEvent(message.TopicSmartCodeEvent,PushSmartCodeEvent)
	go func() {
		ws = websocket.InitWsServer()
		ws.Start()
	}()
}
func SendBlock2WSclient(v interface{}) {
	if Parameters.HttpWsPort != 0 && pushBlockFlag {
		go func() {
			PushBlock(v)
		}()
	}
	if Parameters.HttpWsPort != 0 && pushBlockTxsFlag {
		go func() {
			PushBlockTransactions(v)
		}()
	}
}
func Stop() {
	if ws == nil {
		return
	}
	ws.Stop()
}
func ReStartServer() {
	if ws == nil {
		ws = websocket.InitWsServer()
		ws.Start()
		return
	}
	ws.Restart()
}
func GetWsPushBlockFlag() bool {
	return pushBlockFlag
}
func SetWsPushBlockFlag(b bool) {
	pushBlockFlag = b
}
func GetPushRawBlockFlag() bool {
	return pushRawBlockFlag
}
func SetPushRawBlockFlag(b bool) {
	pushRawBlockFlag = b
}
func GetPushBlockTxsFlag() bool {
	return pushBlockTxsFlag
}
func SetPushBlockTxsFlag(b bool) {
	pushBlockTxsFlag = b
}
func SetTxHashMap(txhash string, sessionid string) {
	if ws == nil {
		return
	}
	ws.SetTxHashMap(txhash, sessionid)
}

func PushSmartCodeEvent(v interface{}) {
	if ws != nil {
		rs, ok := v.(map[string]interface{})
		if !ok {
			return
		}
		go func() {
			switch object := rs["Result"].(type) {
			case event.LogEventArgs:
				type LogEventArgsInfo struct {
					Container   string
					CodeHash    string
					Message     string
					BlockHeight uint32
				}
				msg := LogEventArgsInfo{
					Container:   ToHexString(object.Container.ToArray()),
					CodeHash:    ToHexString(object.CodeHash.ToArray()),
					Message:     object.Message,
					BlockHeight: ledger.DefLedger.GetCurrentBlockHeight() + 1,
				}
				PushEvent(rs["TxHash"].(string), rs["Error"].(int64), rs["Action"].(string), msg)
				return
			case event.NotifyEventArgs:
				type NotifyEventArgsInfo struct {
					Container   string
					CodeHash    string
					State       []sc.States
					BlockHeight uint32
				}
				msg := NotifyEventArgsInfo{
					Container:   ToHexString(object.Container.ToArray()),
					CodeHash:    ToHexString(object.CodeHash.ToArray()),
					State:       sc.ConvertTypes(object.States),
					BlockHeight: ledger.DefLedger.GetCurrentBlockHeight() + 1,
				}
				PushEvent(rs["TxHash"].(string), rs["Error"].(int64), rs["Action"].(string), msg)
				return
			default:
				PushEvent(rs["TxHash"].(string), rs["Error"].(int64), rs["Action"].(string), rs["Result"])
				return
			}
		}()
	}
}

func PushEvent(txHash string, errcode int64, action string, result interface{}) {
	if ws != nil {
		resp := ResponsePack(Err.SUCCESS)
		resp["Result"] = result
		resp["Error"] = errcode
		resp["Action"] = action
		resp["Desc"] = Err.ErrMap[resp["Error"].(int64)]
		ws.PushTxResult(txHash, resp)
		//ws.BroadcastResult(resp)
	}
}

func PushBlock(v interface{}) {
	if ws == nil {
		return
	}
	resp := ResponsePack(Err.SUCCESS)
	if block, ok := v.(*types.Block); ok {
		if pushRawBlockFlag {
			w := bytes.NewBuffer(nil)
			block.Serialize(w)
			resp["Result"] = ToHexString(w.Bytes())
		} else {
			resp["Result"] = GetBlockInfo(block)
		}
		resp["Action"] = "sendrawblock"
		ws.BroadcastResult(resp)
	}
}
func PushBlockTransactions(v interface{}) {
	if ws == nil {
		return
	}
	resp := ResponsePack(Err.SUCCESS)
	if block, ok := v.(*types.Block); ok {
		if pushBlockTxsFlag {
			resp["Result"] = GetBlockTransactions(block)
		}
		resp["Action"] = "sendblocktransactions"
		ws.BroadcastResult(resp)
	}
}
