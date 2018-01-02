package types

type VmType byte

const (
	NEOVM VmType = iota
	EVM
)

type TriggerType byte

const (
	Verification TriggerType = iota
	Application = 0x10
)
