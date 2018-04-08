package req

import (
	"github.com/Ontology/eventbus/actor"
)

var ConsensusPid *actor.PID

func SetConsensusPid(conPid *actor.PID) {
	ConsensusPid = conPid
}
