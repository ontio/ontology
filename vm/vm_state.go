package vm

type VMState byte

const (
	NONE  VMState = 0
	HALT  VMState = 1 << 0
	FAULT VMState = 1 << 1
	BREAK VMState = 1 << 2

	INSUFFICIENT_RESOURCE VMState = 1 << 4
)
