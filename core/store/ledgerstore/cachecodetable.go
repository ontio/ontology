package ledgerstore

import (
	"fmt"
	"github.com/Ontology/core/states"
	."github.com/Ontology/core/store/common"
)

type CacheCodeTable struct {
	store IStateStore
}

func (table *CacheCodeTable) GetCode(codeHash []byte) ([]byte, error) {
	value, _ := table.store.TryGet(ST_Contract, codeHash)
	if value == nil {
		return nil, fmt.Errorf("[GetCode] TryGet contract error! codeHash:%x", codeHash)
	}

	return value.Value.(*states.ContractState).Code.Code, nil
}
