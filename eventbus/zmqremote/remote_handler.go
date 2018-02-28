package zmqremote

import "github.com/Ontology/eventbus/actor"

func remoteHandler(pid *actor.PID) (actor.Process, bool) {
	ref := newProcess(pid)
	return ref, true
}
