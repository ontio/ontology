package merkle

import (
	"os"
)

type HashStore interface {
	Append(hash []Uint256) error
	Flush() error
	GetHash(pos uint32) (Uint256, error)
}

type FileHashStore struct {
	file_name string
	file      *os.File
}

func NewFileHashStore(name string) *FileHashStore {
	f, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return nil
	}
	return &FileHashStore{
		file_name: name,
		file:      f,
	}
}

func (self *FileHashStore) Append(hash []Uint256) error {
	buf := make([]byte, 0, len(hash)*UINT256SIZE)
	for _, h := range hash {
		buf = append(buf, h[:]...)
	}
	_, err := self.file.Write(buf)
	return err
}

func (self *FileHashStore) Flush() error {
	return self.file.Sync()
}

func (self *FileHashStore) GetHash(pos uint32) (Uint256, error) {
	hash := EMPTY_HASH
	_, err := self.file.ReadAt(hash[:], int64(pos)*int64(UINT256SIZE))
	if err != nil {
		return EMPTY_HASH, err
	}

	return hash, nil
}

type MemHashStore struct {
	hashes []Uint256
}

func (self *MemHashStore) Append(hash []Uint256) error {
	self.hashes = append(self.hashes, hash...)
	return nil
}

func (self *MemHashStore) GetHash(pos uint32) (Uint256, error) {
	return self.hashes[pos], nil
}

func (self *MemHashStore) Flush() error {
	return nil
}
