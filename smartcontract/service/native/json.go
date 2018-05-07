package native

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"sort"
	"strings"

	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/core/store/common"
)

func PutJson(srvc *NativeService, key, data []byte) error {
	m := make(map[string]interface{})
	err := json.Unmarshal(data, &m)
	if err != nil {
		return err
	}
	return storeJson(srvc, string(key), m)
}

func GetJson(srvc *NativeService, key []byte) ([]byte, error) {
	items, err := srvc.CloneCache.Store.Find(common.ST_STORAGE, key)
	if err != nil {
		return nil, err
	}
	sort.Sort(sitems(items))
	obj, err := parseJsonObject(items, string(key))
	if err != nil {
		return nil, err
	}

	return json.Marshal(obj)
}

func DelJson(srvc *NativeService, key []byte) error {
	key1 := append(key, byte('.'))
	items, err := srvc.CloneCache.Store.Find(common.ST_STORAGE, key1)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		return errors.New("data not found")
	}
	for _, item := range items {
		srvc.CloneCache.Delete(common.ST_STORAGE, []byte(item.Key))
	}
	srvc.CloneCache.Delete(common.ST_STORAGE, key)
	return nil
}

const (
	vt_boolean byte = 1 + iota
	vt_number
	vt_string
	vt_array
	vt_object
)

type storeUnit struct {
	valueType byte
	value     interface{}
}

func (this *storeUnit) Serialize(w io.Writer) error {
	err := serialization.WriteByte(w, this.valueType)
	if err != nil {
		return err
	}
	switch t := this.value.(type) {
	case bool:
		err = serialization.WriteBool(w, t)
	case string:
		err = serialization.WriteString(w, t)
	case float64:
		err = serialization.WriteUint64(w, math.Float64bits(t))
	case []byte:
		err = serialization.WriteVarBytes(w, t)
	default:
		panic("wrong value type")
	}

	return err
}

func (this *storeUnit) Deserialize(r io.Reader) error {
	vt, err := serialization.ReadByte(r)
	if err != nil {
		return err
	}

	var val interface{}
	switch vt {
	case vt_boolean:
		val, err = serialization.ReadBool(r)
	case vt_string:
		val, err = serialization.ReadString(r)
	case vt_number:
		v, err1 := serialization.ReadUint64(r)
		val = math.Float64frombits(v)
		err = err1
	case vt_array, vt_object:
		val, err = serialization.ReadVarBytes(r)
	}
	if err != nil {
		return err
	}
	this.valueType = vt
	this.value = val
	return nil
}

func (this *storeUnit) Bytes() ([]byte, error) {
	var buf bytes.Buffer
	err := this.Serialize(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (this *storeUnit) SetBytes(data []byte) error {
	buf := bytes.NewBuffer(data)
	return this.Deserialize(buf)
}

func makePath(prefix, key string) string {
	//TODO confirm
	return prefix + "." + key
}

func makeValue(srvc *NativeService, key string, val interface{}) (*storeUnit, error) {
	var u = storeUnit{value: val}
	switch t := val.(type) {
	case bool:
		u.valueType = vt_boolean
	case string:
		u.valueType = vt_string
	case float64:
		u.valueType = vt_number
	case []interface{}:
		head, err := storeArray(srvc, key, t)
		if err != nil {
			return nil, err
		}
		u.valueType = vt_array
		u.value = head
	case map[string]interface{}:
		err := storeJson(srvc, key, t)
		if err != nil {
			return nil, err
		}
		u.valueType = vt_object
		u.value = nil
	}
	return &u, nil
}

func storeJson(srvc *NativeService, path string, data map[string]interface{}) error {
	for k, v := range data {
		p := makePath(path, k)
		u, err := makeValue(srvc, p, v)
		if err != nil {
			return err
		}
		val, err := u.Bytes()
		if err != nil {
			return err
		}
		srvc.CloneCache.Add(common.ST_STORAGE, []byte(p), &states.StorageItem{Value: val})
	}

	return nil
}

func storeArray(srvc *NativeService, path string, data []interface{}) ([]byte, error) {
	path1 := path + ".val" //TODO confirm
	for i, v := range data {
		path2 := fmt.Sprintf("%s%d", path1, i)
		u, err := makeValue(srvc, path2, v)
		if err != nil {
			return nil, err
		}
		var buf bytes.Buffer
		err = u.Serialize(&buf)
		if err != nil {
			return nil, err
		}
		err = LinkedlistInsert(srvc, []byte(path1), []byte(path2), buf.Bytes())
		if err != nil {
			return nil, err
		}
	}
	return []byte(path1), nil
}

type sitems []*common.StateItem

func (this sitems) Len() int {
	return len(this)
}

func (this sitems) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}

func (this sitems) Less(i, j int) bool {
	return this[i].Key < this[j].Key
}

// return the end index of sub items
// search from the start index
func findSubItems(data []*common.StateItem, prefix string, start int) int {
	var i = start
	for i < len(data) {
		v := data[i]
		if !strings.HasPrefix(v.Key, prefix) {
			break
		}
		i += 1
	}
	return i
}

func parseJsonObject(data []*common.StateItem, path string) (map[string]interface{}, error) {
	prefix := path + "."
	res := make(map[string]interface{})
	for i := 0; i < len(data); {
		//TODO confirm parse key
		if !strings.HasPrefix(data[i].Key, prefix) {
			return nil, errors.New("error prefix")
		}
		path1 := strings.TrimPrefix(data[i].Key, prefix)
		s := strings.Split(path1, ".")
		if len(s) != 1 {
			return nil, errors.New("error key")
		}

		val, next, err := parseValue(data, i, prefix+s[0])
		if err != nil {
			return nil, err
		}

		res[s[0]] = val
		i = next
	}

	return res, nil
}

func parseJsonArray(data []*common.StateItem, path string) ([]interface{}, error) {
	var res []interface{}
	for i := 0; i < len(data); {
		val, next, err := parseValue(data, i, path)
		if err != nil {
			return nil, err
		}
		res = append(res, val)
		i = next
	}
	return res, nil
}

func parseValue(data []*common.StateItem, index int, prefix string) (val interface{}, next int, err error) {
	v := data[index]
	index += 1

	t, ok := v.Value.(*states.StorageItem)
	if !ok {
		panic("error storage item")
	}
	buf := bytes.NewBuffer(t.Value)
	u := new(storeUnit)
	err = u.Deserialize(buf)
	if err != nil {
		return
	}
	switch u.valueType {
	case vt_boolean, vt_string, vt_number:
		val = u.value
		next = index
	case vt_array:
		end := findSubItems(data, prefix, index)
		if end <= index {
			//empty array
			val = make([]interface{}, 0)
		} else {
			val, err = parseJsonArray(data[index:end], prefix)
			if err != nil {
				return
			}
			next = end
		}
	case vt_object:
		end := findSubItems(data, prefix, index)
		if end <= index {
			//empty object
			val = make(map[string]interface{})
		} else {
			val, err = parseJsonObject(data[index:end], prefix)
			if err != nil {
				return
			}
			next = end
		}
	}

	return
}
