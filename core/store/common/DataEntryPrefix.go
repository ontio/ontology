package common

// DataEntryPrefix
type DataEntryPrefix byte

const (
	// DATA
	DATA_Block DataEntryPrefix = iota
	DATA_Header = 0x01
	DATA_Transaction = 0x02

	// Transaction
	ST_BookKeeper = 0x03
	ST_Contract = 0x04
	ST_Storage = 0x05
	ST_Validator = 0x07
	ST_Vote = 0x08

	IX_HeaderHashList = 0x09

	//SYSTEM
	SYS_CurrentBlock = 0x10
	SYS_Version = 0x11
	SYS_CurrentStateRoot = 0x12
	SYS_BlockMerkleTree = 0x13

	EVENT_Notify = 0x14
)
