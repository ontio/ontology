package utils

import (
	"bytes"

	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/common/serialization"
	cstates "github.com/ontio/ontology/core/states"
	scommon "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/service/native"
)

type LinkedlistNode struct {
	next    []byte
	prev    []byte
	payload []byte
}

func (this *LinkedlistNode) GetPrevious() []byte {
	return this.prev
}

func (this *LinkedlistNode) GetNext() []byte {
	return this.next
}

func (this *LinkedlistNode) GetPayload() []byte {
	return this.payload
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
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[linked list] serialize next error!")
	}
	if err := serialization.WriteVarBytes(bf, this.prev); err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[linked list] serialize prev error!")
	}
	if err := serialization.WriteVarBytes(bf, this.payload); err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[linked list] serialize payload error!")
	}
	return bf.Bytes(), nil
}

func (this *LinkedlistNode) Deserialize(r []byte) error {
	bf := bytes.NewReader(r)
	next, err := serialization.ReadVarBytes(bf)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[linked list] deserialize next error!")
	}
	prev, err := serialization.ReadVarBytes(bf)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[linked list] deserialize prev error!")
	}
	payload, err := serialization.ReadVarBytes(bf)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[linked list] deserialize payload error!")
	}
	this.next = next
	this.prev = prev
	this.payload = payload
	return nil
}

func getListHead(native *native.NativeService, index []byte) ([]byte, error) {
	head, err := native.CloneCache.Get(scommon.ST_STORAGE, index)
	if err != nil {
		return nil, err
	}
	if head == nil {
		return nil, nil
	}
	item, ok := head.(*cstates.StorageItem)
	if !ok {
		err := errors.NewErr("")
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[linked list] get header error")
	}
	return item.Value, nil
}

func getListNode(native *native.NativeService, index []byte, item []byte) (*LinkedlistNode, error) {
	node := new(LinkedlistNode)
	data, err := native.CloneCache.Get(scommon.ST_STORAGE, append(index, item...))
	if err != nil {
		//log.Trace(err)
		return nil, err
	}
	if data == nil {
		return nil, nil
	}
	raw_node, ok := data.(*cstates.StorageItem)
	if !ok {
		err := errors.NewErr("")
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[linked list] get list node error")
	}
	if raw_node.Value == nil || len(raw_node.Value) == 0 {
		return nil, nil
	}
	err = node.Deserialize(raw_node.Value)
	if err != nil {
		//log.Tracef("[index: %s, item: %s] error %s", hex.EncodeToString(index), hex.EncodeToString(item), err)
		return nil, err
	}
	return node, nil
}

func LinkedlistInsert(native *native.NativeService, index []byte, item []byte, payload []byte) error {
	null := []byte{}
	if item == nil {
		return errors.NewErr("[linked list] invalid item")
	}
	head, err := getListHead(native, index) //list head
	if err != nil {
		//log.Trace(err)
		return err
	}

	q, err := getListNode(native, index, item) //list node
	if err != nil {
		//log.Trace(err)
		return err
	}

	if q != nil { //already exists
		//log.Trace(err)
		node, err := makeLinkedlistNode(q.next, q.prev, payload)
		if err != nil {
			return err
		}
		PutBytes(native, append(index, item...), node) //update it
		return nil
	}
	if head == nil { //doubly-linked list contains zero element
		node, err := makeLinkedlistNode(null, null, payload)
		if err != nil {
			//log.Trace(err)
			return err
		}
		PutBytes(native, append(index, item...), node) //item is the only element
		PutBytes(native, index, item)                  //item becomes head
	} else {
		null := []byte{}
		node, err := makeLinkedlistNode(head, null, payload)
		if err != nil {
			//log.Trace(err)
			return err
		}
		PutBytes(native, append(index, item...), node) //item.next = head, item.prev = null,
		// item.payload = payload
		qhead, err := getListNode(native, index, head)
		if err != nil {
			//log.Trace(err)
			return err
		}

		node, err = makeLinkedlistNode(qhead.next, item, qhead.payload)
		if err != nil {
			//log.Trace(err)
			return err
		}
		PutBytes(native, append(index, head...), node) //head.next = head.next, head.prev = item,
		// head.payload = head.payload
		PutBytes(native, index, item) // item becomes head
	}
	return nil
}

func LinkedlistDelete(native *native.NativeService, index []byte, item []byte) (bool, error) {
	null := []byte{}
	if item == nil {
		return false, errors.NewErr("[linked list] invalid item")
	}
	q, err := getListNode(native, index, item)
	if err != nil {
		return false, err
	}
	if q == nil {
		return false, nil
	}

	prev, next := q.prev, q.next
	if prev == nil {
		if next == nil {
			PutBytes(native, index, null) //clear linked list
		} else {
			qnext, err := getListNode(native, index, next)
			if err != nil {
				return false, err
			}
			node, err := makeLinkedlistNode(qnext.next, null, qnext.payload) //qnext.next = qnext.next
			if err != nil {                                                  // qnext.prev = nil
				return false, err
			}
			PutBytes(native, append(index, next...), node)
			PutBytes(native, index, next) //next becomes head
		}
	} else {
		if next == nil {
			qprev, err := getListNode(native, index, prev)
			if err != nil {
				return false, err
			}
			node, err := makeLinkedlistNode(null, qprev.prev, qprev.payload) //qprev becomes end
			if err != nil {
				return false, err
			}
			PutBytes(native, append(index, prev...), node)
		} else {
			qprev, err := getListNode(native, index, prev)
			if err != nil {
				return false, err
			}
			qnext, err := getListNode(native, index, next)
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
			PutBytes(native, append(index, prev...), node_prev)
			PutBytes(native, append(index, next...), node_next)
		}
	}
	native.CloneCache.Delete(scommon.ST_STORAGE, append(index, item...))
	return true, nil
}

func LinkedlistGetItem(native *native.NativeService, index []byte, item []byte) (*LinkedlistNode, error) {
	if item == nil {
		return nil, errors.NewErr("[linkedlist getNext] item is nil")
	}
	q, err := getListNode(native, index, item)
	if err != nil {
		return nil, err
	}
	return q, nil
}

func LinkedlistGetHead(native *native.NativeService, index []byte) ([]byte, error) {
	head, err := getListHead(native, index)
	if err != nil {
		return nil, err
	}
	return head, nil
}
func LinkedlistGetNumOfItems(native *native.NativeService, index []byte) (int, error) {
	n := 0
	head, err := getListHead(native, index)
	if err != nil {
		return 0, err
	}
	q := head
	for q != nil {
		n += 1
		qnode, err := getListNode(native, index, q)
		if err != nil {
			return 0, err
		}
		q = qnode.next
	}
	return n, nil
}

func linkedlistTest(native *native.NativeService) error {
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
