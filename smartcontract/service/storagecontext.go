package service

import (
	"github.com/Ontology/common"
)

type StorageContext struct {
	codeHash common.Uint160
}

func NewStorageContext(codeHash common.Uint160) *StorageContext {
	var storageContext StorageContext
	storageContext.codeHash = codeHash
	return &storageContext
}

func (sc *StorageContext) ToArray() []byte {
	return sc.codeHash.ToArray()
}
