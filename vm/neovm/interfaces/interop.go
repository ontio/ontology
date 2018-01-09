package interfaces

type IInteropInterface interface {
	ToArray() []byte
	Clone() IInteropInterface
}
