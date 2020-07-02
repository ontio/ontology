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
package ontid

import (
	"errors"
	"fmt"
	"io"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

const (
	MAX_KEY_SIZE   = 80
	MAX_TYPE_SIZE  = 64
	MAX_VALUE_SIZE = 512 * 1024

	MAX_NUM = 100
)

type attributeJson struct {
	Key   string `json:"key"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

type attribute struct {
	key       []byte
	valueType []byte
	value     []byte
}

func (this *attribute) Value() []byte {
	sink := common.NewZeroCopySink(nil)
	sink.WriteVarBytes(this.value)
	sink.WriteVarBytes(this.valueType)
	return sink.Bytes()
}

func (this *attribute) SetValue(data []byte) error {
	source := common.NewZeroCopySource(data)
	val, _, irregular, eof := source.NextVarBytes()
	if irregular {
		return common.ErrIrregularData
	}
	if eof {
		return io.ErrUnexpectedEOF
	}

	vt, _, irregular, eof := source.NextVarBytes()
	if irregular {
		return common.ErrIrregularData
	}
	if eof {
		return io.ErrUnexpectedEOF
	}

	this.valueType = vt
	this.value = val
	return nil
}

func (this *attribute) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes(this.key)
	sink.WriteVarBytes(this.valueType)
	sink.WriteVarBytes(this.value)
}

func (this *attribute) Deserialization(source *common.ZeroCopySource) error {
	k, _, irregular, eof := source.NextVarBytes()
	if irregular {
		return common.ErrIrregularData
	}
	if eof {
		return io.ErrUnexpectedEOF
	}
	if len(k) > MAX_KEY_SIZE {
		return fmt.Errorf("key size %d is too large, max limit is %d", len(k), MAX_KEY_SIZE)
	}

	vt, _, irregular, eof := source.NextVarBytes()
	if irregular {
		return common.ErrIrregularData
	}
	if eof {
		return io.ErrUnexpectedEOF
	}
	if len(vt) > MAX_TYPE_SIZE {
		return fmt.Errorf("type size %d is too large, max limit is %d", len(vt), MAX_TYPE_SIZE)
	}

	v, _, irregular, eof := source.NextVarBytes()
	if irregular {
		return common.ErrIrregularData
	}
	if eof {
		return io.ErrUnexpectedEOF
	}
	if len(v) > MAX_VALUE_SIZE {
		return fmt.Errorf("value size %d is too large, max limit is %d", len(v), MAX_VALUE_SIZE)
	}

	this.key = k
	this.value = v
	this.valueType = vt
	return nil
}

func insertOrUpdateAttr(srvc *native.NativeService, encId []byte, attr *attribute) error {
	key := append(encId, FIELD_ATTR)
	val := attr.Value()
	err := utils.LinkedlistInsert(srvc, key, attr.key, val)
	if err != nil {
		return errors.New("store attribute error: " + err.Error())
	}
	return nil
}

func findAttr(srvc *native.NativeService, encId, item []byte) (*utils.LinkedlistNode, error) {
	key := append(encId, FIELD_ATTR)
	return utils.LinkedlistGetItem(srvc, key, item)
}

func batchInsertAttr(srvc *native.NativeService, encId []byte, attr []attribute) error {
	for i, v := range attr {
		err := insertOrUpdateAttr(srvc, encId, &v)
		if err != nil {
			return fmt.Errorf("store attribute %d error: %s", i, err)
		}
	}

	key := append(encId, FIELD_ATTR)
	n, err := utils.LinkedlistGetNumOfItems(srvc, key)
	if err != nil {
		return err
	}
	if n > MAX_NUM {
		return fmt.Errorf("too many attributes, max is %d", MAX_NUM)
	}

	return nil
}

func deleteAttr(srvc *native.NativeService, encId, path []byte) error {
	key := append(encId, FIELD_ATTR)
	ok, err := utils.LinkedlistDelete(srvc, key, path)
	if err != nil {
		return err
	} else if !ok {
		return errors.New("attribute not exist")
	}
	return nil
}

func getAllAttr(srvc *native.NativeService, encId []byte) ([]byte, error) {
	key := append(encId, FIELD_ATTR)
	item, err := utils.LinkedlistGetHead(srvc, key)
	if err != nil {
		return nil, fmt.Errorf("get list head error, %s", err)
	} else if len(item) == 0 {
		// not exists
		return nil, nil
	}

	res := common.NewZeroCopySink(nil)
	var i uint16 = 0
	for len(item) > 0 {
		node, err := utils.LinkedlistGetItem(srvc, key, item)
		if err != nil {
			return nil, fmt.Errorf("get storage item error, %s", err)
		} else if node == nil {
			return nil, fmt.Errorf("storage item not exists, %v", item)
		}

		var attr attribute
		err = attr.SetValue(node.GetPayload())
		if err != nil {
			return nil, fmt.Errorf("parse attribute failed, %s", err)
		}
		attr.key = item
		attr.Serialization(res)

		i += 1
		item = node.GetNext()
	}
	return res.Bytes(), nil
}

func getAttrByKey(srvc *native.NativeService, encId []byte, item []byte) ([]byte, error) {
	key := append(encId, FIELD_ATTR)

	res := common.NewZeroCopySink(nil)

	node, err := utils.LinkedlistGetItem(srvc, key, item)
	if err != nil {
		return nil, fmt.Errorf("get storage item error, %s", err)
	} else if node == nil {
		return nil, nil
	}

	var attr attribute
	err = attr.SetValue(node.GetPayload())
	if err != nil {
		return nil, fmt.Errorf("parse attribute failed, %s", err)
	}
	attr.key = item
	attr.Serialization(res)

	return res.Bytes(), nil
}

func getAllAttrJson(srvc *native.NativeService, encId []byte) ([]*attributeJson, error) {
	key := append(encId, FIELD_ATTR)
	item, err := utils.LinkedlistGetHead(srvc, key)
	if err != nil {
		return nil, fmt.Errorf("get list head error, %s", err)
	} else if len(item) == 0 {
		// not exists
		return nil, nil
	}

	res := make([]*attributeJson, 0)
	var i uint16 = 0
	for len(item) > 0 {
		node, err := utils.LinkedlistGetItem(srvc, key, item)
		if err != nil {
			return nil, fmt.Errorf("get storage item error, %s", err)
		} else if node == nil {
			return nil, fmt.Errorf("storage item not exists, %v", item)
		}

		var attr attribute
		attrJson := new(attributeJson)
		err = attr.SetValue(node.GetPayload())
		if err != nil {
			return nil, fmt.Errorf("parse attribute failed, %s", err)
		}
		attr.key = item
		ontId, err := decodeID(encId)
		if err != nil {
			return nil, fmt.Errorf("decodeID failed, %s", err)
		}
		attrJson.Key = fmt.Sprintf("%s#%s", string(ontId), string(attr.key))
		attrJson.Type = string(attr.valueType)
		attrJson.Value = string(attr.value)
		res = append(res, attrJson)

		i += 1
		item = node.GetNext()
	}
	return res, nil
}

func getAttrKeys(attr []attribute) [][]byte {
	var paths = make([][]byte, 0)
	for _, v := range attr {
		paths = append(paths, v.key)
	}
	return paths
}

func deleteAllAttr(srvc *native.NativeService, encId []byte) error {
	key := append(encId, FIELD_ATTR)
	return utils.LinkedlistDeleteAll(srvc, key)
}
