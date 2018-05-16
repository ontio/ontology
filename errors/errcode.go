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

package errors

import (
	"fmt"
)

type ErrCoder interface {
	GetErrCode() ErrCode
}

type ErrCode int32

const (
	ErrNoCode               ErrCode = -2
	ErrNoError              ErrCode = 0
	ErrUnknown              ErrCode = -1
	ErrDuplicatedTx         ErrCode = 45002
	ErrDuplicateInput       ErrCode = 45003
	ErrAssetPrecision       ErrCode = 45004
	ErrTransactionBalance   ErrCode = 45005
	ErrAttributeProgram     ErrCode = 45006
	ErrTransactionContracts ErrCode = 45007
	ErrTransactionPayload   ErrCode = 45008
	ErrDoubleSpend          ErrCode = 45009
	ErrTxHashDuplicate      ErrCode = 45010
	ErrStateUpdaterVaild    ErrCode = 45011
	ErrSummaryAsset         ErrCode = 45012
	ErrXmitFail             ErrCode = 45013
	ErrNoAccount            ErrCode = 45014
	ErrRetryExhausted       ErrCode = 45015
	ErrTxPoolFull           ErrCode = 45016
	ErrNetPackFail          ErrCode = 45017
	ErrNetUnPackFail        ErrCode = 45018
	ErrNetVerifyFail        ErrCode = 45019
)

func (err ErrCode) Error() string {
	switch err {
	case ErrNoCode:
		return "no error code"
	case ErrNoError:
		return "not an error"
	case ErrUnknown:
		return "unknown error"
	case ErrDuplicatedTx:
		return "duplicated transaction detected"
	case ErrDuplicateInput:
		return "duplicated transaction input detected"
	case ErrAssetPrecision:
		return "invalid asset precision"
	case ErrTransactionBalance:
		return "transaction balance unmatched"
	case ErrAttributeProgram:
		return "attribute program error"
	case ErrTransactionContracts:
		return "invalid transaction contract"
	case ErrTransactionPayload:
		return "invalid transaction payload"
	case ErrDoubleSpend:
		return "double spent transaction detected"
	case ErrTxHashDuplicate:
		return "duplicated transaction hash detected"
	case ErrStateUpdaterVaild:
		return "invalid state updater"
	case ErrSummaryAsset:
		return "invalid summary asset"
	case ErrXmitFail:
		return "transmit error"
	case ErrRetryExhausted:
		return "retry exhausted"
	case ErrTxPoolFull:
		return "tx pool full"
	case ErrNetPackFail:
		return "net msg pack fail"
	case ErrNetUnPackFail:
		return "net msg unpack fail"
	case ErrNetVerifyFail:
		return "net msg verify fail"
	}

	return fmt.Sprintf("Unknown error? Error code = %d", err)
}

func ErrerCode(err error) ErrCode {
	if err, ok := err.(ErrCoder); ok {
		return err.GetErrCode()
	}
	return ErrUnknown
}
