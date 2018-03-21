package rpc

import (
	Err "github.com/Ontology/http/base/error"
)

var (
	RpcInvalidHash        = responsePacking(Err.INVALID_PARAMS, "invalid hash")
	RpcInvalidBlock       = responsePacking(Err.INVALID_BLOCK, "invalid block")
	RpcInvalidTransaction = responsePacking(Err.INVALID_TRANSACTION, "invalid transaction")
	RpcInvalidParameter   = responsePacking(Err.INVALID_PARAMS, "invalid parameter")

	RpcUnknownBlock       = responsePacking(Err.UNKNOWN_BLOCK, "unknown block")
	RpcUnknownTransaction = responsePacking(Err.UNKNOWN_TRANSACTION, "unknown transaction")
	RpcUnKnownContact = responsePacking(Err.UNKNWN_CONTRACT, "unknown contract")

	RpcNil           = responsePacking(Err.INVALID_PARAMS, nil)
	RpcUnsupported   = responsePacking(Err.INTERNAL_ERROR, "Unsupported")
	RpcInternalError = responsePacking(Err.INTERNAL_ERROR, "internal error")
	RpcIOError       = responsePacking(Err.INTERNAL_ERROR, "internal IO error")
	RpcAPIError      = responsePacking(Err.INTERNAL_ERROR, "internal API error")

	RpcFailed          = responsePacking(Err.INTERNAL_ERROR, false)
	RpcAccountNotFound = responsePacking(Err.INTERNAL_ERROR, "Account not found")

	RpcSuccess = responsePacking(Err.SUCCESS, true)
	Rpc        = responseSuccess
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
