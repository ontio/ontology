package error

import . "github.com/Ontology/errors"

const (
	SUCCESS            int64 = 0
	SESSION_EXPIRED    int64 = 41001
	SERVICE_CEILING    int64 = 41002
	ILLEGAL_DATAFORMAT int64 = 41003
	INVALID_VERSION    int64 = 41004

	INVALID_METHOD int64 = 42001
	INVALID_PARAMS int64 = 42002

	INVALID_TRANSACTION int64 = 43001
	INVALID_ASSET       int64 = 43002
	INVALID_BLOCK       int64 = 43003

	UNKNOWN_TRANSACTION int64 = 44001
	UNKNOWN_ASSET       int64 = 44002
	UNKNOWN_BLOCK       int64 = 44003

	INTERNAL_ERROR  int64 = 45001
	SMARTCODE_ERROR int64 = 47001
)

var ErrMap = map[int64]string{
	SUCCESS:            "SUCCESS",
	SESSION_EXPIRED:    "SESSION EXPIRED",
	SERVICE_CEILING:    "SERVICE CEILING",
	ILLEGAL_DATAFORMAT: "ILLEGAL DATAFORMAT",
	INVALID_VERSION:    "INVALID VERSION",

	INVALID_METHOD: "INVALID METHOD",
	INVALID_PARAMS: "INVALID PARAMS",

	INVALID_TRANSACTION: "INVALID TRANSACTION",
	INVALID_ASSET:       "INVALID ASSET",
	INVALID_BLOCK:       "INVALID BLOCK",

	UNKNOWN_TRANSACTION: "UNKNOWN TRANSACTION",
	UNKNOWN_ASSET:       "UNKNOWN ASSET",
	UNKNOWN_BLOCK:       "UNKNOWN BLOCK",

	INTERNAL_ERROR:                 "INTERNAL ERROR",
	SMARTCODE_ERROR:                "SMARTCODE EXEC ERROR",
	int64(ErrDuplicatedTx):         "INTERNAL ERROR, ErrDuplicatedTx",
	int64(ErrDuplicateInput):       "INTERNAL ERROR, ErrDuplicateInput",
	int64(ErrAssetPrecision):       "INTERNAL ERROR, ErrAssetPrecision",
	int64(ErrTransactionBalance):   "INTERNAL ERROR, ErrTransactionBalance",
	int64(ErrAttributeProgram):     "INTERNAL ERROR, ErrAttributeProgram",
	int64(ErrTransactionContracts): "INTERNAL ERROR, ErrTransactionContracts",
	int64(ErrTransactionPayload):   "INTERNAL ERROR, ErrTransactionPayload",
	int64(ErrDoubleSpend):          "INTERNAL ERROR, ErrDoubleSpend",
	int64(ErrTxHashDuplicate):      "INTERNAL ERROR, ErrTxHashDuplicate",
	int64(ErrStateUpdaterVaild):    "INTERNAL ERROR, ErrStateUpdaterVaild",
	int64(ErrSummaryAsset):         "INTERNAL ERROR, ErrSummaryAsset",
	int64(ErrXmitFail):             "INTERNAL ERROR, ErrXmitFail",
	int64(ErrNoAccount):            "INTERNAL ERROR, ErrNoAccount",
}
