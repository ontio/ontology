package store

// DataEntryPrefix
type DataEntryPrefix byte

const (
	// DATA
	DATA_Block DataEntryPrefix = iota
	DATA_Header
	DATA_Transaction

	// ASSET
	ST_Account
	ST_Coin
	ST_SpentCoin
	ST_BookKeeper
	ST_Asset
	ST_Contract
	ST_Storage
	ST_Identity
	ST_Program_Coin
	ST_Validator
	ST_Vote

	IX_HeaderHashList

	//SYSTEM
	SYS_CurrentBlock
	SYS_Version
	SYS_CurrentStateRoot
	SYS_BlockMerkleTree
	SYS_MPTTrie
)
