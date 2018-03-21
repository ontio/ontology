package service

import (
	"github.com/Ontology/common"
)

type StorageContext struct {
	codeHash common.Address
}

func NewStorageContext(codeHash common.Address) *StorageContext {
	var storageContext StorageContext
	storageContext.codeHash = codeHash
	return &storageContext
}

func (sc *StorageContext) ToArray() []byte {
	return sc.codeHash.ToArray()
}

