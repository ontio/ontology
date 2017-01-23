package client

import (
	ct "GoOnchain/core/contract"
	. "GoOnchain/common"
)

type ClientStore interface {

	SaveStoredData(name string,value []byte)

	LoadStoredData(name string) []byte

	LoadAccount()  map[Uint160]*Account

	LoadContracts() map[Uint160]*ct.Contract
}
