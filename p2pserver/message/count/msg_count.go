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

package msgcount

import (
	"encoding/json"
	"sync"
)

//MsgCount represent the count of received message in type
type MsgCount struct {
	sync.RWMutex
	countMap map[string]uint32
}

//IncreCount increase 1 of the message type
func (this *MsgCount) IncreCount(msgType string) {
	this.Lock()
	defer this.Unlock()
	if this.countMap == nil {
		this.countMap = make(map[string]uint32, 0)
	}
	if cnt, ok := this.countMap[msgType]; ok {
		this.countMap[msgType] = cnt + 1
	} else {
		this.countMap[msgType] = 1
	}
}

//GetCount get the total message count of the type
func (this *MsgCount) GetCount(msgType string) uint32 {
	this.RLock()
	defer this.RUnlock()
	return this.countMap[msgType]
}

//GetCount get the total message count of all the type
func (this *MsgCount) GetAllCount() uint32 {
	this.RLock()
	defer this.RUnlock()
	if this.countMap == nil {
		return 0
	}
	var cnt uint32
	for k := range this.countMap {
		cnt += this.countMap[k]
	}
	return cnt
}

//Clean clean all the message count
func (this *MsgCount) Clean() {
	this.Lock()
	defer this.Unlock()
	if this.countMap != nil {
		for k := range this.countMap {
			delete(this.countMap, k)
		}
	}
}

//Copy copy one message count to another
func (this *MsgCount) Copy(c *MsgCount) {
	this.Lock()
	defer this.Unlock()
	if this.countMap == nil {
		this.countMap = make(map[string]uint32, 0)
	} else {
		for k := range this.countMap {
			delete(this.countMap, k)
		}
	}
	for k := range c.countMap {
		this.countMap[k] = c.countMap[k]
	}
}

//IsEmpty check if the map is empty
func (this *MsgCount) IsEmpty() bool {
	this.RLock()
	defer this.RUnlock()
	if this.countMap == nil {
		return true
	}
	for k := range this.countMap {
		if this.countMap[k] > 0 {
			return false
		}
	}
	return true
}

// Serialize custom string format
func (this *MsgCount) Serialization() ([]byte, error) {
	this.RLock()
	defer this.RUnlock()
	if this.countMap == nil {
		return []byte{}, nil
	}
	nonZeroMap := make(map[string]uint32, 0)
	for k := range this.countMap {
		if this.countMap[k] > 0 {
			nonZeroMap[k] = this.countMap[k]
		}
	}
	return json.Marshal(nonZeroMap)
}

// Deserialization deserialize the data
func (this *MsgCount) Deserialization(data []byte) error {
	this.Lock()
	defer this.Unlock()
	if len(data) == 0 {
		return nil
	}
	m := make(map[string]uint32)
	err := json.Unmarshal(data, &m)
	if err != nil {
		return err
	}
	this.countMap = m
	return nil
}
