package types

type MapValue struct {
	data map[string]VmValue
}

func NewMapValue() *MapValue {
	return &MapValue{data: make(map[string]VmValue)}
}

func (this *MapValue) Set(key, value VmValue) error {
	skey, err := key.GetMapKey()
	if err != nil {
		return err
	}

	this.data[skey] = value
	return nil
}

func (this *MapValue) Reset() {
	this.data = make(map[string]VmValue)
}

func (this *MapValue) Remove(key VmValue) {
	skey, err := key.GetMapKey()
	if err != nil {
		return
	}

	delete(this.data, skey)
}

func (this *MapValue) Get(key VmValue) (value VmValue, ok bool) {
	skey, e := key.GetMapKey()
	if e != nil {
		ok = false
		return
	}

	value, ok = this.data[skey]
	return
}
