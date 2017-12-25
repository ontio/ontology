package interfaces

type ICodeTable interface {
	GetCode(scriptHash []byte) ([]byte, error)
}
