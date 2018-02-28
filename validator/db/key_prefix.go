package db

import (
	pool "github.com/valyala/bytebufferpool"

	"github.com/Ontology/common"
)

// DataEntryPrefix
type KeyPrefix byte

const (
	//SYSTEM
	SYS_Version      KeyPrefix = 0
	SYS_GenesisBlock KeyPrefix = 1 // key: prefix, value: gensisBlock

	SYS_BestBlock       KeyPrefix = 2 // key : prefix, value: bestblock
	SYS_BestBlockHeader KeyPrefix = 3 // key: prefix, value: BlockHeader

	// DATA
	//DATA_Block KeyPrefix = iota
	//DATA_Header
	DATA_Transaction KeyPrefix = 10 // key: prefix+txid, value: height + tx

	TX_Meta KeyPrefix = 20 // key: TX_Meta + txid, value: height + spend bits

	// ASSET
	//ST_SpentCoin KeyPrefix = 30

	//ST_Account
	//ST_Coin
	//ST_BookKeeper
	//ST_Asset
	//ST_Contract
	//ST_Storage
	//ST_Identity
	//ST_Program_Coin
	//ST_Validator
	//ST_Vote
	//
	//IX_HeaderHashList
)

func GenGenesisBlockKey() *pool.ByteBuffer {
	key := keyPool.Get()
	key.WriteByte(byte(SYS_GenesisBlock))
	return key
}

func GenBestBlockHeaderKey() *pool.ByteBuffer {
	key := keyPool.Get()
	key.WriteByte(byte(SYS_BestBlockHeader))
	return key
}

func GenDataTransactionKey(hash common.Uint256) *pool.ByteBuffer {
	key := keyPool.Get()
	key.WriteByte(byte(DATA_Transaction))
	key.Write(hash.ToArray())
	return key
}

func GenTxMetaKey(hash common.Uint256) *pool.ByteBuffer {
	key := keyPool.Get()
	key.WriteByte(byte(TX_Meta))
	key.Write(hash.ToArray())

	return key
}
