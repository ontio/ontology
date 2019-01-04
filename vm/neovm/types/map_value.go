package types

import (
	"sort"
)

type MapValue struct {
	Data map[string]VmValue
}

func NewMapValue() *MapValue {
	return &MapValue{Data: make(map[string]VmValue)}
}
func (this *MapValue) Set(key, value VmValue) error {
	skey, err := key.GetMapKey()
	if err != nil {
		return err
	}

	this.Data[skey] = value
	return nil
}

func (this *MapValue) Reset() {
	this.Data = make(map[string]VmValue)
}

func (this *MapValue) Remove(key VmValue) error {
	skey, err := key.GetMapKey()
	if err != nil {
		return err
	}

	delete(this.Data, skey)

	return nil
}

func (this *MapValue) Get(key VmValue) (value VmValue, ok bool, err error) {
	skey, e := key.GetMapKey()
	if e != nil {
		err = e
		return
	}

	value, ok = this.Data[skey]
	return
}

func (this *MapValue) GetMapSortedKey() ([]string, error) {
	var unsortKey []string
	for k := range this.Data {
		unsortKey = append(unsortKey, k)
	}
	sort.Strings(unsortKey)
	return unsortKey, nil
}

func (this *MapValue) GetValues() ([]VmValue, error) {
	keys, err := this.GetMapSortedKey()
	if err != nil {
		return nil, err
	}
	values := make([]VmValue, len(this.Data))
	for j, v := range keys {
		values[j] = this.Data[v]
	}
	return values, nil
}
