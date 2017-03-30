package client

import (
	ct "github.com/DNAProject/DNA/core/contract"
	. "github.com/DNAProject/DNA/common"
)

type IClientStore interface {
	BuildDatabase(path string)

	SaveStoredData(name string,value []byte)

	LoadStoredData(name string) []byte

	LoadAccount()  map[Uint160]*Account

	LoadContracts() map[Uint160]*ct.Contract
}
