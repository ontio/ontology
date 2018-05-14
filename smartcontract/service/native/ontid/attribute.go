package ontid

import (
	"bytes"
	"errors"
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
