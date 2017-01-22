package message

import (
	. "GoOnchain/common"
	sig "GoOnchain/core/signature"
)

type InventoryType byte

const (
	Transaction InventoryType = 0x01
	Block InventoryType = 0x02
	Consensus InventoryType = 0xe0
)

//TODO: temp inventory
type Inventory interface {
	sig.SignableData
	Hash() Uint256
	Verify() error
	InvertoryType() InventoryType
}
