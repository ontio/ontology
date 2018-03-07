package actor

import "github.com/Ontology/core/types"

type StartConsensus struct{}
type StopConsensus struct{}

//internal Message
type TimeOut struct{}
type BlockCompleted struct {
	Block *types.Block
}
