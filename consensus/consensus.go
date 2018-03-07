package consensus

import (
	"strings"

	"github.com/Ontology/account"
	"github.com/Ontology/common/config"
	"github.com/Ontology/common/log"
	"github.com/Ontology/consensus/dbft"
	"github.com/Ontology/consensus/solo"
	"github.com/Ontology/eventbus/actor"
	"github.com/Ontology/net"
)

type ConsensusService interface {
	Start() error
	Halt() error
	GetPID() *actor.PID
}

const (
	CONSENSUS_DBFT = "dbft"
	CONSENSUS_SOLO = "solo"
)

func NewConsensusService(account *account.Account, txpool *actor.PID, ledger *actor.PID, localNet net.Neter) (ConsensusService, error) {
	consensusType := strings.ToLower(config.Parameters.ConsensusType)
	if consensusType == "" {
		consensusType = CONSENSUS_DBFT
	}

	var consensus ConsensusService
	var err error
	switch consensusType {
	case CONSENSUS_DBFT:
		consensus, err = dbft.NewDbftService(account, txpool)
	case CONSENSUS_SOLO:
		consensus, err = solo.NewSoloService(account, txpool, ledger)
	}
	log.Infof("ConsensusType:%s", consensusType)
	return consensus, err
}
