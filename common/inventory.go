package common

type InventoryType byte

const (
	TRANSACTION	InventoryType = 0x01
	BLOCK		InventoryType = 0x02
	CONSENSUS	InventoryType = 0xe0
)

//TODO: temp inventory
type Inventory interface {
	//sig.SignableData
	Hash() Uint256
	Verify() error
	Type() InventoryType
}

