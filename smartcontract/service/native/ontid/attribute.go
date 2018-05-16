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
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

type attribute struct {
	key       []byte
	value     []byte
	valueType []byte
}

func (this *attribute) Value() ([]byte, error) {
	var buf bytes.Buffer
	err := serialization.WriteVarBytes(&buf, this.value)
	if err != nil {
		return nil, err
	}
	err = serialization.WriteVarBytes(&buf, this.valueType)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (this *attribute) SetValue(data []byte) error {
	buf := bytes.NewBuffer(data)
	val, err := serialization.ReadVarBytes(buf)
	if err != nil {
		return err
	}
	vt, err := serialization.ReadVarBytes(buf)
	if err != nil {
		return err
	}
	this.valueType = vt
	this.value = val
	return nil
}

func (this *attribute) Serialize(w io.Writer) error {
	err := serialization.WriteVarBytes(w, this.key)
	if err != nil {
		return err
	}
	err = serialization.WriteVarBytes(w, this.valueType)
	if err != nil {
		return err
	}
	err = serialization.WriteVarBytes(w, this.value)
	if err != nil {
		return err
	}
	return nil
}

func (this *attribute) Deserialize(r io.Reader) error {
	k, err := serialization.ReadVarBytes(r)
	if err != nil {
		return err
	}
	vt, err := serialization.ReadVarBytes(r)
	if err != nil {
		return err
	}
	v, err := serialization.ReadVarBytes(r)
	if err != nil {
		return err
	}
	this.key = k
	this.value = v
	this.valueType = vt
	return nil
}

func insertOrUpdateAttr(srvc *native.NativeService, encID []byte, attr *attribute) error {
	key := append(encID, FIELD_ATTR)
	val, err := attr.Value()
	if err != nil {
		return errors.New("serialize attribute value error: " + err.Error())
	}
	err = utils.LinkedlistInsert(srvc, key, attr.key, val)
	if err != nil {
		return errors.New("store attribute error: " + err.Error())
	}
	return nil
}

func findAttr(srvc *native.NativeService, encID, item []byte) (*utils.LinkedlistNode, error) {
	key := append(encID, FIELD_ATTR)
	return utils.LinkedlistGetItem(srvc, key, item)
}

func batchInsertAttr(srvc *native.NativeService, encID, data []byte) error {
	// parse attributes
	buf := bytes.NewBuffer(data)
	attr := make([]*attribute, 0)
	for buf.Len() > 0 {
		t := new(attribute)
		err := t.Deserialize(buf)
		if err != nil {
			return errors.New("parse attribute error: " + err.Error())
		}
		attr = append(attr, t)
	}
	for _, v := range attr {
		err := insertOrUpdateAttr(srvc, encID, v)
		if err != nil {
			return errors.New("store attributes error: " + err.Error())
		}
	}
	return nil
}

func getAllAttr(srvc *native.NativeService, encID []byte) ([]byte, error) {
	key := append(encID, FIELD_ATTR)
	item, err := utils.LinkedlistGetHead(srvc, key)
	if err != nil {
		return nil, fmt.Errorf("get list head error, %s", err)
	} else if len(item) == 0 {
		return nil, errors.New("cannot get list head")
	}

	var res bytes.Buffer
	var i uint16 = 0
	for len(item) > 0 {
		node, err := utils.LinkedlistGetItem(srvc, key, item)
		if err != nil {
			return nil, fmt.Errorf("get storage item error, %s", err)
		} else if node == nil {
			return nil, fmt.Errorf("storage item not exists")
		}

		var attr attribute
		err = attr.SetValue(node.GetPayload())
		if err != nil {
			return nil, fmt.Errorf("parse attribute failed, %s", err)
		}
		attr.key = item
		err = attr.Serialize(&res)
		if err != nil {
			return nil, fmt.Errorf("serialize error, %s", err)
		}

		i += 1
		item = node.GetNext()
	}
	return res.Bytes(), nil
}
