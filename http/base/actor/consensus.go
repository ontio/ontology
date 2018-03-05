package actor

import (
	"time"
	"github.com/Ontology/eventbus/actor"
)

var consensusSrvPid *actor.PID

func SetConsensusActor(actr *actor.PID) {
	consensusSrvPid = actr
}

func ConsensusSrvStart() (error) {
	future := consensusSrvPid.RequestFuture(nil, 10*time.Second)
	_, err := future.Result()
	if err != nil {
		return nil
	}
	return nil
}
func ConsensusSrvHalt() (error) {
	future := consensusSrvPid.RequestFuture(nil, 10*time.Second)
	_, err := future.Result()
	if err != nil {
		return nil
	}
	return nil
}
