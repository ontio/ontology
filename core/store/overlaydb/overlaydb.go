package overlaydb

import (
	"github.com/ontio/ontology/core/store/common"
)

type OverlayDB struct {
	store      common.PersistStore
	memdb      *MemDB
	keyScratch []byte
	dbErr error
}

func NewOverlayDB(store common.PersistStore) *OverlayDB {
	return &OverlayDB{
		store: store,
		memdb: NewMemDB(0),
	}
}

func makePrefixedKey(dst []byte, prefix byte, key []byte) []byte {
	dst = ensureBuffer(dst, len(key)+1)
	dst[0] = prefix
	copy(dst[1:], key)
	return dst
}

func ensureBuffer(b []byte, n int) []byte {
	if cap(b) < n {
		return make([]byte, n)
	}
	return b[:n]
}

// if key is deleted, value == nil
func (self *OverlayDB) Get(prefix byte, key []byte) (value []byte, err error) {
	var unknown bool
	self.keyScratch = makePrefixedKey(self.keyScratch, prefix, key)
	value, unknown = self.memdb.Get(self.keyScratch)
	if unknown == false {
		return value, nil
	}

	value, err = self.store.Get(self.keyScratch)
	if err != nil {
		if err == common.ErrNotFound {
			return nil, nil
		}
		self.dbErr = err
		return nil, err
	}

	return
}

func (self *OverlayDB) Put(prefix byte, key []byte, value []byte) {
	self.keyScratch = makePrefixedKey(self.keyScratch, prefix, key)
	self.memdb.Put(self.keyScratch, value)
}

func (self *OverlayDB) Delete(prefix byte, key []byte) {
	self.keyScratch = makePrefixedKey(self.keyScratch, prefix, key)
	self.memdb.Delete(self.keyScratch)
}

type Iter struct  {
	backend common.StoreIterator
	memdb common.StoreIterator
	key, value []byte
}

func (iter *Iter)First() bool {
	var bkey, bval,mkey, mval []byte
	back := iter.backend.First()
	mem := iter.memdb.First()
	if back {
		bkey = iter.backend.Key()
		bval = iter.backend.Value()
		if mem == false {
			iter.key = bkey
			iter.value = bval
			return true
		}
	}
	if mem {
		mkey = iter.memdb.Key()
		mval = iter.memdb.Value()
	}

}

/*
func (self *OverlayDB)NewIterator(prefix byte, key []byte) common.StoreIterator {

}
*/


/*
func (self *OverlayDB) Find(prefix common.DataEntryPrefix, key []byte) ([]*common.StateItem, error) {
	stats, err := self.store.Find(prefix, key)
	if err != nil {
		return nil, err
	}
	var sts []*common.StateItem
	pkey := append([]byte{byte(prefix)}, key...)

	index := make(map[string]int)
	for i, v := range stats {
		index[v.Key] = i
	}

	deleted := make([]int, 0)
	for k, v := range self.memdb {
		if strings.HasPrefix(k, string(pkey)) {
			if v == nil { // deleted but in inner db, need remove
				if i, ok := index[k]; ok {
					deleted = append(deleted, i)
				}
			} else {
				if i, ok := index[k]; ok {
					sts[i] = &common.StateItem{Key: k, Value: v}
				} else {
					sts = append(sts, &common.StateItem{Key: k, Value: v})
				}
			}
		}
	}

	sort.Ints(deleted)
	for i := len(deleted) - 1; i >= 0; i-- {
		sts = append(sts[:deleted[i]], sts[deleted[i]+1:]...)
	}

	return sts, nil
}

func (self *OverlayDB) TryAdd(prefix common.DataEntryPrefix, key []byte, value states.StateValue) {
	pkey := append([]byte{byte(prefix)}, key...)
	self.memdb[string(pkey)] = value
}

func (self *OverlayDB) TryGet(prefix common.DataEntryPrefix, key []byte) (states.StateValue, error) {
	pkey := append([]byte{byte(prefix)}, key...)
	if state, ok := self.memdb[string(pkey)]; ok {
		return state, nil
	}
	return self.store.TryGet(prefix, key)
}

func (self *OverlayDB) TryDelete(prefix common.DataEntryPrefix, key []byte) {
	pkey := append([]byte{byte(prefix)}, key...)
	self.memdb[string(pkey)] = nil
}

func (self *OverlayDB) CommitTo() error {
	for k, v := range self.memdb {
		pkey := []byte(k)
		self.store.TryAdd(common.DataEntryPrefix(pkey[0]), pkey[1:], v)
	}
	return nil
}
*/
