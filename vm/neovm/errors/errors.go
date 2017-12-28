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
)
