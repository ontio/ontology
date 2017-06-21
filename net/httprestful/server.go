package httprestful

import (
	. "DNA/common/config"
	"DNA/common/log"
	"DNA/core/ledger"
	"DNA/events"
	"DNA/net/httprestful/common"
	Err "DNA/net/httprestful/error"
	. "DNA/net/httprestful/restful"
	. "DNA/net/protocol"
	"strconv"
)

const OAUTH_SUCCESS_CODE = "r0000"

func StartServer(n Noder) {
	common.SetNode(n)
	ledger.DefaultLedger.Blockchain.BCEvents.Subscribe(events.EventBlockPersistCompleted, SendBlock2NoticeServer)
	func() common.ApiServer {
		rest := InitRestServer(checkAccessToken)
		go rest.Start()
		return rest
	}()
}

func SendBlock2NoticeServer(v interface{}) {

	if len(Parameters.NoticeServerAddr) == 0 || !common.CheckPushBlock() {
		return
	}
	go func() {
		req := make(map[string]interface{})
		req["Height"] = strconv.FormatInt(int64(ledger.DefaultLedger.Blockchain.BlockHeight), 10)
		req = common.GetBlockByHeight(req)

		repMsg, _ := common.PostRequest(req, Parameters.NoticeServerAddr)
		if repMsg[""] == nil {
			//TODO
		}
	}()
}

func checkAccessToken(auth_type, access_token string) (cakey string, errCode int64, result interface{}) {

	if len(Parameters.OauthServerAddr) == 0 {
		return "", Err.SUCCESS, ""
	}
	req := make(map[string]interface{})
	req["token"] = access_token
	req["auth_type"] = auth_type
	repMsg, err := common.OauthRequest("GET", req, Parameters.OauthServerAddr)
	if err != nil {
		log.Error("Oauth timeout:", err)
		return "", Err.OAUTH_TIMEOUT, repMsg
	}
	if repMsg["code"] == OAUTH_SUCCESS_CODE {
		msg, ok := repMsg["msg"].(map[string]interface{})
		if !ok {
			return "", Err.INVALID_TOKEN, repMsg
		}
		if CAkey, ok := msg["cakey"].(string); ok {
			return CAkey, Err.SUCCESS, repMsg
		}
	}
	return "", Err.INVALID_TOKEN, repMsg
}
