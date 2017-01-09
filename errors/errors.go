package errors

import (
	"errors"
)

const callStackDepth = 10

type DetailError interface {
	error
	ErrCoder
	CallStacker
	GetRoot()  error
}


func  NewErr(errmsg string) error {
	return errors.New(errmsg)
}

func NewDetailErr(err error,errcode ErrCode,errmsg string) DetailError{
	if err == nil {return nil}

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



