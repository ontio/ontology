package consensus

type ConsensusService interface {
	Start() error
	Halt() error
}


