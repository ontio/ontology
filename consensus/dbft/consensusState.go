package dbft

type ConsensusState byte

const (
	Initial         ConsensusState = 0x00
	Primary         ConsensusState = 0x01
	Backup          ConsensusState = 0x02
	RequestSent     ConsensusState = 0x04
	RequestReceived ConsensusState = 0x08
	SignatureSent   ConsensusState = 0x10
	BlockGenerated  ConsensusState = 0x20
)

func (state ConsensusState) HasFlag(flag ConsensusState) bool {
	return (state & flag) == flag
}
