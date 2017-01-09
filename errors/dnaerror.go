package errors

type dnaError struct {
	errmsg string
	callstack *CallStack
	root error
	code ErrCode
}

func (e dnaError) Error() string {
	return e.errmsg
}

func (e dnaError) GetErrCode()  ErrCode {
	return e.code
}

func (e dnaError) GetRoot()  error {
	return e.root
}

func (e dnaError) GetCallStack()  *CallStack {
	return e.callstack
}
