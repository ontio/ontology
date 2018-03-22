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
