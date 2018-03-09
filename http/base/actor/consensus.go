package actor

import (
	"github.com/Ontology/eventbus/actor"
	actorTypes "github.com/Ontology/consensus/actor"
)

var consensusSrvPid *actor.PID

func SetConsensusActor(actr *actor.PID) {
	consensusSrvPid = actr
}

func ConsensusSrvStart() (error) {
	consensusSrvPid.Tell(&actorTypes.StartConsensus{})
	return nil
}
func ConsensusSrvHalt() (error) {
	consensusSrvPid.Tell(&actorTypes.StopConsensus{})
	return nil
}
