package ChainStore

import (
	"github.com/Ontology/core/store"
	"github.com/Ontology/errors"
	"github.com/Ontology/core/states"
)

type CacheCodeTable struct {
	store store.IStateStore
}

func (table *CacheCodeTable) GetCode(codeHash []byte) ([]byte, error) {
	value, err := table.store.TryGet(store.ST_Contract, codeHash)
	if err != nil {
		return nil, errors.NewErr("[GetCode] TryGet contract error!")
	}
	return value.Value.(*states.ContractState).Code.Code, nil
}
