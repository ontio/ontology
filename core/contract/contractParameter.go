package contract

//parameter defined type.
type ContractParameterType byte

const (
	Signature ContractParameterType = iota
	Boolean
	Integer
	Hash160
	Hash256
	ByteArray
	PublicKey
	String
	Array = 0x10
	InteropInterface = 0xf0
	Void = 0xff
)
