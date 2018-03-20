package rest

import Err "github.com/Ontology/http/base/error"

var (
	RspInvalidMethod     = ResponsePack(Err.INVALID_METHOD)
	RspIllegalDataFormat = ResponsePack(Err.ILLEGAL_DATAFORMAT)
	rspInvalidParams     = ResponsePack(Err.INVALID_PARAMS)
	rspInvalidTx         = ResponsePack(Err.INVALID_TRANSACTION)
	rspUnkownBlock       = ResponsePack(Err.UNKNOWN_BLOCK)
	rspUnkownTx          = ResponsePack(Err.UNKNOWN_TRANSACTION)
	rspSmartCodeError    = ResponsePack(Err.SMARTCODE_ERROR)
	rspInternalError     = ResponsePack(Err.INTERNAL_ERROR)
	rspSuccess           = ResponsePack(Err.SUCCESS)
)

func ResponsePack(errCode int64) map[string]interface{} {
	resp := map[string]interface{}{
		"Action":  "",
		"Result":  "",
		"Error":   errCode,
		"Desc":    "",
		"Version": "1.0.0",
	}
	return resp
}
