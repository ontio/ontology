package actor

import (
	"github.com/Ontology/eventbus/actor"
	//"github.com/Ontology/net/message"
)

var ConsensusPid *actor.PID

//func PushConsensus(cons *message.ConsensusPayload){
//	ConsensusPid.Tell(cons)
//}

func SetConsensusPid(conPid * actor.PID){
	ConsensusPid = conPid
}