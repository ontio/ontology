package vm

type ISignableObject interface {
	GetMessage() ([]byte)
}
