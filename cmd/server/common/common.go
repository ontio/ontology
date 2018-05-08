package common

import (
	"encoding/json"
	"github.com/ontio/ontology/account"
)

var DefAccount *account.Account

type CliRpcRequest struct {
	Qid    string          `json:"qid"`
	Params json.RawMessage `json:"params"`
	Method string          `json:"method"`
}

type CliRpcResponse struct {
	Qid       string      `json:"qid"`
	Method    string      `json:"method"`
	Result    interface{} `json:"result"`
	ErrorCode int         `json:"error_code"`
	ErrorInfo string      `json:"error_info"`
}
