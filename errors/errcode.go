package errors

import (
	"fmt"
)

type ErrCoder interface {
	GetErrCode() ErrCode
}

type ErrCode int16

const (
	ErrNoCode			ErrCode = -2
	ErrNoError                      ErrCode = 0
	ErrUnknown                      ErrCode = -1
	ErrDuplicatedTx 		ErrCode = 1
)

func (err ErrCode) Error() string {
	switch err {
	case ErrNoCode:
		return "No error code"
	case ErrNoError:
		return "Not an error"
	case ErrUnknown:
		return "Unknown error"
	case ErrDuplicatedTx:
		return "There are duplicated Transactions"

	}

	return fmt.Sprintf("Unknown error? Error code = %d", err)
}


func ErrerCode(err error) ErrCode{
	if err, ok := err.(ErrCoder); ok {
		return err.GetErrCode()
	}
	return ErrUnknown
}
