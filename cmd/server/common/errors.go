package common

const (
	CLIERR_OK                  = 0
	CLIERR_HTTP_METHOD_INVALID = 1001
	CLIERR_INVALID_REQUEST     = 1002
	CLIERR_INVALID_PARAMS      = 1003
	CLIERR_UNSUPPORT_METHOD    = 1004
	CLIERR_ACCOUNT_UNLOCK      = 1005
	CLIERR_INVALID_TX          = 1006
	CLIERR_INTERNAL_ERR        = 900
)

var RPCErrorDesc = map[int]string{
	CLIERR_OK:                  "",
	CLIERR_HTTP_METHOD_INVALID: "invalid http method",
	CLIERR_INVALID_REQUEST:     "invalid request",
	CLIERR_INVALID_PARAMS:      "invalid params",
	CLIERR_UNSUPPORT_METHOD:    "unsupport method",
	CLIERR_INVALID_TX:          "invalid tx",
	CLIERR_INTERNAL_ERR:        "internal error",
}

func GetCLIErrorDesc(errorCode int) string {
	desc, ok := RPCErrorDesc[errorCode]
	if !ok {
		return RPCErrorDesc[CLIERR_INTERNAL_ERR]
	}
	return desc
}
