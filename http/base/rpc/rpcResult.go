package rpc

var (
	DnaRpcInvalidHash        = responsePacking("invalid hash")
	DnaRpcInvalidBlock       = responsePacking("invalid block")
	DnaRpcInvalidTransaction = responsePacking("invalid transaction")
	DnaRpcInvalidParameter   = responsePacking("invalid parameter")

	DnaRpcUnknownBlock       = responsePacking("unknown block")
	DnaRpcUnknownTransaction = responsePacking("unknown transaction")

	DnaRpcNil             = responsePacking(nil)
	DnaRpcUnsupported     = responsePacking("Unsupported")
	DnaRpcInternalError   = responsePacking("internal error")
	DnaRpcIOError         = responsePacking("internal IO error")
	DnaRpcAPIError        = responsePacking("internal API error")
	DnaRpcSuccess         = responsePacking(true)
	DnaRpcFailed          = responsePacking(false)
	DnaRpcAccountNotFound = responsePacking(("Account not found"))

	DnaRpc = responsePacking
)

func responsePacking(result interface{}) map[string]interface{} {
	resp := map[string]interface{}{
		"result": result,
	}
	return resp
}
