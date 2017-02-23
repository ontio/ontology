package vm


import "errors"

var (
	ErrBadValue           = errors.New("bad value")
	ErrOverLen	      = errors.New("the count over the size")
	ErrFault	      = errors.New("The exeution meet fault")
)
