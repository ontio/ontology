package errors


import "errors"

var (
	ErrBadValue           = errors.New("bad value")
	ErrBadType            = errors.New("bad type")
	ErrOverLen	      = errors.New("the count over the size")
	ErrFault	      = errors.New("The exeution meet fault")
)
