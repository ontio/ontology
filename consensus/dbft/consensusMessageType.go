package dbft

type ConsensusMessageType byte

const (
	ChangeViewMsg ConsensusMessageType = 0x00
	PrepareRequestMsg ConsensusMessageType = 0x20
	PrepareResponseMsg ConsensusMessageType = 0x21
)