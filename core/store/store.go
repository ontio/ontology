package store

type IIterator interface {
	Next() bool
	Prev() bool
	First() bool
	Last() bool
	Seek(key []byte) bool
	Key() []byte
	Value() []byte
	Release()
}

type IStore interface {
	Put(key []byte, value []byte) error
	Get(key []byte) ([]byte, error)
	Delete(key []byte) error
	NewBatch() error
	BatchPut(key []byte, value []byte) error
	BatchDelete(key []byte) error
	BatchCommit() error
	Close() error
	NewIterator(prefix []byte) IIterator
}
