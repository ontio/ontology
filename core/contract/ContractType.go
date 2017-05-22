package contract

type ContractType byte

const (
	SignatureContract ContractType = iota
	MultiSigContract
	CustomContract
)
