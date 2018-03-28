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

package db

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

// 第一个为长度
type FixedBitMap struct {
	len   uint32
	value []byte
	data  []byte // first 4bytes used to store len to avoid reallocate mem when put to db
}

// 创建指定初始化大小的bitSet
func NewFixedBitMap(nbits uint32) *FixedBitMap {
	nbyte := (nbits+7)/8 + 4
	data := make([]byte, nbyte, nbyte)
	binary.BigEndian.PutUint32(data, nbits)
	return &FixedBitMap{
		len:   nbits,
		value: data[4:],
		data:  data,
	}
}

func (self *FixedBitMap) Serialize(w io.Writer) error {
	_, err := w.Write(self.data)
	return err
}

func (self *FixedBitMap) Deserialize(r io.Reader) error {
	err := binary.Read(r, binary.BigEndian, &self.len)
	if err != nil {
		return nil
	}

	nbyte := (self.len+7)/8 + 4
	data := make([]byte, nbyte, nbyte)
	binary.BigEndian.PutUint32(data, self.len)
	n, err := r.Read(data[4:])
	if uint32(n) != nbyte-4 || err != nil {
		return errors.New("[FixedBitMap] Deserialize failed: wrong bytes len")
	}
	self.data = data
	self.value = data[4:]

	return nil
}

// 把指定位置设为ture
func (self *FixedBitMap) Set(bitIndex uint32) {
	if bitIndex >= self.len {
		panic("[FixedBitMap] Set index out of range")
	}
	pos := bitIndex / 8
	self.value[pos] |= byte(0x01) << (bitIndex % 8)
}

// 设置指定位置为false
func (self *FixedBitMap) Unset(bitIndex uint32) {
	if bitIndex >= self.len {
		panic("[FixedBitMap] Set index out of range")
	}
	pos := bitIndex / 8
	self.value[pos] &^= byte(0x01) << (bitIndex % 8)
}

func (self *FixedBitMap) boundcheck(bitIndex uint32) {
	if bitIndex >= self.len {
		panic("[FixedBitMap] Set index out of range")
	}
}

// 获取指定位置的值
func (self *FixedBitMap) Get(bitIndex uint32) bool {
	self.boundcheck(bitIndex)
	pos := bitIndex / 8
	return self.value[pos]&(byte(0x01)<<(bitIndex%8)) != 0
}

func (self *FixedBitMap) IsFullSet() bool {
	size := self.len / 8
	for i := uint32(0); i < size; i++ {
		if self.value[i] != 255 {
			return false
		}
	}
	delta := self.len - size*8
	mask := byte(1<<delta - 1)
	return mask&self.value[size] == mask
}

// 以二进制串的格式打印bitMap内容
func (self *FixedBitMap) ToString() string {
	strAppend := &bytes.Buffer{}
	for i := uint32(0); i < self.len; i++ {
		temp := self.value[i]
		for j := 0; j < 64; j++ {
			if temp&(byte(0x01)<<byte(j)) != 0 {
				strAppend.WriteString("1")
			} else {
				strAppend.WriteString("0")
			}
		}
	}
	return strAppend.String()
}
