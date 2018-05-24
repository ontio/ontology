/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package error

import ontErrors "github.com/ontio/ontology/errors"

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
	UNKNOWN_CONTRACT    int64 = 44004

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
	UNKNOWN_CONTRACT:    "UNKNOWN CONTRACT",

	INTERNAL_ERROR:                           "INTERNAL ERROR",
	SMARTCODE_ERROR:                          "SMARTCODE EXEC ERROR",
	int64(ontErrors.ErrDuplicatedTx):         "INTERNAL ERROR, ErrDuplicatedTx",
	int64(ontErrors.ErrDuplicateInput):       "INTERNAL ERROR, ErrDuplicateInput",
	int64(ontErrors.ErrAssetPrecision):       "INTERNAL ERROR, ErrAssetPrecision",
	int64(ontErrors.ErrTransactionBalance):   "INTERNAL ERROR, ErrTransactionBalance",
	int64(ontErrors.ErrAttributeProgram):     "INTERNAL ERROR, ErrAttributeProgram",
	int64(ontErrors.ErrTransactionContracts): "INTERNAL ERROR, ErrTransactionContracts",
	int64(ontErrors.ErrTransactionPayload):   "INTERNAL ERROR, ErrTransactionPayload",
	int64(ontErrors.ErrDoubleSpend):          "INTERNAL ERROR, ErrDoubleSpend",
	int64(ontErrors.ErrTxHashDuplicate):      "INTERNAL ERROR, ErrTxHashDuplicate",
	int64(ontErrors.ErrStateUpdaterVaild):    "INTERNAL ERROR, ErrStateUpdaterVaild",
	int64(ontErrors.ErrSummaryAsset):         "INTERNAL ERROR, ErrSummaryAsset",
	int64(ontErrors.ErrXmitFail):             "INTERNAL ERROR, ErrXmitFail",
	int64(ontErrors.ErrNoAccount):            "INTERNAL ERROR, ErrNoAccount",
}
