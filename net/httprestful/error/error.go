package error

import . "DNA/errors"

const (
	SUCCESS            int64 = 0
	SESSION_EXPIRED    int64 = 41001
	SERVICE_CEILING    int64 = 41002
	ILLEGAL_DATAFORMAT int64 = 41003
	OAUTH_TIMEOUT      int64 = 41004

	INVALID_METHOD int64 = 42001
	INVALID_PARAMS int64 = 42002
	INVALID_TOKEN  int64 = 42003

	INVALID_TRANSACTION int64 = 43001
	INVALID_ASSET       int64 = 43002
	INVALID_BLOCK       int64 = 43003

	UNKNOWN_TRANSACTION int64 = 44001
	UNKNOWN_ASSET       int64 = 44002
	UNKNOWN_BLOCK       int64 = 44003

	INVALID_VERSION int64 = 45001
	INTERNAL_ERROR  int64 = 45002

	OAUTH_INVALID_APPID    int64 = 46001
	OAUTH_INVALID_CHECKVAL int64 = 46002
	SMARTCODE_ERROR        int64 = 47001
)

var ErrMap = map[int64]string{
	SUCCESS:            "SUCCESS",
	SESSION_EXPIRED:    "SESSION EXPIRED",
	SERVICE_CEILING:    "SERVICE CEILING",
	ILLEGAL_DATAFORMAT: "ILLEGAL DATAFORMAT",
	OAUTH_TIMEOUT:      "CONNECT TO OAUTH TIMEOUT",

	INVALID_METHOD: "INVALID METHOD",
	INVALID_PARAMS: "INVALID PARAMS",
	INVALID_TOKEN:  "VERIFY TOKEN ERROR",

	INVALID_TRANSACTION: "INVALID TRANSACTION",
	INVALID_ASSET:       "INVALID ASSET",
	INVALID_BLOCK:       "INVALID BLOCK",

	UNKNOWN_TRANSACTION: "UNKNOWN TRANSACTION",
	UNKNOWN_ASSET:       "UNKNOWN ASSET",
	UNKNOWN_BLOCK:       "UNKNOWN BLOCK",

	INVALID_VERSION:                "INVALID VERSION",
	INTERNAL_ERROR:                 "INTERNAL ERROR",
	SMARTCODE_ERROR:                "SMARTCODE EXEC ERROR",
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
}
