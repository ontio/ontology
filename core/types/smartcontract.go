package types

type SmartCodeEvent struct {
	TxHash string
	Action string
	Result interface{}
	Error  int64
}
