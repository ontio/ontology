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
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"io"
)

type attribute struct {
	key       []byte
	value     []byte
	valueType []byte
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

	vt, _, irregular, eof := source.NextVarBytes()
	if irregular {
		return common.ErrIrregularData
	}
	if eof {
		return io.ErrUnexpectedEOF
	}

	v, _, irregular, eof := source.NextVarBytes()
	if irregular {
		return common.ErrIrregularData
	}
	if eof {
		return io.ErrUnexpectedEOF
	}
	this.key = k
	this.value = v
	this.valueType = vt
	return nil
}

func insertOrUpdateAttr(srvc *native.NativeService, encID []byte, attr *attribute) error {
	key := append(encID, FIELD_ATTR)
	val := attr.Value()
	err := utils.LinkedlistInsert(srvc, key, attr.key, val)
	if err != nil {
		return errors.New("store attribute error: " + err.Error())
	}
	return nil
}

func findAttr(srvc *native.NativeService, encID, item []byte) (*utils.LinkedlistNode, error) {
	key := append(encID, FIELD_ATTR)
	return utils.LinkedlistGetItem(srvc, key, item)
}

func batchInsertAttr(srvc *native.NativeService, encID []byte, attr []attribute) error {
	res := make([][]byte, len(attr))
	for i, v := range attr {
		err := insertOrUpdateAttr(srvc, encID, &v)
		if err != nil {
			return errors.New("store attributes error: " + err.Error())
		}
		res[i] = v.key
	}

	return nil
}

func getAllAttr(srvc *native.NativeService, encID []byte) ([]byte, error) {
	key := append(encID, FIELD_ATTR)
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
