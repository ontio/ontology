package rpc

import (
	Err "github.com/Ontology/http/base/error"
)

var (
	DnaRpcInvalidHash        = responsePacking(Err.INVALID_PARAMS, "invalid hash")
	DnaRpcInvalidBlock       = responsePacking(Err.INVALID_BLOCK, "invalid block")
	DnaRpcInvalidTransaction = responsePacking(Err.INVALID_TRANSACTION, "invalid transaction")
	DnaRpcInvalidParameter   = responsePacking(Err.INVALID_PARAMS, "invalid parameter")

	DnaRpcUnknownBlock       = responsePacking(Err.UNKNOWN_BLOCK, "unknown block")
	DnaRpcUnknownTransaction = responsePacking(Err.UNKNOWN_TRANSACTION, "unknown transaction")

	DnaRpcNil             = responsePacking(Err.INVALID_PARAMS, nil)
	DnaRpcUnsupported     = responsePacking(Err.INTERNAL_ERROR, "Unsupported")
	DnaRpcInternalError   = responsePacking(Err.INTERNAL_ERROR, "internal error")
	DnaRpcIOError         = responsePacking(Err.INTERNAL_ERROR, "internal IO error")
	DnaRpcAPIError        = responsePacking(Err.INTERNAL_ERROR, "internal API error")
	DnaRpcSuccess         = responsePacking(Err.SUCCESS, true)
	DnaRpcFailed          = responsePacking(Err.INTERNAL_ERROR, false)
	DnaRpcAccountNotFound = responsePacking(Err.INTERNAL_ERROR, "Account not found")

	DnaRpc = responseSuccess
)

func responseSuccess(result interface{}) map[string]interface{} {
	return responsePacking(Err.SUCCESS, result)
}
func responsePacking(errcode int64, result interface{}) map[string]interface{} {
	resp := map[string]interface{}{
		"error":  errcode,
		"desc":   Err.ErrMap[errcode],
		"result": result,
	}
	return resp
}
