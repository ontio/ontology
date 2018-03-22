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

import "errors"

var (
	ErrBadValue              = errors.New("bad value")
	ErrBadType               = errors.New("bad type")
	ErrOverStackLen          = errors.New("the count over the stack length")
	ErrOverCodeLen           = errors.New("the count over the code length")
	ErrUnderStackLen         = errors.New("the count under the stack length")
	ErrFault                 = errors.New("the exeution meet fault")
	ErrNotSupportService     = errors.New("the service is not registered")
	ErrNotSupportOpCode      = errors.New("does not support the operation code")
	ErrOverLimitStack        = errors.New("the stack over max size")
	ErrOverMaxItemSize       = errors.New("the item over max size")
	ErrOverMaxArraySize      = errors.New("the array over max size")
	ErrOverMaxBigIntegerSize = errors.New("the biginteger over max size 32bit")
	ErrOutOfGas              = errors.New("out of gas")
	ErrNotArray              = errors.New("not array")
	ErrTableIsNil            = errors.New("table is nil")
	ErrServiceIsNil          = errors.New("service is nil")
	ErrDivModByZero          = errors.New("div or mod by zore")
	ErrShiftByNeg            = errors.New("shift by negtive value")
	ErrExecutionContextNil   = errors.New("execution context is nil")
	ErrCurrentContextNil     = errors.New("current context is nil")
	ErrCallingContextNil     = errors.New("calling context is nil")
	ErrEntryContextNil       = errors.New("entry context is nil")
	ErrAppendNotArray        = errors.New("append not array")
)
