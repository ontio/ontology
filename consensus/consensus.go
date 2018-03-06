package consensus

import (
	cl "github.com/Ontology/account"
	"github.com/Ontology/common/config"
	"github.com/Ontology/common/log"
	"github.com/Ontology/consensus/dbft"
	"github.com/Ontology/consensus/solo"
	"github.com/Ontology/net"
	"strings"
)

type ConsensusService interface {
	Start() error
	Halt() error
}

const (
	CONSENSUS_DBFT = "dbft"
	CONSENSUS_SOLO = "solo"
)

func NewConsensusService(client cl.Client, localNet net.Neter) ConsensusService {
	consensusType := strings.ToLower(config.Parameters.ConsensusType)
	if consensusType == "" {
		consensusType = CONSENSUS_DBFT
	}
	var consensus ConsensusService
	switch consensusType {
	case CONSENSUS_DBFT:
		consensus = dbft.NewDbftService(client, "dbft", localNet)
	case CONSENSUS_SOLO:
		consensus = solo.NewSoloService(client, localNet)
	}
	log.Infof("ConsensusType:%s", consensusType)
	return consensus
}
