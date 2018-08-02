package store

import "encoding/binary"

const (
	DEFAULT_WALLET_NAME = "MyWallet"
	WALLET_VERSION      = "1.1"
	WALLET_INIT_DATA    = "walletStore"
)

const (
	WALLET_INIT_PREFIX               = 0x00
	WALLET_NAME_PREFIX               = 0x01
	WALLET_VERSION_PREFIX            = 0x02
	WALLET_SCRYPT_PREFIX             = 0x03
	WALLET_NEXT_ACCOUNT_INDEX_PREFIX = 0x04
	WALLET_ACCOUNT_INDEX_PREFIX      = 0x05
	WALLET_ACCOUNT_PREFIX            = 0x06
	WALLET_EXTRA_PREFIX              = 0x07
)

func GetWalletInitKey() []byte {
	return []byte{WALLET_INIT_PREFIX}
}

func GetWalletNameKey() []byte {
	return []byte{WALLET_NAME_PREFIX}
}

func GetWalletVersionKey() []byte {
	return []byte{WALLET_VERSION_PREFIX}
}

func GetWalletScryptKey() []byte {
	return []byte{WALLET_SCRYPT_PREFIX}
}

func GetAccountIndexKey(index uint32) []byte {
	data := make([]byte, 4, 4)
	binary.LittleEndian.PutUint32(data, index)
	return append([]byte{WALLET_ACCOUNT_INDEX_PREFIX}, data...)
}

func GetNextAccountIndexKey() []byte {
	return []byte{WALLET_NEXT_ACCOUNT_INDEX_PREFIX}
}

func GetAccountKey(address string) []byte {
	return append([]byte{WALLET_ACCOUNT_PREFIX}, []byte(address)...)
}

func GetWalletExtraKey() []byte {
	return []byte{WALLET_EXTRA_PREFIX}
}
