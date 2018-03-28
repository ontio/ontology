/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package merkle

import (
	"errors"
	"io"
	"os"

	"github.com/Ontology/common"
)

type HashStore interface {
	Append(hash []common.Uint256) error
	Flush() error
	Close()
	GetHash(pos uint32) (common.Uint256, error)
}

type FileHashStore struct {
	file_name string
	file      *os.File
}

func NewFileHashStore(name string, tree_size uint32) (*FileHashStore, error) {
	f, err := os.OpenFile(name, os.O_RDWR | os.O_CREATE, 0755)
	if err != nil {
		return nil, err
	}
	store := &FileHashStore{
		file_name: name,
		file:      f,
	}

	err = store.checkConsistence(tree_size)
	if err != nil {
		return nil, err
	}

	num_hashes := getStoredHashNum(tree_size)
	size := int64(num_hashes) * int64(common.UINT256_SIZE)

	_, err = store.file.Seek(size, io.SeekStart)
	if err != nil {
		return nil, err
	}
	return store, nil
}

func getStoredHashNum(tree_size uint32) int64 {
	subtreesize := getSubTreeSize(tree_size)
	sum := int64(0)
	for _, v := range subtreesize {
		sum += int64(v)
	}

	return sum
}

func (self *FileHashStore) checkConsistence(tree_size uint32) error {
	num_hashes := getStoredHashNum(tree_size)

	stat, err := self.file.Stat()
	if err != nil {
		return err
	} else if stat.Size() < int64(num_hashes) * int64(common.UINT256_SIZE) {
		return errors.New("stored hashes are less than expected")
	}

	return nil
}

func (self *FileHashStore) Append(hash []common.Uint256) error {
	if self == nil {
		return nil
	}
	buf := make([]byte, 0, len(hash) *common.UINT256_SIZE)
	for _, h := range hash {
		buf = append(buf, h[:]...)
	}
	_, err := self.file.Write(buf)
	return err
}

func (self *FileHashStore) Flush() error {
	if self == nil {
		return nil
	}
	return self.file.Sync()
}

func (self *FileHashStore) Close() {
	if self == nil {
		return
	}
	self.file.Close()
}

func (self *FileHashStore) GetHash(pos uint32) (common.Uint256, error) {
	if self == nil {
		return EMPTY_HASH, errors.New("FileHashstore is nil")
	}
	hash := EMPTY_HASH
	_, err := self.file.ReadAt(hash[:], int64(pos) * int64(common.UINT256_SIZE))
	if err != nil {
		return EMPTY_HASH, err
	}

	return hash, nil
}

type MemHashStore struct {
	hashes []common.Uint256
}

func (self *MemHashStore) Append(hash []common.Uint256) error {
	self.hashes = append(self.hashes, hash...)
	return nil
}

func (self *MemHashStore) GetHash(pos uint32) (common.Uint256, error) {
	return self.hashes[pos], nil
}

func (self *MemHashStore) Flush() error {
	return nil
}

func (self *MemHashStore) Close() {}
