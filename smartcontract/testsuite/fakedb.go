package testsuite

import (
	"github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/store/overlaydb"
)

type FakeDB struct {
	common.PersistStore
}

func (self *FakeDB) Get(key []byte) ([]byte, error) {
	return nil, common.ErrNotFound
}

func NewOverlayDB() *overlaydb.OverlayDB {
	return overlaydb.NewOverlayDB(&FakeDB{nil})
}

