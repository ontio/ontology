package serialization

const (
	ByteArrayType byte = 0x00
	AddressType   byte = 0x01
	BooleanType   byte = 0x02
	IntType       byte = 0x03
	H256Type      byte = 0x04
	//reserved for other types
	ListType byte = 0x10

	MAX_PARAM_LENGTH      = 1024
	VERSION          byte = 0
)