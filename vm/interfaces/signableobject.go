package interfaces

type ISignableObject interface {
	GetMessage() ([]byte)
}
