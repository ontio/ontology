package req

import (
	"github.com/Ontology/eventbus/actor"
	//"github.com/Ontology/net/message"
)

var ConsensusPid *actor.PID

func SetConsensusPid(conPid *actor.PID) {
	ConsensusPid = conPid
}
