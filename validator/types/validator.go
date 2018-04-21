package types

import (
	"github.com/ontio/ontology-eventbus/actor"
)

// Validator wraps validator actor's pid
type Validator interface {
	// Register send a register message to poolId
	Register(poolId *actor.PID)
	// UnRegister send an unregister message to poolId
	UnRegister(poolId *actor.PID)
	// VerifyType returns the type of validator
	VerifyType() VerifyType
}

type ValidatorActor struct {
	Pid       *actor.PID
	Id        string
}