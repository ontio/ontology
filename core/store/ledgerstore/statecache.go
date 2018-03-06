package ledgerstore

import (
	"github.com/hashicorp/golang-lru"
	"github.com/Ontology/core/states"
)

const(
	StateCacheSize = 100000
)

type StateCache struct {
	stateCache *lru.ARCCache
}

func NewStateCache() (*StateCache, error){
	stateCache, err := lru.NewARC(StateCacheSize)
	if err != nil {
		return nil, err
	}
	return &StateCache{
		stateCache:stateCache,
	}, nil
}

func (this *StateCache) GetState(key []byte)states.IStateValue{
	state, ok := this.stateCache.Get(string(key))
	if !ok {
		return nil
	}
	return state.(states.IStateValue)
}

func (this *StateCache) AddState(key []byte, state states.IStateValue){
	this.stateCache.Add(string(key), state)
}

func (this *StateCache) DeleteState(key []byte){
	this.stateCache.Remove(string(key))
}
