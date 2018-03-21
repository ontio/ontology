package account

import (
	. "github.com/Ontology/common"
	ct "github.com/Ontology/core/contract"
)

type IClientStore interface {
	BuildDatabase(path string)

	SaveStoredData(name string, value []byte)

	LoadStoredData(name string) []byte

	LoadAccount() map[Address]*Account

	LoadContracts() map[Address]*ct.Contract
}
