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
	"errors"
)

const callStackDepth = 10

type DetailError interface {
	error
	ErrCoder
	CallStacker
	GetRoot() error
}

func NewErr(errmsg string) error {
	return errors.New(errmsg)
}

func NewDetailErr(err error, errcode ErrCode, errmsg string) DetailError {
	if err == nil {
		return nil
	}

	dnaerr, ok := err.(dnaError)
	if !ok {
		dnaerr.root = err
		dnaerr.errmsg = err.Error()
		dnaerr.callstack = getCallStack(0, callStackDepth)
		dnaerr.code = errcode

	}
	if errmsg != "" {
		dnaerr.errmsg = errmsg + ": " + dnaerr.errmsg
	}

	return dnaerr
}

func RootErr(err error) error {
	if err, ok := err.(DetailError); ok {
		return err.GetRoot()
	}
	return err
}
