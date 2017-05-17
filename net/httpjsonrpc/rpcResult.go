package httpjsonrpc

var (
	DnaRpcInvalidHash        = responsePacking("invalid hash")
	DnaRpcInvalidBlock       = responsePacking("invalid block")
	DnaRpcInvalidTransaction = responsePacking("invalid transaction")
	DnaRpcInvalidParameter   = responsePacking("invalid parameter")

	DnaRpcUnknownBlock       = responsePacking("unknown block")
	DnaRpcUnknownTransaction = responsePacking("unknown transaction")

	DnaRpcNil           = responsePacking(nil)
	DnaRpcUnsupported   = responsePacking("Unsupported")
	DnaRpcInternalError = responsePacking("internal error")

	DnaRpcSuccess = responsePacking(true)
	DnaRpcFailed  = responsePacking(false)

	DnaRpc = responsePacking
)
