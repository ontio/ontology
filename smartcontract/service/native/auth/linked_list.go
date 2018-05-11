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

package auth

import (
	"bytes"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/errors"
	. "github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

type LinkedlistNode struct {
	next    []byte
	prev    []byte
	payload []byte
}

func makeLinkedlistNode(next []byte, prev []byte, payload []byte) ([]byte, error) {
	node := &LinkedlistNode{next: next, prev: prev, payload: payload}
	node_bytes, err := node.Serialize()
	if err != nil {
		return nil, err
	}
	return node_bytes, nil
}
func (this *LinkedlistNode) Serialize() ([]byte, error) {
	bf := new(bytes.Buffer)
	if err := serialization.WriteVarBytes(bf, this.next); err != nil {
		return nil, err
	}
	if err := serialization.WriteVarBytes(bf, this.prev); err != nil {
		return nil, err
	}
	if err := serialization.WriteVarBytes(bf, this.payload); err != nil {
		return nil, err
	}
	return bf.Bytes(), nil
}

func (this *LinkedlistNode) Deserialize(r []byte) error {
	bf := bytes.NewReader(r)
	next, err := serialization.ReadVarBytes(bf)
	if err != nil {
		return err
	}
	prev, err := serialization.ReadVarBytes(bf)
	if err != nil {
		return err
	}
	payload, err := serialization.ReadVarBytes(bf)
	if err != nil {
		return err
	}
	this.next = next
	this.prev = prev
	this.payload = payload
	return nil
}

func GetListHead(native *NativeService, prefix []byte) ([]byte, error) {
	item, err := utils.GetStorageItem(native, prefix)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, nil
	}
	return item.Value, nil
}

func GetListNode(native *NativeService, prefix []byte, item []byte) (*LinkedlistNode, error) {
	node := new(LinkedlistNode)
	data, err := utils.GetStorageItem(native, append(prefix, item...))
	if err != nil {
		return nil, err
	}
	if data == nil || data.Value == nil || len(data.Value) == 0 {
		return nil, nil
	}
	err = node.Deserialize(data.Value)
	if err != nil {
		//log.Tracef("[prefix: %s, item: %s] error %s", hex.EncodeToString(prefix), hex.EncodeToString(item), err)
		return nil, err
	}
	return node, nil
}

func LinkedlistInsert(native *NativeService, prefix []byte, item []byte, payload []byte) error {
	null := []byte{}
	if item == nil {
		return errors.NewErr("item is nil")
	}
	head, err := GetListHead(native, prefix) //list head
	if err != nil {
		return err
	}
	q, err := GetListNode(native, prefix, item) //list node
	if err != nil {
		return err
	}

	if q != nil { //already exists
		node, err := makeLinkedlistNode(q.next, q.prev, payload)
		if err != nil {
			return err
		}
		PutBytes(native, append(prefix, item...), node) //update it
		return nil
	}
	if head == nil { //doubly-linked list contains zero element
		node, err := makeLinkedlistNode(null, null, payload)
		if err != nil {
			return err
		}
		PutBytes(native, append(prefix, item...), node) //item is the only element
		PutBytes(native, prefix, item)                  //item becomes head
	} else {
		null := []byte{}
		node, err := makeLinkedlistNode(head, null, payload)
		if err != nil {
			return err
		}
		PutBytes(native, append(prefix, item...), node) //item.next = head, item.prev = null,
		// item.payload = payload
		qhead, err := GetListNode(native, prefix, head)
		if err != nil {
			return err
		}

		node, err = makeLinkedlistNode(qhead.next, item, qhead.payload)
		if err != nil {
			return err
		}
		PutBytes(native, append(prefix, head...), node) //head.next = head.next, head.prev = item,
		// head.payload = head.payload
		PutBytes(native, prefix, item) // item becomes head
	}
	return nil
}

func LinkedlistDelete(native *NativeService, prefix []byte, item []byte) (bool, error) {
	null := []byte{}
	if item == nil {
		return false, errors.NewErr("item is nil")
	}
	q, err := GetListNode(native, prefix, item)
	if err != nil {
		return false, err
	}
	if q == nil {
		return false, nil
	}

	prev, next := q.prev, q.next
	if prev == nil {
		if next == nil {
			PutBytes(native, prefix, null) //clear linked list
		} else {
			qnext, err := GetListNode(native, prefix, next)
			if err != nil {
				return false, err
			}
			node, err := makeLinkedlistNode(qnext.next, null, qnext.payload) //qnext.next = qnext.next
			if err != nil {                                                  // qnext.prev = nil
				return false, err
			}
			PutBytes(native, append(prefix, next...), node)
			PutBytes(native, prefix, next) //next becomes head
		}
	} else {
		if next == nil {
			qprev, err := GetListNode(native, prefix, prev)
			if err != nil {
				return false, err
			}
			node, err := makeLinkedlistNode(null, qprev.prev, qprev.payload) //qprev becomes end
			if err != nil {
				return false, err
			}
			PutBytes(native, append(prefix, prev...), node)
		} else {
			qprev, err := GetListNode(native, prefix, prev)
			if err != nil {
				return false, err
			}
			qnext, err := GetListNode(native, prefix, next)
			if err != nil {
				return false, err
			}
			node_prev, err := makeLinkedlistNode(next, qprev.prev, qprev.payload) //
			if err != nil {
				return false, err
			}
			node_next, err := makeLinkedlistNode(qnext.next, prev, qnext.payload)
			if err != nil {
				return false, err
			}
			PutBytes(native, append(prefix, prev...), node_prev)
			PutBytes(native, append(prefix, next...), node_next)
		}
	}
	PutBytes(native, append(prefix, item...), null)
	return true, nil
}

func LinkedlistGetItem(native *NativeService, prefix []byte, item []byte) (*LinkedlistNode, error) {
	if item == nil {
		return nil, errors.NewErr("[linkedlist getNext] item is nil")
	}
	q, err := GetListNode(native, prefix, item)
	if err != nil {
		return nil, err
	}
	return q, nil
}

func LinkedlistGetHead(native *NativeService, prefix []byte) ([]byte, error) {
	head, err := GetListHead(native, prefix)
	if err != nil {
		return nil, err
	}
	return head, nil
}
func LinkedlistGetNumOfItems(native *NativeService, prefix []byte) (int, error) {
	n := 0
	head, err := GetListHead(native, prefix)
	if err != nil {
		return 0, err
	}
	q := head
	for q != nil {
		n += 1
		qnode, err := GetListNode(native, prefix, q)
		if err != nil {
			return 0, err
		}
		q = qnode.next
	}
	return n, nil
}

/*
func linkedlistTest(native *NativeService) error {
	contract1 := native.ContextRef.CurrentContext().ContractAddress
	//contract2 := []byte{ 0x01, 0x02 }

	{
		log.Trace("'empty linkedlist delete' test")
		//basic delete
		s, err := LinkedlistDelete(native, contract1[:], []byte{byte(11)})
		if err != nil {
			return err
		}
		if s {
			err := errors.NewErr("")
			return errors.NewDetailErr(err, errors.ErrNoCode, "[linked list basic test] delete nonexistent item returns true")
		}
	}

	{
		log.Trace("'basic insert' test")
		//basic insert
		for i := 0; i < 10; i++ {
			err := LinkedlistInsert(native, contract1[:], []byte{byte(i)}, []byte{byte(i * i)})
			if err != nil {
				//log.Errorf("insertion %d failed: %s\n", i, err)
				return err
			}

		}
		n, err := LinkedlistGetNumOfItems(native, contract1[:])
		if err != nil {
			return err
		}
		if n != 10 {
			err := errors.NewErr("")
			return errors.NewDetailErr(err, errors.ErrNoCode, "[linked list basic test] num != 10")
		}
	}

	{
		log.Trace("'basic delete' test")
		//basic delete
		s, err := LinkedlistDelete(native, contract1[:], []byte{byte(11)})
		if err != nil {
			return err
		}
		if s {
			err := errors.NewErr("")
			return errors.NewDetailErr(err, errors.ErrNoCode, "[linked list basic test] delete nonexistent item returns true")
		}
	}

	{
		log.Trace("'more delete' test")
		//insert & delete
		for i := 0; i < 10; i += 2 {
			suc, err := LinkedlistDelete(native, contract1[:], []byte{byte(i)})
			if err != nil {
				//log.Errorf("")
				return err
			}
			if suc != true {
				err := errors.NewErr("")
				return errors.NewDetailErr(err, errors.ErrNoCode, "[linked list insert&delete test] delete failed")
			}
		}
		n, err := LinkedlistGetNumOfItems(native, contract1[:])
		if err != nil {
			return err
		}
		if n != 5 {
			err := errors.NewErr("")
			return errors.NewDetailErr(err, errors.ErrNoCode, "[linked list basic test] num != 5")
		}
	}

	{
		log.Trace("'insert & delete' test")
		for i := 2; i < 12; i++ {
			item := make([]byte, i)
			for j := 0; j < i; j++ {
				item[j] = byte(j)
			}
			err := LinkedlistInsert(native, contract1[:], item, []byte("test"))
			if err != nil {
				return err
			}

			if i%2 == 0 {
				suc, err := LinkedlistDelete(native, contract1[:], item)
				if err != nil {
					return err
				}
				if !suc {
					return errors.NewErr("[linkedlist] delete failed")
				}
			}
		}
		n, err := LinkedlistGetNumOfItems(native, contract1[:])
		if err != nil {
			return err
		}
		if n != 10 {
			err := errors.NewErr("")
			return errors.NewDetailErr(err, errors.ErrNoCode, "[linked list basic test] num != 5")
		}
	}
	return nil
}
*/
