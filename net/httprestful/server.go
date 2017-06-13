package httprestful

import (
	. "DNA/common/config"
	"DNA/core/ledger"
	"DNA/events"
	"DNA/net/httprestful/common"
	. "DNA/net/httprestful/restful"
	. "DNA/net/protocol"
	"strconv"
)

const OAUTH_SSUCCESS_CODE = "r0000"

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

func checkAccessToken(auth_type, access_token string) bool {
	if len(Parameters.OauthServerAddr) == 0 {
		return true
	}
	req := make(map[string]interface{})
	req["token"] = access_token
	req["auth_type"] = auth_type
	repMsg, err := common.PostRequest(req, Parameters.OauthServerAddr)
	if err != nil {
		return false
	}
	if repMsg["code"] == OAUTH_SSUCCESS_CODE {
		return true
	}
	return false
}
