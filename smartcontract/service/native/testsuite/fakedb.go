package testsuite

import (
	"github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/store/overlaydb"
)

type MockDB struct {
	common.PersistStore
	db map[string]string
}

func (self *MockDB) Get(key []byte) ([]byte, error) {
	val, ok := self.db[string(key)]
	if ok == false {
		return nil, common.ErrNotFound
	}
	return []byte(val), nil
}

func (self *MockDB) BatchPut(key []byte, value []byte) {
	self.db[string(key)] = string(value)
}

func (self *MockDB) BatchDelete(key []byte) {
	delete(self.db, string(key))
}

func NewOverlayDB() *overlaydb.OverlayDB {
	return overlaydb.NewOverlayDB(&MockDB{nil, make(map[string]string)})
}
